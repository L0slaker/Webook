package fixer

import (
	"Prove/webook/migrator"
	"Prove/webook/migrator/events"
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Fixer[T migrator.Entity] struct {
	base    *gorm.DB
	target  *gorm.DB
	columns []string
}

// FixV1 修复数据
// 数据不相等时/校验的目标数据缺失时：UPSERT
// 校验的基础数据缺失：DELETE
func (f *Fixer[T]) FixV1(ctx context.Context, evt events.InconsistentEvent) error {
	switch evt.Type {
	case events.InconsistentEventTypeNEQ,
		events.InconsistentEventTypeTM:
		var t T
		err := f.base.WithContext(ctx).Where("id = ?", evt.ID).First(&t).Error
		switch err {
		case nil:
			return f.target.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns(f.columns),
			}).Create(&t).Error
		case gorm.ErrRecordNotFound:
			return f.target.WithContext(ctx).Where("id = ?", evt.ID).Delete(&t).Error
		default:
			return err
		}
	case events.InconsistentEventTypeBM:
		var t T
		return f.base.WithContext(ctx).Where("id = ?", evt.ID).Delete(&t).Error
	default:
		return errors.New("未知的不一致类型")
	}
}

// FixV2 修复数据
// 当数据不相等时：UPDATE
// 校验的目标数据缺失时：INSERT
// 校验的基础数据缺失：DELETE
func (f *Fixer[T]) FixV2(ctx context.Context, evt events.InconsistentEvent) error {
	switch evt.Type {
	case events.InconsistentEventTypeNEQ:
		var t T
		err := f.base.WithContext(ctx).Where("id = ?", evt.ID).First(&t).Error
		switch err {
		case nil:
			return f.target.WithContext(ctx).Updates(&t).Error
		case gorm.ErrRecordNotFound:
			return f.target.WithContext(ctx).Where("id = ?", evt.ID).Delete(&t).Error
		default:
			return err
		}
	case events.InconsistentEventTypeTM:
		var t T
		err := f.base.WithContext(ctx).Where("id = ?", evt.ID).First(&t).Error
		switch err {
		case nil:
			return f.target.WithContext(ctx).Create(&t).Error
		case gorm.ErrRecordNotFound:
			return nil
		default:
			return err
		}
	case events.InconsistentEventTypeBM:
		var t T
		return f.base.WithContext(ctx).Where("id = ?", evt.ID).Delete(&t).Error
	default:
		return errors.New("未知的不一致类型")
	}
}

// FixV3 修复数据
// 纯覆盖的写法，当校验的基础数据存在时，UPSERT
// 当校验的基础数据缺失时： DELETE
func (f *Fixer[T]) FixV3(ctx context.Context, evt events.InconsistentEvent) error {
	var t T
	err := f.base.WithContext(ctx).Where("id = ?", evt.ID).First(&t).Error
	switch err {
	case nil:
		return f.target.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns(f.columns),
		}).Create(&t).Error
	case gorm.ErrRecordNotFound:
		return f.target.WithContext(ctx).
			Where("id = ?", evt.ID).Delete(&t).Error
	default:
		return err
	}
}
