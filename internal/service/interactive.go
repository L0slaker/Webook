package service

import (
	"Prove/webook/internal/domain"
	"Prove/webook/internal/repository"
	"Prove/webook/pkg/logger"
	"context"
	"golang.org/x/sync/errgroup"
)

type InteractiveService interface {
	// IncrReadCnt 阅读量计数
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	// Like 点赞
	Like(ctx context.Context, biz string, bizId, uid int64) error
	// CancelLike 取消点赞
	CancelLike(ctx context.Context, biz string, bizId, uid int64) error
	// Collect 收藏
	Collect(ctx context.Context, biz string, bizId, uid, cid int64) error
	Get(ctx context.Context, biz string, bizId, uid int64) (domain.Interactive, error)
	GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error)
}

type interactiveService struct {
	repo repository.InteractiveRepository
	l    logger.LoggerV1
}

func NewInteractiveService(repo repository.InteractiveRepository, l logger.LoggerV1) InteractiveService {
	return &interactiveService{
		repo: repo,
		l:    l,
	}
}

func (i *interactiveService) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return i.repo.IncrReadCnt(ctx, biz, bizId)
}

func (i *interactiveService) Like(ctx context.Context, biz string, bizId, uid int64) error {
	return i.repo.IncrLike(ctx, biz, bizId, uid)
}

func (i *interactiveService) CancelLike(ctx context.Context, biz string, bizId, uid int64) error {
	return i.repo.DecrLike(ctx, biz, bizId, uid)
}

func (i *interactiveService) Collect(ctx context.Context, biz string, bizId, uid, cid int64) error {
	return i.repo.AddCollectionItem(ctx, biz, bizId, uid, cid)
}

func (i *interactiveService) Get(ctx context.Context, biz string, bizId, uid int64) (domain.Interactive, error) {
	inter, err := i.repo.Get(ctx, biz, bizId)
	if err != nil {
		return domain.Interactive{}, err
	}
	var eg errgroup.Group
	eg.Go(func() error {
		inter.Liked, err = i.repo.Liked(ctx, biz, bizId, uid)
		return err
	})
	eg.Go(func() error {
		inter.Collected, err = i.repo.Collected(ctx, biz, bizId, uid)
		return err
	})
	// 说明是登录过的，补充用户是否点赞或者新的打印日志的形态 zap 本身就有这种用法
	err = eg.Wait()
	if err != nil {
		// 这个查询失败只需要记录日志就可以，不需要中断执行
		i.l.Error("查询用户是否点赞的信息失败",
			logger.String("biz", biz),
			logger.Int64("bizId", bizId),
			logger.Int64("uid", uid),
			logger.Error(err))
	}
	return inter, nil
}

func (i *interactiveService) GetByIds(ctx context.Context, biz string, bizIds []int64) (map[int64]domain.Interactive, error) {
	inters, err := i.repo.GetByIds(ctx, biz, bizIds)
	if err != nil {
		return nil, err
	}
	res := make(map[int64]domain.Interactive, len(inters))
	for _, inter := range inters {
		res[inter.BizId] = inter
	}
	return res, nil
}
