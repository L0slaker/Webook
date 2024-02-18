package validator

import (
	"Prove/webook/pkg/logger"
	"Prove/webook/pkg/migrator"
	"Prove/webook/pkg/migrator/events"
	"context"
	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"time"
)

// Validator T必须实现Entity接口
type Validator[T migrator.Entity] struct {
	// 1.如果以源表为准(第二阶段)，那么 base = 源表，target = 目标表
	// 2.如果以目标表为准(第三阶段)，那么 base = 目标表，target = 源表
	base          *gorm.DB // 校验，以 XXX 为准
	target        *gorm.DB // 校验的数据
	direction     string   // 基准
	batchSize     int      // 批量
	l             logger.LoggerV1
	p             events.Producer
	updateTime    int64         // 初始化时指定 update_time
	sleepInterval time.Duration // 当 <= 0时，直接退出校验循环；否则睡眠
	//highLoad      *atomicx.Value[bool]
	//order         string        // 以 xx 排序
}

func NewValidator[T migrator.Entity](base, target *gorm.DB, direction string,
	l logger.LoggerV1, p events.Producer) *Validator[T] {
	//highLoad := atomicx.NewValueOf[bool](false)
	return &Validator[T]{
		base:          base,
		target:        target,
		direction:     direction,
		batchSize:     100,
		l:             l,
		p:             p,
		sleepInterval: 0, // 默认是全量校验
		//highLoad:  highLoad,
	}
}

//// NewIncrValidator 增量校验
//func NewIncrValidator[T migrator.Entity](base, target *gorm.DB, l logger.LoggerV1,
//	direction string, p events.Producer) {
//	v := NewValidator[T](base, target, l, direction, p)
//	v.order = "update_time"
//}
//
//// NewFullValidator 全量校验
//func NewFullValidator[T migrator.Entity](base, target *gorm.DB, l logger.LoggerV1,
//	direction string, p events.Producer) {
//	v := NewValidator[T](base, target, l, direction, p)
//	v.order = "id"
//}

func (v *Validator[T]) Validate(ctx context.Context) error {
	var eg errgroup.Group
	eg.Go(func() error {
		return v.baseToTarget(ctx)
	})
	eg.Go(func() error {
		return v.targetToBase(ctx)
	})
	return eg.Wait()
}

// 借助 utime 实现增量校验，要保证utime要有一个独立的索引，以便提升查询的速度
func (v *Validator[T]) baseToTarget(ctx context.Context) error {
	offset := 0
	for {
		//if v.highLoad.Load() {
		//	// 负载过高时，暂时挂起校验
		//}
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		var src T
		err := v.base.WithContext(dbCtx).Where("update_time > ?", v.updateTime).
			Offset(offset).First(&src).Error
		cancel()
		switch err {
		case nil: // 正常拿到了数据，去 target 中拿取数 据比较
			v.dstDiff(ctx, src)
		case gorm.ErrRecordNotFound: // 校验完了所有数据
			// 如果要支持全量校验和增量校验，这里就不能直接退出，因为后续还会有新数据产生
			// 有些情况下用户可能希望退出，有些情况下希望继续(sleep)，所以我们交由用户来决定
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		case context.DeadlineExceeded, context.Canceled:
			return nil
		default: // 未知的数据库错误
			v.l.Error("查询 base 出错", logger.Error(err))
		}
		offset++
	}
}

func (v *Validator[T]) dstDiff(ctx context.Context, src T) {
	var dst T
	err := v.target.Where("id = ?", dst.ID()).First(&dst).Error
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
		v.l.Error("查询 target 出错", logger.Error(err))
	}
}

// validateTargetToBase 反向校验，找出 target 中存在而 base 中不存在的数据
func (v *Validator[T]) targetToBase(ctx context.Context) error {
	offset := 0
	for {
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		var dstTs []T
		err := v.target.WithContext(dbCtx).Select("id").
			Where("update_time > ?", v.updateTime).
			Limit(v.batchSize).Find(&dstTs).Error
		cancel()
		// 没有数据直接返回
		if len(dstTs) == 0 {
			if v.sleepInterval > 0 {
				time.Sleep(v.sleepInterval)
				continue
			}
		}

		switch err {
		case nil:
			v.srcMissingRecords(ctx, dstTs)
		case gorm.ErrRecordNotFound: //没有数据，直接返回
			if v.sleepInterval > 0 {
				time.Sleep(v.sleepInterval)
				continue
			}
		case context.DeadlineExceeded, context.Canceled:
			return nil
		default: // 数据库未知错误
			v.l.Error("查询 target 失败", logger.Error(err))
		}
		offset += len(dstTs)
		if len(dstTs) < v.batchSize {
			// 没数据了
			if v.sleepInterval > 0 {
				time.Sleep(v.sleepInterval)
				continue
			}
		}
	}
}

func (v *Validator[T]) srcMissingRecords(ctx context.Context, dstTs []T) {
	ids := slice.Map(dstTs, func(idx int, t T) int64 {
		return t.ID()
	})
	var srcTs []T
	err := v.base.Where("id IN ?", ids).Find(&srcTs).Error
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
	}
}

func (v *Validator[T]) notify(ctx context.Context, id int64, typ string) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	err := v.p.ProduceInconsistentEvent(ctx, events.InconsistentEvent{
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

func (v *Validator[T]) SleepInterval(i time.Duration) *Validator[T] {
	v.sleepInterval = i
	return v
}

func (v *Validator[T]) UpdateTime(utime int64) *Validator[T] {
	v.updateTime = utime
	return v
}
