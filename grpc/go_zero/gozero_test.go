package go_zero

import (
	grpc2 "Prove/grpc"
	"context"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"testing"
	"time"
)

type GoZeroTestSuite struct {
	suite.Suite
}

func (s *GoZeroTestSuite) SetupSuite() {

}

func (s *GoZeroTestSuite) TestServer() {
	go func() {
		s.startServer(":8090")
	}()
	s.startServer(":8091")
}

func (s *GoZeroTestSuite) TestClient() {
	c := zrpc.MustNewClient(zrpc.RpcClientConf{
		Etcd: discov.EtcdConf{
			Hosts: []string{"localhost:12379"},
			// 服务名
			Key: "user",
		},
	}, zrpc.WithDialOption(
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`),
	))
	client := grpc2.NewUserServiceClient(c.Conn())
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.GetById(ctx, &grpc2.GetByIdReq{Id: 123})
		cancel()
		require.NoError(s.T(), err)
		s.T().Log(resp.User)
	}
}

func (s *GoZeroTestSuite) startServer(addr string) {
	c := zrpc.RpcServerConf{
		ListenOn: addr,
		Etcd: discov.EtcdConf{
			Hosts: []string{"localhost:12379"},
			// 服务名
			Key: "user",
		},
	}

	// 创建一个服务器，并注册服务实例
	server := zrpc.MustNewServer(c, func(grpcServer *grpc.Server) {
		grpc2.RegisterUserServiceServer(grpcServer, &grpc2.Server{
			Name: addr,
		})
	})

	// 增加拦截器，或插件
	//server.AddUnaryInterceptors(interceptor)
	server.Start()
}

func TestGoZero(t *testing.T) {
	suite.Run(t, new(GoZeroTestSuite))
}
