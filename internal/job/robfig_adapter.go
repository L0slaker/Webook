package job

import (
	"Prove/webook/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type RankingJobAdapter struct {
	j Job
	l logger.LoggerV1
	p prometheus.Summary
}

func NewRankingJobAdapter(j Job, l logger.LoggerV1) *RankingJobAdapter {
	p := prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "cron_job",
		ConstLabels: map[string]string{
			"name": j.Name(),
		},
	})
	prometheus.MustRegister(p)
	return &RankingJobAdapter{
		j: j,
		l: l,
		p: p,
	}
}

func (r *RankingJobAdapter) Run() {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		r.p.Observe(float64(duration))
	}()
	err := r.j.Run()
	if err != nil {
		r.l.Error("任务运行失败！", logger.Error(err),
			logger.String("job", r.j.Name()))
	}
}
