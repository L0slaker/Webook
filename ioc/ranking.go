package ioc

import (
	"Prove/webook/internal/job"
	"Prove/webook/internal/service"
	"Prove/webook/pkg/logger"
	rlock "github.com/gotomicro/redis-lock"
	"github.com/robfig/cron/v3"
	"time"
)

func InitRankingJob(svc service.RankingService, l logger.LoggerV1,
	client *rlock.Client) *job.RankingJob {
	return job.NewRankingJob(svc, time.Minute, l, client)
}

func InitJobs(l logger.LoggerV1, rankingJob *job.RankingJob) *cron.Cron {
	res := cron.New(cron.WithSeconds())
	cbd := job.NewCronJobBuilder(l)
	// 每五分钟一次
	_, err := res.AddJob("0 */5 * * * ?", cbd.Build(rankingJob))
	if err != nil {
		panic(err)
	}
	return res
}
