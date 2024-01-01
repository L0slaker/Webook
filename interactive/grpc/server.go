package grpc

import (
	"Prove/webook/api/proto/gen/inter/v1"
	"Prove/webook/interactive/domain"
	"Prove/webook/interactive/service"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// InteractiveServiceServer 将 service 包装成一个 gRPC 接口
type InteractiveServiceServer struct {
	svc service.InteractiveService
	interv1.UnimplementedInteractiveServiceServer
}

func NewInteractiveServiceServer(svc service.InteractiveService) *InteractiveServiceServer {
	return &InteractiveServiceServer{
		svc: svc,
	}
}

func (i *InteractiveServiceServer) Register(server *grpc.Server) {
	interv1.RegisterInteractiveServiceServer(server, i)
}

func (i *InteractiveServiceServer) IncrReadCnt(ctx context.Context, request *interv1.IncrReadCntRequest) (*interv1.IncrReadCntResponse, error) {
	err := i.svc.IncrReadCnt(ctx, request.GetBiz(), request.GetBizId())
	return &interv1.IncrReadCntResponse{}, err
}

func (i *InteractiveServiceServer) Like(ctx context.Context, request *interv1.LikeRequest) (*interv1.LikeResponse, error) {
	err := i.svc.Like(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	return &interv1.LikeResponse{}, err
}

func (i *InteractiveServiceServer) CancelLike(ctx context.Context, request *interv1.CancelLikeRequest) (*interv1.CancelLikeResponse, error) {
	// 校验也可以使用 gRPC 的插件
	if request.Uid <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "uid 错误")
	}
	err := i.svc.CancelLike(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	return &interv1.CancelLikeResponse{}, err
}

func (i *InteractiveServiceServer) Collect(ctx context.Context, request *interv1.CollectRequest) (*interv1.CollectResponse, error) {
	err := i.svc.Collect(ctx, request.GetBiz(), request.GetBizId(), request.GetUid(), request.GetCid())
	return &interv1.CollectResponse{}, err
}

func (i *InteractiveServiceServer) Get(ctx context.Context, request *interv1.GetRequest) (*interv1.GetResponse, error) {
	inter, err := i.svc.Get(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	if err != nil {
		return nil, err
	}
	return &interv1.GetResponse{
		Inter: i.toDTO(inter),
	}, nil
}

func (i *InteractiveServiceServer) GetByIds(ctx context.Context, request *interv1.GetByIdsRequest) (*interv1.GetByIdsResponse, error) {
	inters, err := i.svc.GetByIds(ctx, request.GetBiz(), request.GetBizIds())
	if err != nil {
		return nil, err
	}
	m := make(map[int64]*interv1.Interactive, len(inters))
	for k, v := range inters {
		m[k] = i.toDTO(v)
	}
	return &interv1.GetByIdsResponse{
		Inters: m,
	}, nil
}

// Data Transfer Object
func (i *InteractiveServiceServer) toDTO(inter domain.Interactive) *interv1.Interactive {
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
