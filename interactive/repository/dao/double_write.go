package dao

import (
	"Prove/webook/pkg/logger"
	"context"
	"errors"
	"github.com/ecodeclub/ekit/syncx/atomicx"
)

const (
	patternSrcOnly  = "SRC_ONLY"
	patternSrcFirst = "SRC_FIRST"
	patternDstFirst = "DST_FIRST"
	patternDstOnly  = "DST_ONLY"
)

var ErrUnknownPattern = errors.New("未知的写入模式")

type DoubleWriteDAO struct {
	src     InteractiveDAO
	dst     InteractiveDAO
	pattern *atomicx.Value[string]
	l       logger.LoggerV1
}

func NewDoubleWriteDAO(src InteractiveDAO, dst InteractiveDAO,
	l logger.LoggerV1) InteractiveDAO {
	return &DoubleWriteDAO{
		src:     src,
		dst:     dst,
		pattern: atomicx.NewValueOf(patternSrcOnly), // 默认只写 src
		l:       l,
	}
}

func (d *DoubleWriteDAO) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	switch d.pattern.Load() {
	case patternSrcOnly:
		return d.src.IncrReadCnt(ctx, biz, bizId)
	case patternSrcFirst:
		err := d.src.IncrReadCnt(ctx, biz, bizId)
		if err != nil {
			return err
		}
		err = d.dst.IncrReadCnt(ctx, biz, bizId)
		if err != nil {
			// 可以直接处理，也可以等后续修复
			d.l.Error("写入 dst 失败！", logger.Error(err))
		}
		return nil
	case patternDstFirst:
		err := d.dst.IncrReadCnt(ctx, biz, bizId)
		if err != nil {
			return err
		}
		err = d.src.IncrReadCnt(ctx, biz, bizId)
		if err != nil {
			// 可以直接处理，也可以等后续修复
			d.l.Error("写入 src 失败！", logger.Error(err))
		}
		return nil
	case patternDstOnly:
		return d.dst.IncrReadCnt(ctx, biz, bizId)
	default:
		return ErrUnknownPattern
	}
}

func (d *DoubleWriteDAO) BatchIncrReadCnt(ctx context.Context, bizs []string, ids []int64) error {
	switch d.pattern.Load() {
	case patternSrcOnly:
		return d.src.BatchIncrReadCnt(ctx, bizs, ids)
	case patternSrcFirst:
		err := d.src.BatchIncrReadCnt(ctx, bizs, ids)
		if err != nil {
			return err
		}
		err = d.dst.BatchIncrReadCnt(ctx, bizs, ids)
		if err != nil {
			d.l.Error("写入 dst 失败！", logger.Error(err))
		}
		return nil
	case patternDstFirst:
		err := d.dst.BatchIncrReadCnt(ctx, bizs, ids)
		if err != nil {
			return err
		}
		err = d.src.BatchIncrReadCnt(ctx, bizs, ids)
		if err != nil {
			d.l.Error("写入 src 失败！", logger.Error(err))
		}
		return nil
	case patternDstOnly:
		return d.dst.BatchIncrReadCnt(ctx, bizs, ids)
	default:
		return ErrUnknownPattern
	}
}

func (d *DoubleWriteDAO) InsertLikeInfo(ctx context.Context, biz string, bizId, uid int64) error {
	switch d.pattern.Load() {
	case patternSrcOnly:
		return d.src.InsertLikeInfo(ctx, biz, bizId, uid)
	case patternSrcFirst:
		err := d.src.InsertLikeInfo(ctx, biz, bizId, uid)
		if err != nil {
			return err
		}
		err = d.dst.InsertLikeInfo(ctx, biz, bizId, uid)
		if err != nil {
			d.l.Error("写入 dst 失败！", logger.Error(err))
		}
		return nil
	case patternDstFirst:
		err := d.dst.InsertLikeInfo(ctx, biz, bizId, uid)
		if err != nil {
			return err
		}
		err = d.src.InsertLikeInfo(ctx, biz, bizId, uid)
		if err != nil {
			d.l.Error("写入 src 失败！", logger.Error(err))
		}
		return nil
	case patternDstOnly:
		return d.dst.InsertLikeInfo(ctx, biz, bizId, uid)
	default:
		return ErrUnknownPattern
	}
}

func (d *DoubleWriteDAO) GetLikeInfo(ctx context.Context, biz string, bizId, uid int64) (UserLikeBiz, error) {
	switch d.pattern.Load() {
	case patternSrcOnly, patternSrcFirst:
		return d.src.GetLikeInfo(ctx, biz, bizId, uid)
	case patternDstFirst, patternDstOnly:
		return d.dst.GetLikeInfo(ctx, biz, bizId, uid)
	default:
		return UserLikeBiz{}, ErrUnknownPattern
	}
}

func (d *DoubleWriteDAO) DeleteLikeInfo(ctx context.Context, biz string, bizId, uid int64) error {
	switch d.pattern.Load() {
	case patternSrcOnly:
		return d.src.DeleteLikeInfo(ctx, biz, bizId, uid)
	case patternSrcFirst:
		err := d.src.DeleteLikeInfo(ctx, biz, bizId, uid)
		if err != nil {
			return err
		}
		err = d.dst.DeleteLikeInfo(ctx, biz, bizId, uid)
		if err != nil {
			d.l.Error("从 dst 删除失败！", logger.Error(err))
		}
		return nil
	case patternDstFirst:
		err := d.dst.DeleteLikeInfo(ctx, biz, bizId, uid)
		if err != nil {
			return err
		}
		err = d.src.DeleteLikeInfo(ctx, biz, bizId, uid)
		if err != nil {
			d.l.Error("从 src 删除失败！", logger.Error(err))
		}
		return nil
	case patternDstOnly:
		return d.dst.DeleteLikeInfo(ctx, biz, bizId, uid)
	default:
		return ErrUnknownPattern
	}
}

func (d *DoubleWriteDAO) Get(ctx context.Context, biz string, bizId int64) (Interactive, error) {
	switch d.pattern.Load() {
	case patternSrcOnly, patternSrcFirst:
		return d.src.Get(ctx, biz, bizId)
	case patternDstFirst, patternDstOnly:
		return d.dst.Get(ctx, biz, bizId)
	default:
		return Interactive{}, ErrUnknownPattern
	}
}

func (d *DoubleWriteDAO) InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error {
	switch d.pattern.Load() {
	case patternSrcOnly:
		return d.src.InsertCollectionBiz(ctx, cb)
	case patternSrcFirst:
		err := d.src.InsertCollectionBiz(ctx, cb)
		if err != nil {
			return err
		}
		err = d.dst.InsertCollectionBiz(ctx, cb)
		if err != nil {
			d.l.Error("写入 dst 失败！", logger.Error(err))
		}
		return nil
	case patternDstFirst:
		err := d.dst.InsertCollectionBiz(ctx, cb)
		if err != nil {
			return err
		}
		err = d.src.InsertCollectionBiz(ctx, cb)
		if err != nil {
			d.l.Error("写入 src 失败！", logger.Error(err))
		}
		return nil
	case patternDstOnly:
		return d.dst.InsertCollectionBiz(ctx, cb)
	default:
		return ErrUnknownPattern
	}
}

func (d *DoubleWriteDAO) GetCollectionInfo(ctx context.Context, biz string, bizId, uid int64) (UserCollectionBiz, error) {
	switch d.pattern.Load() {
	case patternSrcOnly, patternSrcFirst:
		return d.src.GetCollectionInfo(ctx, biz, bizId, uid)
	case patternDstFirst, patternDstOnly:
		return d.dst.GetCollectionInfo(ctx, biz, bizId, uid)
	default:
		return UserCollectionBiz{}, ErrUnknownPattern
	}
}

func (d *DoubleWriteDAO) GetByIds(ctx context.Context, biz string, ids []int64) ([]Interactive, error) {
	switch d.pattern.Load() {
	case patternSrcOnly, patternSrcFirst:
		return d.src.GetByIds(ctx, biz, ids)
	case patternDstFirst, patternDstOnly:
		return d.dst.GetByIds(ctx, biz, ids)
	default:
		return []Interactive{}, ErrUnknownPattern
	}
}
