package dao

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"time"
)

const (
	jobStatusWaiting = iota // 可抢的任务
	jobStatusRunning        // 已被抢占的任务
	jobStatusPaused         // 暂停调度的任务
)

type JobDAO interface {
	Preempt(ctx context.Context) (Job, error)
	Release(ctx context.Context, id int64) error
	UpdateUTime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, next time.Time) error
	Stop(ctx context.Context, id int64) error
}

type Job struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	Cfg        string // 配置
	Status     int    // 任务状态
	NextTime   int64  `gorm:"index"` // 下一次被调度的时间
	Version    int
	Cron       string
	Name       string `gorm:"unique"` // 任务名称
	Executor   string // 执行器
	CreateTime int64
	UpdateTime int64
}

type GormJobDAO struct {
	db *gorm.DB
}

func NewGormJobDAO(db *gorm.DB) JobDAO {
	return &GormJobDAO{
		db: db,
	}
}

// Preempt 引入了版本链，一定程度上优化了性能，类似于使用乐观锁替代悲观锁的场景
func (dao *GormJobDAO) Preempt(ctx context.Context) (Job, error) {
	// 高并发情况下，大部分情况都没啥用
	// 假如有 100 个goroutine，那么特定一个goroutine最差的情况下可能要循环一百次
	// 那总共可能就要转 1+2+3...+99+100 次
	db := dao.db.WithContext(ctx)
	for {
		now := time.Now().UnixMilli()
		var j Job
		// if next_time <= now && status == 0，这时候就可以抢占任务了
		err := db.WithContext(ctx).Where("status = ? AND next_time <= ?",
			jobStatusWaiting, now).First(&j).Error
		if err != nil {
			return Job{}, err
		}
		// 找到了可被抢占的任务，进行抢占
		res := db.Where("id = ? AND version = ?",
			j.Id, j.Version).Model(&Job{}).Updates(map[string]any{
			"status":      jobStatusRunning,
			"update_time": now,
			"version":     j.Version + 1,
		})
		if res.Error != nil {
			return Job{}, err
		}
		if res.RowsAffected == 0 {
			continue
		}
		return j, nil
	}
}

func (dao *GormJobDAO) Release(ctx context.Context, id int64) error {
	// 释放操作也要同时防止释放掉其他任务
	db := dao.db.WithContext(ctx)
	var j Job
	err := db.WithContext(ctx).Where("id = ?", id).First(&j).Error
	if err != nil {
		return err
	}

	res := db.WithContext(ctx).Where("id = ? AND version = ?",
		id, j.Version).Model(&Job{}).Updates(map[string]any{
		"status":      jobStatusWaiting,
		"update_time": time.Now().UnixMilli(),
		"version":     j.Version + 1,
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("该任务未运行，无法释放！")
	}
	return nil
}

func (dao *GormJobDAO) UpdateUTime(ctx context.Context, id int64) error {
	return dao.db.WithContext(ctx).Where("id = ?", id).
		Model(&Job{}).Updates(map[string]any{
		"update_time": time.Now().UnixMilli(),
	}).Error
}

func (dao *GormJobDAO) UpdateNextTime(ctx context.Context, id int64, next time.Time) error {
	return dao.db.WithContext(ctx).Where("id = ?", id).
		Model(&Job{}).Updates(map[string]any{
		"next_time": next.UnixMilli(),
	}).Error
}

func (dao *GormJobDAO) Stop(ctx context.Context, id int64) error {
	return dao.db.WithContext(ctx).Where("id = ?", id).
		Model(&Job{}).Updates(map[string]any{
		"status":      jobStatusPaused,
		"update_time": time.Now().UnixMilli(),
	}).Error
}

//// PreemptV2 不使用版本链，要额外引入悲观锁
//func (dao *GormJobDAO) PreemptV2(ctx context.Context) (Job, error) {
//	for {
//		now := time.Now().UnixMilli()
//		var j Job
//		// 两个事务同时查询相同的数据，并且其中一个事务在另一个事务完成之前进行了修改，
//		// 那第一个事务读取的数据可能已经过时，不再准确。
//		err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
//			res := tx.Model(&Job{}).Where("status = ? AND next_time <= ?",
//				jobStatusWaiting, now).Set("gorm:query_option", "FOR UPDATE").First(&j)
//			if res.Error != nil {
//				return res.Error
//			}
//			// 找到了可被抢占的任务，进行抢占
//			res = dao.db.Where("id = ? AND status = ?",
//				j.Id, jobStatusWaiting).Updates(map[string]any{
//				"status":      jobStatusRunning,
//				"update_time": now,
//			})
//			if res.Error != nil {
//				return res.Error
//			}
//			if res.RowsAffected == 0 {
//				return errors.New("没抢到！")
//			}
//			return nil
//		})
//		return j, err
//	}
//}
