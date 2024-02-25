package nacaos

import (
	"errors"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"google.golang.org/grpc/resolver"
)

type nacosResolverBuilder struct {
	client naming_client.INamingClient
}

func (n *nacosResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	res := &nacosResolver{
		client: n.client,
		target: target,
		cc:     cc,
	}
	return res, res.subscribe()
}

func (n *nacosResolverBuilder) Scheme() string {
	return "nacos"
}

type nacosResolver struct {
	client naming_client.INamingClient
	target resolver.Target
	cc     resolver.ClientConn
}

func (r *nacosResolver) ResolveNow(options resolver.ResolveNowOptions) {
	svcs, err := r.client.SelectInstances(vo.SelectInstancesParam{
		ServiceName: r.target.Endpoint(),
	})
	if err != nil {
		r.cc.ReportError(err)
		return
	}
	if len(svcs) == 0 {
		r.cc.ReportError(errors.New("无候选可用节点"))
		return
	}
}

// 订阅 Nacos 中的服务实例变化
func (r *nacosResolver) subscribe() error {
	return r.client.Subscribe(&vo.SubscribeParam{
		ServiceName: r.target.Endpoint(),
		SubscribeCallback: func(services []model.Instance, err error) {
			if err != nil {
				// 记录日志
				return
			}
			err = r.reportAddrs(services)
			if err != nil {
				// 记录日志
				return
			}
		},
	})
}

// 将服务实例信息转换为 gRPC Resolver 地址并更新到 gRPC 客户端连接中。
func (r *nacosResolver) reportAddrs(svcs []model.Instance) error {
	addrs := make([]resolver.Address, 0, len(svcs))
	for _, svc := range svcs {
		addrs = append(addrs, resolver.Address{
			Addr: fmt.Sprintf("%s:%d", svc.Ip, svc.Port),
		})
	}
	return r.cc.UpdateState(resolver.State{
		Addresses: addrs,
	})
}

func (r *nacosResolver) Close() {
	// 不需要处理
}
