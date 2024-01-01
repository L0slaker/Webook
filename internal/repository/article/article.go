package article

import (
	"Prove/webook/internal/domain"
	"Prove/webook/internal/events/article"
	"Prove/webook/internal/repository"
	"Prove/webook/internal/repository/cache"
	dao "Prove/webook/internal/repository/dao/article"
	"Prove/webook/pkg/logger"
	"context"
	"github.com/ecodeclub/ekit/slice"
	"gorm.io/gorm"
	"time"
)

type ArticleRepository interface {
	Create(ctx context.Context, article domain.Article) (int64, error)
	Update(ctx context.Context, article domain.Article) error
	Sync(ctx context.Context, article domain.Article) (int64, error)
	SyncStatus(ctx context.Context, article domain.Article) error
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id int64) (domain.Article, error)
	ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error)
}

type articleRepository struct {
	dao      dao.ArticleDAO
	userRepo repository.UserRepository
	artCache cache.ArticleCache
	logger   logger.LoggerV1

	// V1 操作两个DAO
	readerDAO dao.ReaderDAO
	authorDAO dao.AuthorDAO

	// V2 尝试在repository层面上解决事务问题，
	// 确保保存到制作库和线上库同时成功或同时失败
	db       *gorm.DB
	producer article.KafkaProducer
}

func NewArticleRepository(dao dao.ArticleDAO, c cache.ArticleCache, logger logger.LoggerV1) ArticleRepository {
	return &articleRepository{
		dao:    dao,
		logger: logger,
	}
}

func (a *articleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	defer func() {
		// 清空缓存
		err := a.artCache.DelFirstPage(ctx, art.Author.Id)
		a.logger.Error("删除缓存失败！", logger.Error(err))
	}()
	return a.dao.Insert(ctx, dao.Article{
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
	})
}

func (a *articleRepository) Update(ctx context.Context, art domain.Article) error {
	// 从缓存或数据库中读取数据
	return a.dao.UpdateById(ctx, a.toEntity(art))
}

func (a *articleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := a.dao.Sync(ctx, a.toEntity(art))
	if err == nil {
		// 清空缓存
		err = a.artCache.DelFirstPage(ctx, art.Author.Id)
		if err != nil {
			a.logger.Error("删除缓存失败", logger.Error(err))
		}
		// 提前缓存好线上库数据
		err = a.artCache.SetPub(ctx, art)
		if err != nil {
			a.logger.Error("提前设置缓存失败", logger.Error(err))
		}
	}
	return id, err
}

func (a *articleRepository) SyncStatus(ctx context.Context, art domain.Article) error {
	return a.dao.SyncStatus(ctx, a.toEntity(art))
}

func (a *articleRepository) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	// 集成缓存方案，缓存第一页
	if offset == 0 && limit <= 0 {
		data, err := a.artCache.GetFirstPage(ctx, uid)
		if err == nil {
			// 预加载一部分内容，可以考虑异步处理
			go func() {
				a.preCache(ctx, data)
			}()
			return data, nil
		}
	}
	res, err := a.dao.GetByAuthorId(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	data := slice.Map[dao.Article, domain.Article](res, func(idx int, src dao.Article) domain.Article {
		return a.toDomain(src)
	})
	// 回写缓存时要考虑缓存策略 => 是否会有并发问题
	// 如果有高并发，就考虑删除缓存；如果没有高并发就可以考虑 set
	// 可以考虑异步回写缓存
	go func() {
		e := a.artCache.SetFirstPage(ctx, uid, data)
		a.logger.Error("回写缓存失败！", logger.Error(e))
		// 预加载一部分内容
		a.preCache(ctx, data)
	}()
	return data, err
}

func (a *articleRepository) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]domain.Article, error) {
	res, err := a.dao.ListPub(ctx, start, offset, limit)
	if err != nil {
		return nil, err
	}
	return slice.Map(res, func(idx int, src dao.Article) domain.Article {
		return a.toDomain(src)
	}), nil
}

func (a *articleRepository) GetById(ctx context.Context, id int64) (domain.Article, error) {
	res, err := a.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	return a.toDomain(res), nil
}

func (a *articleRepository) GetPublishedById(ctx context.Context, id int64) (domain.Article, error) {
	// 读取线上库数据，如果 Content 被放到 OSS 上，就需要让前端去读 Content 字段
	art, err := a.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	// 组装 user
	u, err := a.userRepo.FindById(ctx, art.Id)
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Status:  domain.ArticleStatus(art.Status),
		Author: domain.Author{
			Id:   u.Id,
			Name: u.Nickname,
		},
		CreateTime: time.UnixMilli(art.CreateTime),
		UpdateTime: time.UnixMilli(art.UpdateTime),
	}, nil
}

func (a *articleRepository) SyncV2(ctx context.Context, art domain.Article) (int64, error) {
	// 开启事务
	tx := a.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	// 防止操作时panic，导致事务既没提交也没回滚
	defer tx.Rollback()
	author := dao.NewAuthorDAO(tx)
	reader := dao.NewReaderDAO(tx)

	var (
		id  = art.Id
		err error
	)

	if id > 0 {
		err = author.UpdateById(ctx, a.toEntity(art))
	} else {
		id, err = author.Insert(ctx, a.toEntity(art))
	}

	if err != nil {
		// 执行失败，回滚
		tx.Rollback()
		return id, err
	}
	// 操作库上线了，保存数据并同步过来
	// 此时线上库无法确定是否有文章，所以要使用 UPSERT 语句
	err = reader.Upsert(ctx, a.toEntity(art))
	// UpsertV1 模拟完全不同的表
	//err = reader.UpsertV1(ctx, dao.PublishedArticle{Article: a.toEntity(art)})
	// 执行成功，提交事务
	tx.Commit()
	return id, err
}

func (a *articleRepository) SyncV1(ctx context.Context, article domain.Article) (int64, error) {
	var (
		id  = article.Id
		err error
	)
	if id > 0 {
		err = a.authorDAO.UpdateById(ctx, a.toEntity(article))
	} else {
		id, err = a.authorDAO.Insert(ctx, a.toEntity(article))
	}

	if err != nil {
		return id, err
	}

	// 操作库上线了，保存数据并同步过来
	// 此时线上库无法确定是否有文章，所以要使用 UPSERT 语句
	err = a.readerDAO.Upsert(ctx, a.toEntity(article))
	return id, err
}

func (a *articleRepository) toEntity(art domain.Article) dao.Article {
	return dao.Article{
		Id:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
		//CreateTime: art.CreateTime.UnixMilli(),
		//UpdateTime: art.UpdateTime.UnixMilli(),
	}
}

func (a *articleRepository) toDomain(art dao.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Status:  domain.ArticleStatus(art.Status),
		Author: domain.Author{
			Id: art.AuthorId,
		},
		CreateTime: time.UnixMilli(art.CreateTime),
		UpdateTime: time.UnixMilli(art.UpdateTime),
	}
}

func (a *articleRepository) preCache(ctx context.Context, data []domain.Article) {
	// 不缓存大对象（1MB）
	if len(data) > 0 && len(data[0].Content) < 1024*1024 {
		err := a.artCache.Set(ctx, data[0])
		if err != nil {
			a.logger.Error("提前预加载缓存失败！", logger.Error(err))
		}
	}
}
