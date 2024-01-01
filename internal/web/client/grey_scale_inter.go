package client

import (
	interv1 "Prove/webook/api/proto/gen/inter/v1"
	"context"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"google.golang.org/grpc"
	"math/rand"
)

// GreyScaleInteractiveServiceClient 灰度控制客户端
// 作为改造的回滚方案，同时支持微服务调用和本地调用
// 怎么控制一个请求过来，是走微服务调用还是本地调用？
// 随机数 + 阈值
type GreyScaleInteractiveServiceClient struct {
	remote    interv1.InteractiveServiceClient
	local     interv1.InteractiveServiceClient
	threshold *atomicx.Value[int32]
}

func NewGreyScaleInteractiveServiceClient(remote interv1.InteractiveServiceClient,
	local interv1.InteractiveServiceClient) *GreyScaleInteractiveServiceClient {
	return &GreyScaleInteractiveServiceClient{
		remote:    remote,
		local:     local,
		threshold: atomicx.NewValue[int32](),
	}
}

func (g *GreyScaleInteractiveServiceClient) IncrReadCnt(ctx context.Context, in *interv1.IncrReadCntRequest, opts ...grpc.CallOption) (*interv1.IncrReadCntResponse, error) {
	return g.client().IncrReadCnt(ctx, in, opts...)
}

func (g *GreyScaleInteractiveServiceClient) Like(ctx context.Context, in *interv1.LikeRequest, opts ...grpc.CallOption) (*interv1.LikeResponse, error) {
	return g.client().Like(ctx, in, opts...)
}

func (g *GreyScaleInteractiveServiceClient) CancelLike(ctx context.Context, in *interv1.CancelLikeRequest, opts ...grpc.CallOption) (*interv1.CancelLikeResponse, error) {
	return g.client().CancelLike(ctx, in, opts...)
}

func (g *GreyScaleInteractiveServiceClient) Collect(ctx context.Context, in *interv1.CollectRequest, opts ...grpc.CallOption) (*interv1.CollectResponse, error) {
	return g.client().Collect(ctx, in, opts...)
}

func (g *GreyScaleInteractiveServiceClient) Get(ctx context.Context, in *interv1.GetRequest, opts ...grpc.CallOption) (*interv1.GetResponse, error) {
	return g.client().Get(ctx, in, opts...)
}

func (g *GreyScaleInteractiveServiceClient) GetByIds(ctx context.Context, in *interv1.GetByIdsRequest, opts ...grpc.CallOption) (*interv1.GetByIdsResponse, error) {
	return g.client().GetByIds(ctx, in, opts...)
}

func (g *GreyScaleInteractiveServiceClient) UpdateThreshold(newThreshold int32) {
	g.threshold.Store(newThreshold)
}

// 生成一个随机数，如果随机数小于等于阈值时，调用微服务；否则调用本地方法
func (g *GreyScaleInteractiveServiceClient) client() interv1.InteractiveServiceClient {
	num := rand.Int31n(100)
	if num <= g.threshold.Load() {
		return g.remote
	}
	return g.local
}
