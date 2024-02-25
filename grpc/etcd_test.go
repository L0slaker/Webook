package grpc

import (
	"Prove/webook/pkg/netx"
	"context"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
	"time"
)

type EtcdTestSuite struct {
	suite.Suite
	client *etcdv3.Client
}

func (s *EtcdTestSuite) SetupSuite() {
	client, err := etcdv3.New(etcdv3.Config{
		Endpoints: []string{"localhost:12379"},
	})
	require.NoError(s.T(), err)
	s.client = client
}

// 查询服务
// etcdctl --endpoints=localhost:12379 get service/user --prefix

func (s *EtcdTestSuite) TestServer() {
	l, err := net.Listen("tcp", ":8090")
	require.NoError(s.T(), err)

	// 控制创建租约的 context
	grantCtx, grantCancel := context.WithTimeout(context.Background(), time.Second)
	defer grantCancel()
	// ttl 租期，单位是秒
	var ttl int64 = 30
	leaseResp, err := s.client.Grant(grantCtx, ttl)
	require.NoError(s.T(), err)

	// endpoint 以服务为维度，一个服务一个 Manager； target 服务
	manager, err := endpoints.NewManager(s.client, "service/user")
	require.NoError(s.T(), err)
	//addr := "127.0.0.1:8090"
	addr := netx.GetOutboundIP()
	// key 实例的值
	key := "service/user/" + addr

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 在此之前完成所有的启动工作，包括缓存预加载之类的事情
	err = manager.AddEndpoint(ctx, key, endpoints.Endpoint{
		Addr: addr,
	}, etcdv3.WithLease(leaseResp.ID))
	require.NoError(s.T(), err)

	// 操作续约
	kaCtx, kaCancel := context.WithCancel(context.Background())
	go func() {
		ch, err1 := s.client.KeepAlive(kaCtx, leaseResp.ID)
		require.NoError(s.T(), err1)
		for kaResp := range ch {
			s.T().Log(kaResp.String(), time.Now().String())
		}
	}()

	// 监听注册信息的变动
	go func() {
		ticker := time.NewTicker(time.Second * 5)
		for now := range ticker.C {
			ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second)
			err = manager.AddEndpoint(ctx1, key, endpoints.Endpoint{
				Addr:     addr,
				Metadata: now.String(),
			})
			require.NoError(s.T(), err)
			//err1 := manager.Update(ctx1,[]*endpoints.UpdateWithOpts{
			//	// 可以声明很多个 Update
			//	{
			//		Update: endpoints.Update{
			//			Op:       endpoints.Add,
			//			Key:      key, // key 需要是唯一的
			//			Endpoint: endpoints.Endpoint{
			//				Addr: addr,
			//				Metadata: now,
			//			},
			//		},
			//	},
			//})
			cancel1()
		}
	}()

	server := grpc.NewServer()
	RegisterUserServiceServer(server, &Server{})
	err = server.Serve(l)
	s.T().Log(err)

	// 准备退出
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 先取消续约
	kaCancel()
	// 优雅下线的第一步，通知注册中心，让其去通知客户端准备退出
	err = manager.DeleteEndpoint(ctx, key)
	require.NoError(s.T(), err)
	// 关闭客户端
	s.client.Close()
	server.GracefulStop()
}

func (s *EtcdTestSuite) TestServerV2() {
	// target 服务
	target := "service/user"
	// endpoint 以服务为维度，一个服务一个 Manager
	manager, err := endpoints.NewManager(s.client, target)
	require.NoError(s.T(), err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	addr := "127.0.0.1:8091"
	// key 实例的值
	key := "service/user/" + addr
	err = manager.AddEndpoint(ctx, key, endpoints.Endpoint{
		Addr: addr,
	})
	require.NoError(s.T(), err)

	// 监听注册信息的变动
	go func() {
		ticker := time.NewTicker(time.Second * 5)
		for now := range ticker.C {
			ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second)
			manager.AddEndpoint(ctx1, key, endpoints.Endpoint{
				Addr:     addr,
				Metadata: now.String(),
			})
			cancel1()
		}
	}()

	l, err := net.Listen("tcp", ":8091")
	require.NoError(s.T(), err)
	server := grpc.NewServer()
	RegisterUserServiceServer(server, &Server{})
	err = server.Serve(l)
	s.T().Log(err)

	// 退出时删除 endpoint
	// 优雅下线的第一步，通知注册中心，让其去通知客户端准备退出
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = manager.DeleteEndpoint(ctx, key)
	require.NoError(s.T(), err)
	server.GracefulStop()
}

func (s *EtcdTestSuite) TestClient() {
	builder, err := resolver.NewBuilder(s.client)
	require.NoError(s.T(), err)

	clientConn, err := grpc.Dial("etcd:///service/user",
		grpc.WithResolvers(builder),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(s.T(), err)

	client := NewUserServiceClient(clientConn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := client.GetById(ctx, &GetByIdReq{
		Id: 123,
	})
	require.NoError(s.T(), err)

	s.T().Log(resp.User)
}

func TestEtcd(t *testing.T) {
	suite.Run(t, new(EtcdTestSuite))
}
