package client

import (
	interv1 "Prove/webook/api/proto/gen/inter/v1"
	"Prove/webook/interactive/domain"
	"Prove/webook/interactive/service"
	"context"
	"google.golang.org/grpc"
)

// InteractiveServiceAdapter 适配器，将本地实现伪装成 gRPC 客户端
type InteractiveServiceAdapter struct {
	svc service.InteractiveService
}

func NewInteractiveServiceAdapter(svc service.InteractiveService) *InteractiveServiceAdapter {
	return &InteractiveServiceAdapter{svc: svc}
}

func (i *InteractiveServiceAdapter) IncrReadCnt(ctx context.Context, in *interv1.IncrReadCntRequest, opts ...grpc.CallOption) (*interv1.IncrReadCntResponse, error) {
	err := i.svc.IncrReadCnt(ctx, in.GetBiz(), in.GetBizId())
	return &interv1.IncrReadCntResponse{}, err
}

func (i *InteractiveServiceAdapter) Like(ctx context.Context, in *interv1.LikeRequest, opts ...grpc.CallOption) (*interv1.LikeResponse, error) {
	err := i.svc.Like(ctx, in.GetBiz(), in.GetBizId(), in.GetUid())
	return &interv1.LikeResponse{}, err
}

func (i *InteractiveServiceAdapter) CancelLike(ctx context.Context, in *interv1.CancelLikeRequest, opts ...grpc.CallOption) (*interv1.CancelLikeResponse, error) {
	err := i.svc.CancelLike(ctx, in.GetBiz(), in.GetBizId(), in.GetUid())
	return &interv1.CancelLikeResponse{}, err
}

func (i *InteractiveServiceAdapter) Collect(ctx context.Context, in *interv1.CollectRequest, opts ...grpc.CallOption) (*interv1.CollectResponse, error) {
	err := i.svc.Collect(ctx, in.GetBiz(), in.GetBizId(), in.GetUid(), in.GetCid())
	return &interv1.CollectResponse{}, err
}

func (i *InteractiveServiceAdapter) Get(ctx context.Context, in *interv1.GetRequest, opts ...grpc.CallOption) (*interv1.GetResponse, error) {
	resp, err := i.svc.Get(ctx, in.GetBiz(), in.GetBizId(), in.GetUid())
	return &interv1.GetResponse{
		Inter: i.toDTO(resp),
	}, err
}

func (i *InteractiveServiceAdapter) GetByIds(ctx context.Context, in *interv1.GetByIdsRequest, opts ...grpc.CallOption) (*interv1.GetByIdsResponse, error) {
	inters, err := i.svc.GetByIds(ctx, in.GetBiz(), in.GetBizIds())
	if err != nil {
		return nil, err
	}
	m := make(map[int64]*interv1.Interactive, len(inters))
	for k, v := range inters {
		m[k] = i.toDTO(v)
	}
	return &interv1.GetByIdsResponse{
		Inters: m,
	}, err
}

// Data Transfer Object
func (i *InteractiveServiceAdapter) toDTO(inter domain.Interactive) *interv1.Interactive {
	return &interv1.Interactive{
		Biz:        inter.Biz,
		BizId:      inter.BizId,
		ReadCnt:    inter.ReadCnt,
		LikeCnt:    inter.LikeCnt,
		CollectCnt: inter.CollectCnt,
		Liked:      inter.Liked,
		Collected:  inter.Collected,
	}
}
