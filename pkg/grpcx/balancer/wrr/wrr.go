package wrr

import (
	"context"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"sync"
)

const name = "custom_wrr"

// balancer.Picker 接口
// balancer.Balancer 接口
// balancer.Builder 接口
// base.PickerBuilder 接口
// 可以把 Balancer 看作 Picker 的装饰器，因为它不仅需要筛选节点，还需要管理连接池

func init() {
	balancer.Register(base.NewBalancerBuilder(name, &PickerBuilder{},
		base.Config{HealthCheck: true}))
}

type PickerBuilder struct {
}

func (p *PickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*conn, 0, len(info.ReadySCs))
	// sc => subConn
	// sci => subConnInfo
	for sc, sci := range info.ReadySCs {
		cc := &conn{
			cc: sc,
		}

		val, ok := sci.Address.Metadata.(map[string]any)
		if ok {
			weightVal := val["weight"]
			cc.weight = int(weightVal.(float64))
		}
		if cc.weight == 0 {
			// 没有数据的情况下可以考虑设置默认值
			cc.weight = 10
		}

		conns = append(conns, cc)
	}
	return &Picker{
		conns: conns,
	}
}

// Picker 执行负载均衡
type Picker struct {
	conns []*conn
	lock  sync.Mutex
}

// Pick 实现基于 wrr（加权轮询）负载均衡算法
func (p *Picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	// 没有候选节点
	if len(p.conns) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	var total int
	var maxCc *conn
	// 计算总权重 & 更新当前权重 & 选择节点
	for _, cc := range p.conns {
		if !cc.available {
			continue
		}
		total += cc.weight
		cc.currentWeight += cc.weight
		if maxCc == nil || cc.currentWeight > maxCc.currentWeight {
			maxCc = cc
		}
	}
	if maxCc.currentWeight > maxCc.threshold {
		// 超出阈值要限流
		maxCc.currentWeight = 0
	}
	// 更新
	maxCc.currentWeight -= total

	return balancer.PickResult{
		SubConn: maxCc.cc,
		// Done 回调方法。很多动态算法，根据调用结果来调整权重，就在这里
		Done: func(info balancer.DoneInfo) {
			err := info.Err
			if err == nil {
				// 可以考虑增加权重（weight/currentWeight）
				return
			}
			switch err {
			case context.Canceled:
				// 一般是手动取消，不用处理
				return
			case context.DeadlineExceeded:
				// 这里可以考虑降低权重
			case io.EOF:
				// 基本上这个节点不可用了（节点崩溃）
				maxCc.available = false
			default:
				s, ok := status.FromError(err)
				if ok {
					code := s.Code()
					switch code {
					case codes.Unavailable:
						// 可能是熔断，这个节点已经不可用，要考虑挪走该节点
						maxCc.available = false
						// 还需要开一个额外的 goroutine 去探活,借助 health check
						go func() {
							if p.healthCheck(maxCc) {
								// 通过健康检查之后可以放回
								maxCc.available = true
								// 最好将currentWeight设置的较低，或者一些流量控制的措施
								maxCc.currentWeight = 0
							}
						}()
					case codes.ResourceExhausted:
						// 可能是限流，可以考虑挪走或者适当降低权重
						maxCc.currentWeight = 0
					}
				}
			}
		},
	}, nil
}

// 调用 health check 接口
func (p *Picker) healthCheck(cc *conn) bool {
	//grpc.WithDisableHealthCheck()
	return true
}

type conn struct {
	// gRPC 中代表一个节点
	cc            balancer.SubConn
	weight        int
	currentWeight int
	available     bool
	threshold     int
}
