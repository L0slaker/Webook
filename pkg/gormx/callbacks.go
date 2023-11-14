package gormx

import (
	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
	"time"
)

// Callbacks 监控和捕获 GORM 相关的指标(查询时间)
type Callbacks struct {
	vector *prometheus.SummaryVec
}

func NewCallbacks(Namespace, Subsystem, Name, Help string, ConstLabels map[string]string) *Callbacks {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:   Namespace,
		Subsystem:   Subsystem,
		Name:        Name,
		Help:        Help,
		ConstLabels: ConstLabels,
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
	}, []string{"type", "table"})
	prometheus.MustRegister(vector)
	return &Callbacks{vector: vector}
}

func (c *Callbacks) RegisterAll(db *gorm.DB) {
	err := db.Callback().Create().Before("*").
		Register("prometheus_create_before", c.before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Create().After("*").
		Register("prometheus_create_after", c.after("create"))
	if err != nil {
		panic(err)
	}

	err = db.Callback().Update().Before("*").
		Register("prometheus_update_before", c.before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Update().After("*").
		Register("prometheus_update_after", c.after("update"))
	if err != nil {
		panic(err)
	}

	err = db.Callback().Delete().Before("*").
		Register("prometheus_delete_before", c.before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Delete().After("*").
		Register("prometheus_delete_after", c.after("delete"))
	if err != nil {
		panic(err)
	}

	err = db.Callback().Query().Before("*").
		Register("prometheus_query_before", c.before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Query().After("*").
		Register("prometheus_query_after", c.after("query"))
	if err != nil {
		panic(err)
	}

	err = db.Callback().Raw().Before("*").
		Register("prometheus_raw_before", c.before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Raw().After("*").
		Register("prometheus_raw_after", c.after("raw"))
	if err != nil {
		panic(err)
	}

	err = db.Callback().Row().Before("*").
		Register("prometheus_row_before", c.before())
	if err != nil {
		panic(err)
	}
	err = db.Callback().Row().After("*").
		Register("prometheus_row_after", c.after("row"))
	if err != nil {
		panic(err)
	}
}

func (c *Callbacks) before() func(db *gorm.DB) {
	return func(db *gorm.DB) {
		startTime := time.Now()
		db.Set("start_time", startTime)
	}
}

func (c *Callbacks) after(typ string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		val, _ := db.Get("start_time")
		startTime, ok := val.(time.Time)
		if !ok {
			return
		}
		duration := time.Since(startTime)
		// 上报 Prometheus
		table := db.Statement.Table
		if table == "" {
			table = "unknown"
		}
		c.vector.WithLabelValues(typ, table).Observe(float64(duration.Milliseconds()))
	}
}
