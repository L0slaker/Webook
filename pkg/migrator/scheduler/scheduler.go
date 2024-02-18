package scheduler

import (
	"Prove/webook/pkg/ginx"
	"Prove/webook/pkg/gormx/connpool"
	"Prove/webook/pkg/logger"
	"Prove/webook/pkg/migrator"
	"Prove/webook/pkg/migrator/events"
	"Prove/webook/pkg/migrator/validator"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"sync"
	"time"
)

// Scheduler 统一管理整个迁移过程
// 它不是必须的，这是为了方便用户操作（和你理解）而引入的。
type Scheduler[T migrator.Entity] struct {
	lock       sync.Mutex
	src        *gorm.DB
	dst        *gorm.DB
	pool       *connpool.DoubleWritePool
	l          logger.LoggerV1
	pattern    string
	cancelFull func()
	cancelIncr func()
	producer   events.Producer
}

func NewScheduler[T migrator.Entity](l logger.LoggerV1, src, dst *gorm.DB,
	pool *connpool.DoubleWritePool, producer events.Producer) *Scheduler[T] {
	return &Scheduler[T]{
		l:       l,
		src:     src,
		dst:     dst,
		pattern: connpool.PatternSrcOnly,
		cancelFull: func() {
			// 初始的时候，啥也不用做
		},
		cancelIncr: func() {
			// 初始的时候，啥也不用做
		},
		pool:     pool,
		producer: producer,
	}
}

// RegisterRoutes 可以考虑监听配置中心，根据变化调整
func (s *Scheduler[T]) RegisterRoutes(server *gin.RouterGroup) {
	server.POST("/src_only", ginx.Wrap(s.SrcOnly))
	server.POST("/dst_only", ginx.Wrap(s.DstOnly))
	server.POST("/src_first", ginx.Wrap(s.SrcFirst))
	server.POST("/dst_first", ginx.Wrap(s.DstFirst))
	server.POST("/full/start", ginx.Wrap(s.StartFullValidation))
	server.POST("/full/stop", ginx.Wrap(s.StopFullValidation))
	server.POST("/incr/stop", ginx.Wrap(s.StopIncrValidation))
	server.POST("/incr/start", ginx.WrapReq[StartIncrRequest](s.StartIncrValidation))
}

func (s *Scheduler[T]) SrcOnly(_ *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternSrcOnly
	s.pool.ChangePattern(connpool.PatternSrcOnly)
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (s *Scheduler[T]) DstOnly(_ *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternDstOnly
	s.pool.ChangePattern(connpool.PatternDstOnly)
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (s *Scheduler[T]) SrcFirst(_ *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternSrcFirst
	s.pool.ChangePattern(connpool.PatternSrcFirst)
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (s *Scheduler[T]) DstFirst(_ *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternDstFirst
	s.pool.ChangePattern(connpool.PatternDstFirst)
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (s *Scheduler[T]) StartFullValidation(_ *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	// 取消上一次的校验
	cancel := s.cancelFull
	v, err := s.newValidator()
	if err != nil {
		return ginx.Result{}, err
	}
	var ctx context.Context
	ctx, s.cancelFull = context.WithCancel(context.Background())

	go func() {
		// 取消上一次的校验
		cancel()
		err := v.Validate(ctx)
		if err != nil {
			s.l.Warn("退出全量校验", logger.Error(err))
		}
	}()
	return ginx.Result{
		Msg: "启动全量校验成功",
	}, nil
}

func (s *Scheduler[T]) StopFullValidation(_ *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cancelFull()
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (s *Scheduler[T]) StartIncrValidation(c *gin.Context,
	req StartIncrRequest) (ginx.Result, error) {
	// 开启增量校验
	s.lock.Lock()
	defer s.lock.Unlock()
	// 取消上一次的
	cancel := s.cancelIncr
	v, err := s.newValidator()
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统异常",
		}, nil
	}
	v.SleepInterval(time.Duration(req.Interval) * time.Millisecond).UpdateTime(req.UpdateTime)
	var ctx context.Context
	ctx, s.cancelIncr = context.WithCancel(context.Background())

	go func() {
		cancel()
		err := v.Validate(ctx)
		s.l.Warn("退出增量校验", logger.Error(err))
	}()
	return ginx.Result{
		Msg: "启动增量校验成功",
	}, nil
}

func (s *Scheduler[T]) StopIncrValidation(_ *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cancelIncr()
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (s *Scheduler[T]) newValidator() (*validator.Validator[T], error) {
	switch s.pattern {
	case connpool.PatternSrcFirst, connpool.PatternSrcOnly:
		return validator.NewValidator[T](s.src, s.dst, "SRC", s.l, s.producer), nil
	case connpool.PatternDstFirst, connpool.PatternDstOnly:
		return validator.NewValidator[T](s.dst, s.src, "DST", s.l, s.producer), nil
	default:
		return nil, fmt.Errorf("未知的 pattern %s", s.pattern)
	}
}

type StartIncrRequest struct {
	UpdateTime int64 `json:"update_time"`
	// 毫秒数
	// json 不能正确处理 time.Duration 类型
	Interval int64 `json:"interval"`
}
