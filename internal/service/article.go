package service

import (
	"Prove/webook/internal/domain"
	events "Prove/webook/internal/events/article"
	"Prove/webook/internal/repository/article"
	"Prove/webook/pkg/logger"
	"context"
	"time"
)

type ArticleService interface {
	Save(ctx context.Context, article domain.Article) (int64, error)
	Withdraw(ctx context.Context, article domain.Article) error
	Publish(ctx context.Context, article domain.Article) (int64, error)
	PublishV1(ctx context.Context, article domain.Article) (int64, error)
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id, uid int64) (domain.Article, error)
}

type articleService struct {
	repo     article.ArticleRepository
	author   article.ArticleAuthorRepository
	reader   article.ArticleReaderRepository
	l        logger.LoggerV1
	producer events.Producer
	ch       chan readInfo
}

type readInfo struct {
	uid int64
	aid int64
}

func NewArticleService(repo article.ArticleRepository, l logger.LoggerV1, producer events.Producer) ArticleService {
	return &articleService{
		repo:     repo,
		l:        l,
		producer: producer,
		//ch: make(chan readInfo, 10),
	}
}

func NewArticleServiceV2(repo article.ArticleRepository, l logger.LoggerV1, producer events.Producer) ArticleService {
	ch := make(chan readInfo, 10)
	go func() {
		for {
			uids := make([]int64, 0, 10)
			aids := make([]int64, 0, 10)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			for i := 0; i < 10; i++ {
				select {
				case <-ctx.Done():
					break
				case info, ok := <-ch:
					if !ok {
						cancel()
						return
					}
					uids = append(uids, info.uid)
					aids = append(aids, info.aid)
				}
			}
			cancel()
			ctx, cancel = context.WithTimeout(context.Background(), time.Second)
			producer.ProduceReadEventV1(ctx, events.ReadEventV1{
				Uids: uids,
				Aids: aids,
			})
			cancel()
		}
	}()
	return &articleService{
		repo:     repo,
		l:        l,
		producer: producer,
		// 假定这边批量处理十条消息
		ch: ch,
	}
}

func NewArticleServiceV1(author article.ArticleAuthorRepository, reader article.ArticleReaderRepository, l logger.LoggerV1) ArticleService {
	return &articleService{
		author: author,
		reader: reader,
		l:      l,
	}
}

// Save 区别创建还是更新取决于是否有id
func (a *articleService) Save(ctx context.Context, article domain.Article) (int64, error) {
	article.Status = domain.ArticleStatusUnPublished
	if article.Id > 0 {
		return article.Id, a.repo.Update(ctx, article)
	}
	return a.repo.Create(ctx, article)
}

func (a *articleService) Withdraw(ctx context.Context, article domain.Article) error {
	article.Status = domain.ArticleStatusUnPublished
	// 制作库同步到线上库
	return a.repo.SyncStatus(ctx, article)
}

func (a *articleService) Publish(ctx context.Context, article domain.Article) (int64, error) {
	article.Status = domain.ArticleStatusPublished
	// 制作库同步到线上库
	return a.repo.Sync(ctx, article)
}

func (a *articleService) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return a.repo.List(ctx, uid, offset, limit)
}

func (a *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return a.repo.GetById(ctx, id)
}

func (a *articleService) GetPublishedById(ctx context.Context, id, uid int64) (domain.Article, error) {
	art, err := a.repo.GetPublishedById(ctx, id)
	if err == nil {
		go func() {
			e := a.producer.ProduceReadEvent(ctx, events.ReadEvent{
				Uid: uid,
				Aid: id,
			})
			if e != nil {
				a.l.Error("发送读者阅读事件失败！")
			}
		}()

		//go func() {
		//	// 改批量的做法
		//	a.ch <- readInfo{
		//		uid: uid,
		//		aid: id,
		//	}
		//}()
	}
	return art, err
}

func (a *articleService) PublishV1(ctx context.Context, article domain.Article) (int64, error) {
	var (
		err error
		id  = article.Id
	)
	// 制作库
	if article.Id > 0 {
		err = a.author.Update(ctx, article)
	} else {
		id, err = a.author.Create(ctx, article)
	}
	if err != nil {
		return 0, err
	}
	// 制作库与线上库的 ID 匹配
	article.Id = id
	//if err != nil {
	//	// 制作库更新成功，但线上库更新失败，是否要考虑开启事务，在哪个层面上开？
	//	// 记录日志
	//}
	// 重试
	for i := 0; i < 3; i++ {
		id, err = a.reader.Save(ctx, article)
		if err == nil {
			break
		}
		a.l.Error("保存到线上库失败！",
			logger.Int64("article_id", article.Id),
			logger.Error(err))
	}
	if err != nil {
		a.l.Error("保存到线上库失败！重试彻底失败！",
			logger.Int64("article_id", article.Id),
			logger.Error(err))
		// 接入告警系统，手工处理
	}
	return id, err
}
