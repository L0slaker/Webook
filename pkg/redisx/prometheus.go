package redisx

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"net"
	"strconv"
	"time"
)

type PrometheusHook struct {
	vector *prometheus.SummaryVec
}

func NewPrometheusHook(opt prometheus.SummaryOpts) *PrometheusHook {
	// 监控使用的命令、是否命中缓存
	vector := prometheus.NewSummaryVec(opt, []string{"cmd", "key_exist"})
	prometheus.MustRegister(vector)

	return &PrometheusHook{
		vector: vector,
	}
}

// DialHook 连接redis调用
func (p *PrometheusHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		// 没有处理
		return next(ctx, network, addr)
	}
}

// ProcessHook 发送普通命令时调用
func (p *PrometheusHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		// redis 执行之前
		startTime := time.Now()
		var err error
		defer func() {
			duration := time.Since(startTime).Milliseconds()
			keyExist := err == redis.Nil
			p.vector.WithLabelValues(cmd.Name(), strconv.FormatBool(keyExist)).
				Observe(float64(duration))
		}()
		// 这个会最终发送命令到 redis 上
		err = next(ctx, cmd)
		// redis 执行之后
		return err
	}
}

// ProcessPipelineHook 使用 pipeline 功能时调用
func (p *PrometheusHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		return next(ctx, cmds)
	}
}
