package repository

import (
	"Prove/webook/internal/domain"
	"Prove/webook/internal/repository/cache"
	"Prove/webook/internal/repository/dao"
	"Prove/webook/pkg/logger"
	"context"
)

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, id int64) error
	IncrLike(ctx context.Context, biz string, bizId, uid int64) error
	DecrLike(ctx context.Context, biz string, bizId, uid int64) error
	AddCollectionItem(ctx context.Context, biz string, bizId, uid, cid int64) error
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Liked(ctx context.Context, biz string, bizId, uid int64) (bool, error)
	Collected(ctx context.Context, biz string, bizId, uid int64) (bool, error)
}

type CachedCntRepository struct {
	cache cache.InteractiveCache
	dao   dao.InteractiveDAO
	l     logger.LoggerV1
}

func NewCachedInteractiveRepository(cache cache.InteractiveCache, dao dao.InteractiveDAO, l logger.LoggerV1) InteractiveRepository {
	return &CachedCntRepository{
		cache: cache,
		dao:   dao,
		l:     l,
	}
}

func (c *CachedCntRepository) IncrReadCnt(ctx context.Context, biz string, id int64) error {
	// 优先保证数据库里数据的准确性
	err := c.dao.IncrReadCnt(ctx, biz, id)
	if err != nil {
		return err
	}
	// 也可以考虑异步更新缓存的字段
	//go func() {
	//	c.cache.IncrReadCntIfPresent(ctx, biz, id)
	//}()
	return c.cache.IncrReadCntIfPresent(ctx, biz, id)
}

func (c *CachedCntRepository) IncrLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	// 先插入点赞；然后更新点赞计数；最后更新缓存
	// 前两步交给 dao 统一处理
	err := c.dao.InsertLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}
	return c.cache.IncrLikeCntIfPresent(ctx, biz, bizId)
}

func (c *CachedCntRepository) DecrLike(ctx context.Context, biz string, bizId int64, uid int64) error {
	err := c.dao.DeleteLikeInfo(ctx, biz, bizId, uid)
	if err != nil {
		return err
	}
	return c.cache.DecrLikeCntIfPresent(ctx, biz, bizId)
}

func (c *CachedCntRepository) AddCollectionItem(ctx context.Context, biz string, bizId, uid, cid int64) error {
	err := c.dao.InsertCollectionBiz(ctx, dao.UserCollectionBiz{
		Uid:   uid,
		Cid:   cid,
		BizId: bizId,
		Biz:   biz,
	})
	if err != nil {
		return err
	}
	return c.cache.IncrCollectionCntIfPresent(ctx, biz, bizId)
}

func (c *CachedCntRepository) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	inter, err := c.cache.Get(ctx, biz, bizId)
	if err == nil {
		return inter, nil
	}
	daoInter, err := c.dao.Get(ctx, biz, bizId)
	if err != nil {
		return domain.Interactive{}, err
	}
	inter = c.toDomain(daoInter)
	// 异步回写到缓存中
	go func() {
		setErr := c.cache.Set(ctx, biz, bizId, inter)
		if setErr != nil {
			c.l.Error("回写缓存失败！",
				logger.String("biz", biz),
				logger.Int64("biz_id", bizId),
			)
		}
	}()
	return inter, nil
}

func (c *CachedCntRepository) Liked(ctx context.Context, biz string, bizId, uid int64) (bool, error) {
	_, err := c.dao.GetLikeInfo(ctx, biz, bizId, uid)
	switch err {
	case nil:
		return true, nil
	case dao.ErrRecordNotFound:
		return false, err
	default:
		return false, err
	}
}

func (c *CachedCntRepository) Collected(ctx context.Context, biz string, bizId, uid int64) (bool, error) {
	_, err := c.dao.GetCollectionInfo(ctx, biz, bizId, uid)
	switch err {
	case nil:
		return true, nil
	case dao.ErrRecordNotFound:
		return false, err
	default:
		return false, err
	}
}

func (c *CachedCntRepository) toDomain(inter dao.Interactive) domain.Interactive {
	return domain.Interactive{
		ReadCnt:    inter.ReadCnt,
		LikeCnt:    inter.LikeCnt,
		CollectCnt: inter.CollectCnt,
	}
}

func (c *CachedCntRepository) toEntity(inter domain.Interactive) dao.Interactive {
	return dao.Interactive{
		ReadCnt:    inter.ReadCnt,
		LikeCnt:    inter.LikeCnt,
		CollectCnt: inter.CollectCnt,
	}
}