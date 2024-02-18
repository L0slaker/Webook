package connpool

import (
	"Prove/webook/pkg/logger"
	"context"
	"database/sql"
	"errors"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"gorm.io/gorm"
)

const (
	PatternSrcOnly  = "SRC_ONLY"
	PatternSrcFirst = "SRC_FIRST"
	PatternDstFirst = "DST_FIRST"
	PatternDstOnly  = "DST_ONLY"
)

var errUnknownPattern = errors.New("未知的双写模式")

type DoubleWritePool struct {
	src     gorm.ConnPool
	dst     gorm.ConnPool
	pattern *atomicx.Value[string]
	l       logger.LoggerV1
}

func NewDoubleWritePool(src, dst *gorm.DB, l logger.LoggerV1) *DoubleWritePool {
	return &DoubleWritePool{
		src:     src.ConnPool,
		dst:     dst.ConnPool,
		pattern: atomicx.NewValueOf(PatternSrcOnly),
		l:       l,
	}
}

// UpdatePattern 更改双写模式
// 但在开启事务时，就无法更改了，要想办法在有事务未提交的情况下，禁止修改
func (d *DoubleWritePool) UpdatePattern(pattern string) {
	d.pattern.Store(pattern)
}

// BeginTx 开启事务
func (d *DoubleWritePool) BeginTx(ctx context.Context, opts *sql.TxOptions) (gorm.ConnPool, error) {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		tx, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		return &DoubleWritePoolTx{
			src:     tx,
			pattern: pattern,
		}, err
	case PatternSrcFirst:
		return d.startTwoTx(ctx, d.src, d.dst, pattern, opts)
	case PatternDstFirst:
		return d.startTwoTx(ctx, d.dst, d.src, pattern, opts)
	case PatternDstOnly:
		tx, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		return &DoubleWritePoolTx{
			dst:     tx,
			pattern: pattern,
		}, err
	default:
		return nil, errUnknownPattern
	}
}

func (d *DoubleWritePool) startTwoTx(ctx context.Context, first, second gorm.ConnPool,
	pattern string, opts *sql.TxOptions) (gorm.ConnPool, error) {
	txSrc, err := first.(gorm.TxBeginner).BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	txDst, err := second.(gorm.TxBeginner).BeginTx(ctx, opts)
	if err != nil {
		// 日志
	}
	return &DoubleWritePoolTx{
		src:     txSrc,
		dst:     txDst,
		pattern: pattern,
	}, nil
}

// PrepareContext 用于准备一个 SQL 语句，返回一个预处理的语句对象
// 但是 sql.Stmt 是一个结构体，我们没办法返回一个代表双写的 Stmt
func (d *DoubleWritePool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return nil, errors.New("双写模式下不支持")
}

// ExecContext 用于执行一个 SQL，通常是用于执行 INSERT、UPDATE、DELETE 等操作
func (d *DoubleWritePool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	switch d.pattern.Load() {
	case PatternSrcOnly:
		return d.src.ExecContext(ctx, query, args...)
	case PatternSrcFirst:
		res, err := d.src.ExecContext(ctx, query, args...)
		if err != nil {
			return res, err
		}
		res, err = d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			d.l.Error("写入 dst 时出错，等待修复", logger.Error(err))
		}
		return res, err
	case PatternDstFirst:
		res, err := d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			return res, err
		}
		res, err = d.src.ExecContext(ctx, query, args...)
		if err != nil {
			d.l.Error("写入 src 时出错，等待修复", logger.Error(err))
		}
		return res, err
	case PatternDstOnly:
		return d.dst.ExecContext(ctx, query, args...)
	default:
		return nil, errUnknownPattern
	}
}

// QueryContext 用于执行一个 SQL，通常是用于执行 SELECT 操作
func (d *DoubleWritePool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch d.pattern.Load() {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryContext(ctx, query, args...)
	case PatternDstFirst, PatternDstOnly:
		return d.dst.QueryContext(ctx, query, args...)
	default:
		return nil, errUnknownPattern
	}
}

// QueryRowContext 用于执行一个 SQL 查询，返回结果的第一行
func (d *DoubleWritePool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch d.pattern.Load() {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryRowContext(ctx, query, args...)
	case PatternDstFirst, PatternDstOnly:
		return d.dst.QueryRowContext(ctx, query, args...)
	default:
		// 如何构造返回一个 error？
		//return nil
		panic(errUnknownPattern)
	}
}

func (d *DoubleWritePool) ChangePattern(pattern string) {
	d.pattern.Store(pattern)
}

type DoubleWritePoolTx struct {
	src     *sql.Tx
	dst     *sql.Tx
	pattern string // 事务可以并行，不需要原子操作
}

func (d *DoubleWritePoolTx) Commit() error {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.Commit()
	case PatternSrcFirst:
		err := d.src.Commit()
		if err != nil {
			return err
		}
		if d.dst != nil {
			err = d.dst.Commit()
			if err != nil {
				// 日志
			}
		}
		return nil
	case PatternDstFirst:
		err := d.dst.Commit()
		if err != nil {
			return err
		}
		if d.src != nil {
			err = d.src.Commit()
			if err != nil {
				// 日志
			}
		}
		return nil
	case PatternDstOnly:
		return d.dst.Commit()
	default:
		return errUnknownPattern
	}
}

func (d *DoubleWritePoolTx) Rollback() error {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.Rollback()
	case PatternSrcFirst:
		err := d.src.Rollback()
		if err != nil {
			return err
		}
		if d.dst != nil {
			err = d.dst.Rollback()
			if err != nil {
				// 日志
			}
		}
		return nil
	case PatternDstFirst:
		err := d.dst.Rollback()
		if err != nil {
			return err
		}
		if d.src != nil {
			err = d.src.Rollback()
			if err != nil {
				// 日志
			}
		}
		return nil
	case PatternDstOnly:
		return d.dst.Rollback()
	default:
		return errUnknownPattern
	}
}

func (d *DoubleWritePoolTx) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return nil, errors.New("双写模式下不支持")
}

func (d *DoubleWritePoolTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.ExecContext(ctx, query, args...)
	case PatternSrcFirst:
		res, err := d.src.ExecContext(ctx, query, args...)
		if err != nil {
			return res, err
		}
		// dst 开启事务失败
		if d.dst == nil {
			return res, err
		}
		res, err = d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			//d.l.Error("写入 dst 时出错，等待修复", logger.Error(err))
		}
		return res, err
	case PatternDstFirst:
		res, err := d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			return res, err
		}
		// src 开启事务失败
		if d.src == nil {
			return res, err
		}
		res, err = d.src.ExecContext(ctx, query, args...)
		if err != nil {
			//d.l.Error("写入 src 时出错，等待修复", logger.Error(err))
		}
		return res, err
	case PatternDstOnly:
		return d.dst.ExecContext(ctx, query, args...)
	default:
		return nil, errUnknownPattern
	}
}

func (d *DoubleWritePoolTx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch d.pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryContext(ctx, query, args...)
	case PatternDstFirst, PatternDstOnly:
		return d.dst.QueryContext(ctx, query, args...)
	default:
		return nil, errUnknownPattern
	}
}

func (d *DoubleWritePoolTx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch d.pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryRowContext(ctx, query, args...)
	case PatternDstFirst, PatternDstOnly:
		return d.dst.QueryRowContext(ctx, query, args...)
	default:
		panic(errUnknownPattern)
	}
}
