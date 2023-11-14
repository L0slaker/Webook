package wechat

import (
	"Prove/webook/internal/domain"
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type PrometheusDecorator struct {
	svc Service
	sum prometheus.Summary
}

func NewPrometheusDecorator(svc Service, namespace, subsystem, name, help, instanceId string) Service {
	sum := prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      name,
		Help:      help,
		ConstLabels: map[string]string{
			"instance_id": instanceId,
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.90:  0.005,
			0.98:  0.002,
			0.99:  0.001,
			0.999: 0.0001,
		},
	})
	prometheus.MustRegister(sum)

	return &PrometheusDecorator{
		svc: svc,
		sum: sum,
	}
}

func (p *PrometheusDecorator) AuthURL(ctx context.Context, state string) (string, error) {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Milliseconds()
		p.sum.Observe(float64(duration))
	}()
	return p.svc.AuthURL(ctx, state)
}

func (p *PrometheusDecorator) VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error) {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Milliseconds()
		p.sum.Observe(float64(duration))
	}()
	return p.svc.VerifyCode(ctx, code)
}
