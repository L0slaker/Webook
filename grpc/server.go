package grpc

import (
	"context"
)

var _ UserServiceServer = &Server{}

type Server struct {
	UnimplementedUserServiceServer // 不实现的方法
}

func (s *Server) GetById(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	return &GetByIdResp{
		User: &User{
			Id:   1,
			Name: "Alice",
		},
	}, nil
}

func (s *Server) GetByIdV1(ctx context.Context, req *GetByIdReq) (*GetByIdResp, error) {
	return &GetByIdResp{
		User: &User{
			Id:   1,
			Name: "Alice",
		},
	}, nil
}
