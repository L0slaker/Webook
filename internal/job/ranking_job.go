package job

import (
	"Prove/webook/internal/service"
	"context"
	"time"
)

type RankingJob struct {
	svc     service.RankingService
	timeout time.Duration // 运行的超时时间
}

func NewRankingJob(svc service.RankingService, timeout time.Duration) *RankingJob {
	return &RankingJob{
		svc: svc,
		// 过期时间要根据数据量来计算
		timeout: timeout,
	}
}

func (r *RankingJob) Name() string {
	return "ranking"
}

func (r *RankingJob) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.svc.TopN(ctx)
}
