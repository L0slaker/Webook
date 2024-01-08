package connpool

import (
	"context"
	"database/sql"
	"gorm.io/gorm"
)

// WriteSplit 读写分离
type WriteSplit struct {
	master gorm.ConnPool
	slaves []gorm.ConnPool
}

func (w *WriteSplit) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return w.master.(gorm.TxBeginner).BeginTx(ctx, opts)
}

func (w *WriteSplit) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return w.master.PrepareContext(ctx, query)
}

func (w *WriteSplit) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return w.master.ExecContext(ctx, query, args...)
}

func (w *WriteSplit) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	// slaves 要考虑负载均匀，比如增加轮询、加权轮询、随机或者加权随机等
	// 还可以动态判定 slaves 监控情况的负载均衡策略（挑返回响应最快的）
	panic("implement me")
}

func (w *WriteSplit) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	//TODO implement me
	panic("implement me")
}
