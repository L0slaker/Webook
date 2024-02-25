package grpc

import (
	"context"
)

var _ UserServiceServer = &Server{}

type Server struct {
	UnimplementedUserServiceServer // 不实现的方法
	Name                           string
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
			Id:   2,
			Name: "Lisa",
		},
	}, nil
}
