package nacaos

import (
	grpc2 "Prove/grpc"
	"Prove/webook/pkg/netx"
	"context"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
)

type NacosTestSuite struct {
	suite.Suite
	client naming_client.INamingClient
}

func (s *NacosTestSuite) SetupSuite() {
	clientConfig := constant.ClientConfig{
		NamespaceId:         "e525eafa-f7d7-4029-83d9-008937f9d468",
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "tmp/nacos/cache",
		LogLevel:            "debug",
	}

	// at least one ServerConfig
	serverConfig := []constant.ServerConfig{
		{
			Scheme:      "http",
			ContextPath: "/nacos",
			IpAddr:      "localhost",
			Port:        8848,
		},
	}
	cli, err := clients.CreateNamingClient(map[string]interface{}{
		"serverConfigs": serverConfig,
		"clientConfig":  clientConfig,
	})
	require.NoError(s.T(), err)
	s.client = cli
}

func (s *NacosTestSuite) TestServer() {
	l, err := net.Listen("tcp", ":8090")
	require.NoError(s.T(), err)
	server := grpc.NewServer()
	grpc2.RegisterUserServiceServer(server, &grpc2.Server{})
	ok, err := s.client.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          netx.GetOutboundIP(),
		Port:        8090,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		ServiceName: "user",
	})
	require.NoError(s.T(), err)
	require.True(s.T(), ok)
	err = server.Serve(l)
	s.T().Log(err)
}

func (s *NacosTestSuite) TestClient() {
	rb := &nacosResolverBuilder{
		client: s.client,
	}
	cc, err := grpc.Dial("nacos:///user", grpc.WithResolvers(rb),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(s.T(), err)
	client := grpc2.NewUserServiceClient(cc)
	resp, err := client.GetById(context.Background(), &grpc2.GetByIdReq{Id: 123})
	require.NoError(s.T(), err)
	s.T().Log(resp.User)
}

func TestGoZero(t *testing.T) {
	suite.Run(t, new(NacosTestSuite))
}
