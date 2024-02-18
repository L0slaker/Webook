package fixer

import (
	"Prove/webook/pkg/migrator"
	"Prove/webook/pkg/migrator/events"
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OverrideFixer[T migrator.Entity] struct {
	base    *gorm.DB
	target  *gorm.DB
	columns []string
}

func NewOverrideFixer[T migrator.Entity](base, target *gorm.DB) (*OverrideFixer[T], error) {
	var t T
	rows, err := target.Model(&t).Limit(1).Rows()
	if err != nil {
		return nil, err
	}
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	return &OverrideFixer[T]{
		base:    base,
		target:  target,
		columns: columns,
	}, nil
}

func (f *OverrideFixer[T]) Fix(ctx context.Context, id int64) error {
	var src T
	// 找出数据
	err := f.base.WithContext(ctx).Where("id = ?", id).
		First(&src).Error
	switch err {
	// 找到了数据
	case nil:
		return f.target.Clauses(&clause.OnConflict{
			// 我们需要 Entity 告诉我们，修复哪些数据
			DoUpdates: clause.AssignmentColumns(f.columns),
		}).Create(&src).Error
	case gorm.ErrRecordNotFound:
		return f.target.Delete("id = ?", id).Error
	default:
		return err
	}
}

// FixV1 修复数据
// 数据不相等时/校验的目标数据缺失时：UPSERT
// 校验的基础数据缺失：DELETE
func (f *OverrideFixer[T]) FixV1(ctx context.Context, evt events.InconsistentEvent) error {
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
func (f *OverrideFixer[T]) FixV2(ctx context.Context, evt events.InconsistentEvent) error {
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
func (f *OverrideFixer[T]) FixV3(ctx context.Context, evt events.InconsistentEvent) error {
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
