package job

import (
	"Prove/webook/internal/service"
	"Prove/webook/pkg/logger"
	"context"
	rlock "github.com/gotomicro/redis-lock"
	"sync"
	"time"
)

type RankingJob struct {
	svc       service.RankingService
	timeout   time.Duration // 运行的超时时间
	client    *rlock.Client // 分布式锁
	key       string
	l         logger.LoggerV1
	lock      *rlock.Lock // 全局维护的锁
	localLock *sync.Mutex // 本地锁
}

func NewRankingJob(svc service.RankingService, timeout time.Duration,
	l logger.LoggerV1, client *rlock.Client) *RankingJob {
	return &RankingJob{
		svc: svc,
		// 过期时间要根据数据量来计算
		timeout: timeout,
		client:  client,
		key:     "rlock:cron_job:ranking",
		l:       l,
	}
}

func (r *RankingJob) Name() string {
	return "ranking"
}

// Run 按照时间进行调度，N分钟一次
func (r *RankingJob) Run() error {
	// 没拿到锁，尝试拿锁
	// 可以有两种策略，第一种就是重试，第二种是监听删除事件
	if r.lock == nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		// Lock 的时候使用分布式任务的间隔时间
		// N分钟调用一次，就加锁N分钟，最后也不释放锁
		lock, err := r.client.Lock(ctx, r.key, r.timeout, &rlock.FixIntervalRetry{
			Interval: time.Millisecond * 100,
			Max:      5,
		}, time.Second)
		if err != nil {
			return err
		}
		r.lock = lock
		go func() {
			// 自动续约，保证一直持有分布式锁
			err1 := lock.AutoRefresh(r.timeout/2, time.Second)
			// 续约失败了，可以重试续约，也可以考虑不处理，下一次再抢锁
			// 可能是服务出问题了，应该让出锁
			if err1 != nil {
				r.l.Error("续约失败", logger.Error(err))
			}
			// 多个任务同时试图更新 r.lock，可能会引发意外的行为
			// 通过本地锁，确保了在同一时刻只有一个任务能够执行续约
			r.localLock.Lock()
			r.lock = nil
			r.localLock.Unlock()
		}()
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.svc.TopN(ctx)
}

// Close 释放资源，不释放的话也可以
// 关机之后，分布式锁没有人续约，就会让别人拿到锁
func (r *RankingJob) Close() error {
	r.localLock.Lock()
	lock := r.lock
	r.lock = nil
	r.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}

//// RunV2 我们每次进来都要去竞争锁，毫无意义
//func (r *RankingJob) RunV2() error {
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
//	defer cancel()
//	lock, err := r.client.Lock(ctx, r.key, r.timeout, &rlock.FixIntervalRetry{
//		Interval: time.Millisecond * 100,
//		Max:      5,
//	}, time.Second)
//	if err != nil {
//		return err
//	}
//
//	defer func() {
//		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
//		defer cancel()
//		err = lock.Unlock(ctx)
//		if err != nil {
//			r.l.Error("", logger.Error(err))
//		}
//	}()
//	return r.svc.TopN(ctx)
//}
