package ioc

import (
	"Prove/webook/internal/job"
	"Prove/webook/internal/service"
	"Prove/webook/pkg/logger"
	"github.com/robfig/cron/v3"
	"time"
)

func InitRankingJob(svc service.RankingService) *job.RankingJob {
	return job.NewRankingJob(svc, time.Minute)
}

func InitJobs(l logger.LoggerV1, rankingJob *job.RankingJob) *cron.Cron {
	res := cron.New(cron.WithSeconds())
	cbd := job.NewCronJobBuilder(l)
	// 每五分钟一次
	res.AddJob("0 */5 * * * ?", cbd.Build(rankingJob))
	return res
}
