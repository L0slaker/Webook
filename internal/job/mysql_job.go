package job

import (
	"Prove/webook/internal/domain"
	"Prove/webook/internal/service"
	"Prove/webook/pkg/logger"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/sync/semaphore"
	"net/http"
	"time"
)

// Executor 执行器
type Executor interface {
	Name() string
	// Exec ctx是整个调度器的上下文，当从 ctx.Done 中有信号时，就要考虑结束执行
	Exec(ctx context.Context, j domain.Job) error
}

// HTTPExecutor HTTP任务
type HTTPExecutor struct{}

func (H *HTTPExecutor) Name() string {
	return "http"
}

func (H *HTTPExecutor) Exec(ctx context.Context, j domain.Job) error {
	type Config struct {
		Endpoint string
		Method   string
	}
	var cfg Config
	err := json.Unmarshal([]byte(j.Cfg), &cfg)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(cfg.Method, cfg.Endpoint, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("执行失败！")
	}
	return nil
}

// GRPCExecutor gRPC任务
type GRPCExecutor struct{}

func (H *GRPCExecutor) Name() string {
	return "grpc"
}

func (H *GRPCExecutor) Exec(ctx context.Context, j domain.Job) error {
	//TODO implement me
	panic("implement me")
}

// LocalFuncExecutor 本地方法
type LocalFuncExecutor struct {
	// 存储了不同的已注册的执行函数
	funcs map[string]func(ctx context.Context, j domain.Job) error
}

func NewLocalFuncExecutor() *LocalFuncExecutor {
	return &LocalFuncExecutor{
		funcs: make(map[string]func(ctx context.Context, j domain.Job) error),
	}
}

func (l *LocalFuncExecutor) Name() string {
	return "local"
}

func (l *LocalFuncExecutor) RegisterFunc(name string, fn func(ctx context.Context, j domain.Job) error) {
	l.funcs[name] = fn
}

func (l *LocalFuncExecutor) Exec(ctx context.Context, j domain.Job) error {
	fn, ok := l.funcs[j.Name]
	if !ok {
		return fmt.Errorf("未知任务，你是否注册了：%s", j.Name)
	}
	return fn(ctx, j)
}

// Scheduler 调度器
type Scheduler struct {
	execs   map[string]Executor
	svc     service.JobService
	l       logger.LoggerV1
	limiter *semaphore.Weighted
}

func NewScheduler(svc service.JobService, l logger.LoggerV1) *Scheduler {
	return &Scheduler{
		execs: make(map[string]Executor),
		svc:   svc,
		l:     l,
	}
}

func (s *Scheduler) RegisterExecutor(exec Executor) {
	s.execs[exec.Name()] = exec
}

func (s *Scheduler) Schedule(ctx context.Context) error {
	for {
		// main 函数退出，推出调度循环
		// 主要是考虑通过 ctx 来控制整个调度器的上下文
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err := s.limiter.Acquire(ctx, 1)
		if err != nil {
			return err
		}
		// 抢占任务
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		j, err := s.svc.Preempt(dbCtx)
		cancel()
		if err != nil {
			// s.limiter.Release(1)
			s.l.Error("抢占任务失败", logger.Error(err))
			// 继续下一轮
			continue
		}

		// 获取执行器
		exec, ok := s.execs[j.Name]
		if !ok {
			// DEBUG 时最好中断; 线上情况可以继续
			s.l.Error("未找到对应的执行器", logger.String("executor", j.Executor))
			continue
		}

		// 执行任务，考虑异步执行，主要是为了不影响主调度的循环
		// 这里抢占任务时，可能会有非常多的 goroutine 来抢占任务
		go func() {
			defer func() { // 执行完毕，释放
				s.limiter.Release(1)
				e := j.CancelFunc()
				if e != nil {
					s.l.Error("释放任务失败！",
						logger.Error(e), logger.Int64("jid", j.Id))
				}
			}()
			// 执行要考虑任务的超时控制
			err1 := exec.Exec(ctx, j)
			if err1 != nil {
				s.l.Error("任务执行失败", logger.Error(err1))
			}
			// 考虑下一次调度
			rsCtx, cancel2 := context.WithTimeout(context.Background(), time.Second)
			err1 = s.svc.ResetNextTime(rsCtx, j)
			cancel2()
			if err1 != nil {
				s.l.Error("设置下一次执行时间失败", logger.Error(err1))
			}
		}()
	}
}
