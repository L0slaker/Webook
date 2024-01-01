package grpcx

import (
	"google.golang.org/grpc"
	"net"
)

type Server struct {
	*grpc.Server
	Addr string
}

func (s *Server) Serve() error {
	l, err := net.Listen("tcp", ":8090")
	if err != nil {
		return err
	}
	// 阻塞
	return s.Server.Serve(l)
}
