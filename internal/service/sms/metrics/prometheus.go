package metrics

import (
	"Prove/webook/internal/service/sms"
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

// PrometheusDecorator 利用装饰器封装Prometheus
type PrometheusDecorator struct {
	svc    sms.Service
	vector *prometheus.SummaryVec
}

func NewPrometheusDecorator(svc sms.Service) sms.Service {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "geekbang_l0slakers",
		Subsystem: "webook",
		Name:      "sms_resp_time",
		Help:      "统计 sms 服务的性能数据",
	}, []string{"biz"})
	prometheus.MustRegister(vector)

	return &PrometheusDecorator{
		svc:    svc,
		vector: vector,
	}
}

func (p *PrometheusDecorator) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Milliseconds()
		p.vector.WithLabelValues(biz).Observe(float64(duration))
	}()
	return p.svc.Send(ctx, biz, args, numbers...)
}
