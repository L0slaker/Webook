package service

import (
	"Prove/webook/internal/domain"
	"Prove/webook/internal/repository/article"
	"Prove/webook/pkg/logger"
	"context"
)

type ArticleService interface {
	Save(ctx context.Context, article domain.Article) (int64, error)
	Withdraw(ctx context.Context, article domain.Article) error
	Publish(ctx context.Context, article domain.Article) (int64, error)
	PublishV1(ctx context.Context, article domain.Article) (int64, error)
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx context.Context, id int64) (domain.Article, error)
}

type articleService struct {
	repo   article.ArticleRepository
	author article.ArticleAuthorRepository
	reader article.ArticleReaderRepository
	l      logger.LoggerV1
}

func NewArticleService(repo article.ArticleRepository) ArticleService {
	return &articleService{
		repo: repo,
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

func (a *articleService) GetPublishedById(ctx context.Context, id int64) (domain.Article, error) {
	return a.repo.GetPublishedById(ctx, id)
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
