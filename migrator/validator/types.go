package validator

import (
	"Prove/webook/migrator"
	"Prove/webook/migrator/events"
	"Prove/webook/pkg/logger"
	"context"
	"github.com/ecodeclub/ekit/slice"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"time"
)

// Validator T必须实现Entity接口
type Validator[T migrator.Entity] struct {
	// 1.如果以源表为准(第二阶段)，那么 base = 源表，target = 目标表
	// 2.如果以目标表为准(第三阶段)，那么 base = 目标表，target = 源表
	base      *gorm.DB // 校验，以 XXX 为准
	target    *gorm.DB // 校验的数据
	direction string   // 基准
	batchSize int      // 批量
	l         logger.LoggerV1
	p         events.Producer
	highLoad  *atomicx.Value[bool]
}

func (v *Validator[T]) Validate(ctx context.Context) error {
	var eg errgroup.Group
	eg.Go(func() error {
		v.validateBaseToTarget(ctx)
		return nil
	})
	eg.Go(func() error {
		v.validateTargetToBase(ctx)
		return nil
	})
	return eg.Wait()
}

func (v *Validator[T]) validateBaseToTarget(ctx context.Context) {
	offset := -1
	for {
		if v.highLoad.Load() {
			// 负载过高时，暂时挂起校验
		}
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		offset++
		var src T
		err := v.base.WithContext(dbCtx).Offset(offset).First(&src).Error
		cancel()
		switch err {
		case nil: // 正常拿到了数据，去 target 中拿取数据比较
			var dst T
			err = v.target.Where("id = ?", dst.ID()).First(&dst).Error
			switch err {
			case nil:
				if !src.CompareTo(dst) {
					// 不相等，上报 kafka
					v.notify(ctx, src.ID(), events.InconsistentEventTypeNEQ)
				}
			case gorm.ErrRecordNotFound:
				// 缺少数据
				v.notify(ctx, src.ID(), events.InconsistentEventTypeTM)
			default:
			}
		case gorm.ErrRecordNotFound: // 校验完了所有数据
			return
		default: // 未知的数据库错误
			v.l.Error("校验数据，查询 base 出错", logger.Error(err))
			continue
		}
	}
}

// validateTargetToBase 反向校验，找出 target 中存在而 base 中不存在的数据
func (v *Validator[T]) validateTargetToBase(ctx context.Context) {
	offset := -v.batchSize
	for {
		offset += v.batchSize
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		var dstTs []T
		err := v.target.WithContext(dbCtx).Select("id").Offset(offset).
			Limit(v.batchSize).Order("id").Find(&dstTs).Error
		cancel()
		if len(dstTs) == 0 {
			return
		}

		switch err {
		case nil:
			ids := slice.Map(dstTs, func(idx int, t T) int64 {
				return t.ID()
			})
			var srcTs []T
			err = v.base.Where("id IN ?", ids).Find(&srcTs).Error
			switch err {
			case nil:
				// 计算差集，即 src 中缺少的数据
				srcIds := slice.Map(srcTs, func(idx int, t T) int64 {
					return t.ID()
				})
				diff := slice.DiffSet(ids, srcIds)
				v.notifyBaseMissing(ctx, diff)
			case gorm.ErrRecordNotFound:
				// 没有数据
				v.notifyBaseMissing(ctx, ids)
			default:
				v.l.Error("未知错误", logger.Error(err))
				continue
			}
		case gorm.ErrRecordNotFound: //没有数据，直接返回
			return
		default: // 数据库未知错误
			v.l.Error("未知错误", logger.Error(err))
			continue
		}

		if len(dstTs) < v.batchSize {
			// 没数据了
			return
		}
	}
}

func (v *Validator[T]) notify(ctx context.Context, id int64, typ string) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	err := v.p.ProduceInconsistentEvents(ctx, events.InconsistentEvent{
		ID:        id,
		Direction: v.direction,
		Type:      typ,
	})
	cancel()
	if err != nil {
		// 重试、日志、告警、手动处理
		v.l.Error("发送数据缺失的消息失败", logger.Error(err))
	}
}

func (v *Validator[T]) notifyBaseMissing(ctx context.Context, ids []int64) {
	for _, id := range ids {
		v.notify(ctx, id, events.InconsistentEventTypeBM)
	}
}
