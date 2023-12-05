package prometheus

import "github.com/prometheus/client_golang/prometheus"

// 计数器，统计次数。只能增加，不能减少
func Counter() {
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "my-namespace",
		Subsystem: "my-subsystem",
		Name:      "test-counter",
	})
	prometheus.MustRegister(counter)
	// +1
	counter.Inc()
	// 必须是正数
	counter.Add(12)
}

// 度量：可以增加也可以减少
func Gauge() {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "my-namespace",
		Subsystem: "my-subsystem",
		Name:      "test-gauge",
	})
	prometheus.MustRegister(gauge)
	gauge.Set(12)
	gauge.Add(10.2)
	gauge.Add(-3)
	gauge.Sub(3)
}

// 柱状图，对观察对象进行采样，然后分到一个个桶里
func Histogram() {
	histogram := prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "my-namespace",
		Subsystem: "my-subsystem",
		Name:      "test-histogram",
		Buckets:   []float64{10, 50, 100, 200, 500, 1000, 10000},
	})
	prometheus.MustRegister(histogram)
	histogram.Observe(12.4)
}

// 采样点按照百分位进行统计，比如说99线，999线等
func Summary() {
	summary := prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace: "my-namespace",
		Subsystem: "my-subsystem",
		Name:      "test-summary",
		Objectives: map[float64]float64{
			// key 是百分比，value 是误差
			// 0.5：50% 以内的响应时间是多长，误差在 1 %
			// 0.75：75% 以内的响应时间是多长，误差在 1 %
			// 0.90：90% 以内的响应时间是多长，误差在 0.5 %
			0.5:   0.01,
			0.75:  0.01,
			0.90:  0.005,
			0.98:  0.002,
			0.99:  0.001,
			0.999: 0.0001,
		},
	})
	prometheus.MustRegister(summary)
	summary.Observe(12.3)
}

func Vector() {
	summaryVec := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Subsystem: "http_request",
		Name:      "geekbang",
		ConstLabels: map[string]string{
			"server":  "localhost:9091",
			"env":     "test",
			"appname": "test_app",
		},
		Help: "this static info for http request",
	}, []string{"pattern", "method", "status"})

	// 当次请求 pattern = /user/:id，method = post 和 status = 200时，响应时间为128
	summaryVec.WithLabelValues("/user/:id", "POST", "'200").Observe(128)
}
