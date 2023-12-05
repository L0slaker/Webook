package ioc

import (
	"Prove/webook/internal/domain"
	"Prove/webook/internal/job"
	"Prove/webook/internal/service"
	"Prove/webook/pkg/logger"
	"context"
	"time"
)

func InitScheduler(svc service.JobService, l logger.LoggerV1,
	local *job.LocalFuncExecutor) *job.Scheduler {
	res := job.NewScheduler(svc, l)
	res.RegisterExecutor(local)
	return res
}

func InitLocalFuncExecutor(svc service.RankingService) *job.LocalFuncExecutor {
	res := job.NewLocalFuncExecutor()
	// 要在数据库中也插入一条 ranking job 的记录
	res.RegisterFunc("ranking", func(ctx context.Context, j domain.Job) error {
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()
		return svc.TopN(ctx)
	})
	return res
}
