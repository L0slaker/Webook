package metric

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"time"
)

type MiddlewareBuilder struct {
	Namespace  string
	Subsystem  string
	Name       string
	Help       string
	InstanceId string // 实例标识
}

func (m *MiddlewareBuilder) Build() gin.HandlerFunc {
	labels := []string{"pattern", "method", "status"}
	summaryVec := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: m.Namespace,
		Subsystem: m.Subsystem,
		Name:      m.Name + "_resp_time",
		Help:      m.Help,
		ConstLabels: map[string]string{
			"instance_id": m.InstanceId,
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.90:  0.005,
			0.98:  0.002,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, labels)

	// 使用 gauge 来统计当前活跃请求的数量
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: m.Namespace,
		Subsystem: m.Subsystem,
		Name:      m.Name + "_active_req",
		Help:      m.Help,
		ConstLabels: map[string]string{
			"instance_id": m.InstanceId,
		},
	})

	prometheus.MustRegister(summaryVec, gauge)
	return func(ctx *gin.Context) {
		start := time.Now()
		gauge.Inc()
		defer func() {
			duration := time.Since(start)
			gauge.Dec()
			pattern := ctx.FullPath()
			if pattern == "" {
				// 404 的情况
				pattern = "unknown"
			}
			summaryVec.WithLabelValues(pattern, ctx.Request.Method, strconv.Itoa(ctx.Writer.Status())).
				Observe(float64(duration.Milliseconds()))
		}()
		// 最终会执行到业务里面
		ctx.Next()
	}
}
