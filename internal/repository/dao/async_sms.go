package dao

import (
	"context"
	"github.com/ecodeclub/ekit/sqlx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

const (
	asyncStatusWaiting = iota
	// 失败，且超过重试次数
	asyncStatusFailed
	asyncStatusSuccess
)

var ErrWaitingSMSNotFound = gorm.ErrRecordNotFound

type AsyncSmsDAO interface {
	Insert(ctx context.Context, dao AsyncSms) error
	GetWaitingSMS(ctx context.Context) (AsyncSms, error)
	MarkSuccess(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64) error
}

type GORMAsyncSmsDAO struct {
	db *gorm.DB
}

func NewGORMAsyncSmsDAO(db *gorm.DB) AsyncSmsDAO {
	return &GORMAsyncSmsDAO{
		db: db,
	}
}

func (g *GORMAsyncSmsDAO) Insert(ctx context.Context, s AsyncSms) error {
	return g.db.WithContext(ctx).Create(&s).Error
}

func (g *GORMAsyncSmsDAO) GetWaitingSMS(ctx context.Context) (AsyncSms, error) {
	// 如果在高并发情况下，SELECT for UPDATE 对数据库压力很大
	// 但是我们不是高并发，因为你部署 N 台机器，才有 N 个goroutine来查询
	// 并发不过百，随便写
	var s AsyncSms
	err := g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 为了避免偶发性的失败，我们只找 1 分钟前的异步短信发送
		now := time.Now().UnixMilli()
		endTime := now - time.Minute.Milliseconds()
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("update_time < ? and status = ?", endTime, asyncStatusWaiting).First(&s).Error
		// SELECT xx FROM xxx WHERE xx FOR UPDATE，锁住了
		if err != nil {
			return err
		}
		// 只要更新了更新时间，根据前面的规则，就不可能被别的节点抢占了
		err = tx.Model(&AsyncSms{}).Where("id = ?", s.Id).Updates(map[string]any{
			"retry_cnt": gorm.Expr("retry_cnt + 1"),
			// 更新成了当前时间戳，确保在发送过程中，没人会再次抢占。相当于重试间隔一分钟
			"update_time": now,
		}).Error
		return err
	})
	return s, err
}

func (g *GORMAsyncSmsDAO) MarkSuccess(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Model(&AsyncSms{}).Where("id = ?", id).Updates(map[string]any{
		"update_time": now,
		"status":      asyncStatusSuccess,
	}).Error
}

func (g *GORMAsyncSmsDAO) MarkFailed(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Model(&AsyncSms{}).Where("id = ?,id").Updates(map[string]any{
		"update_time": now,
		"status":      asyncStatusFailed,
	}).Error
}

type AsyncSms struct {
	Id         int64
	Config     sqlx.JsonColumn[SmsConfig]
	RetryCnt   int
	RetryMax   int
	Status     uint8
	CreateTime int64
	UpdateTime int64 `gorm:"index"`
}

type SmsConfig struct {
	TplId   string
	Args    []string
	Numbers []string
}
