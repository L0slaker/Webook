package job

import (
	"Prove/webook/pkg/logger"
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"strconv"
	"time"
)

type CronJobBuilder struct {
	l      logger.LoggerV1
	vector *prometheus.SummaryVec
	tracer trace.Tracer
}

func NewCronJobBuilder(l logger.LoggerV1) *CronJobBuilder {
	vec := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "geekbang_l0slakers",
		Subsystem: "webook",
		Name:      "cron_job",
		Help:      "统计定时任务的执行情况",
	}, []string{"name", "success"})
	prometheus.MustRegister(vec)
	return &CronJobBuilder{
		l:      l,
		vector: vec,
		tracer: otel.GetTracerProvider().Tracer("src/Prove/webook/internal/job/job_builder.go"),
	}
}

func (c *CronJobBuilder) Build(job Job) cron.Job {
	name := job.Name()
	return cronJobFuncAdapter(func() error {
		_, span := c.tracer.Start(context.Background(), name)
		defer span.End()
		start := time.Now()
		c.l.Info("任务开始", logger.String("job", name))
		var success bool
		defer func() {
			c.l.Info("任务结束", logger.String("job", name))
			duration := time.Since(start).Milliseconds()
			c.vector.WithLabelValues(name,
				strconv.FormatBool(success)).Observe(float64(duration))
		}()
		err := job.Run()
		success = err == nil
		if err != nil {
			span.RecordError(err)
			c.l.Error("任务运行失败", logger.String("job", name))
		}
		return nil
	})
}

type cronJobFuncAdapter func() error

func (c cronJobFuncAdapter) Run() {
	_ = c()
}
