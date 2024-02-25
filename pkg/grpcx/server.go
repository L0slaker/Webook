package grpcx

import (
	"Prove/webook/pkg/logger"
	"Prove/webook/pkg/netx"
	"context"
	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"google.golang.org/grpc"
	"net"
	"strconv"
	"time"
)

type Server struct {
	*grpc.Server
	EtcdAddr    []string
	Port        int
	Name        string
	L           logger.LoggerV1
	EtcdClient  *clientv3.Client
	cancel      func()
	etcdManager endpoints.Manager
	etcdKey     string
	EtcdTTL     int64
}

func (s *Server) Serve() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	s.cancel = cancel
	port := strconv.Itoa(s.Port)
	l, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	// 注册信息
	err = s.register(ctx, port)
	if err != nil {
		return err
	}

	return s.Server.Serve(l)
}

// register 服务注册
func (s *Server) register(ctx context.Context, port string) error {
	client := s.EtcdClient
	serviceName := "service/" + s.Name
	manager, err := endpoints.NewManager(client, serviceName)
	if err != nil {
		return err
	}
	s.etcdManager = manager

	ip := netx.GetOutboundIP()
	s.etcdKey = serviceName + "/" + ip
	addr := ip + ":" + port

	// 创建租约
	leaseResp, err := client.Grant(ctx, s.EtcdTTL)
	if err != nil {
		return err
	}

	// 续租
	ch, err := client.KeepAlive(ctx, leaseResp.ID)
	if err != nil {
		return err
	}
	go func() {
		for kaResp := range ch {
			// 输出内容
			s.L.Debug(kaResp.String())
		}
	}()

	return manager.AddEndpoint(ctx, s.etcdKey, endpoints.Endpoint{
		Addr: addr,
	}, clientv3.WithLease(leaseResp.ID))
}

func (s *Server) Shutdown() error {
	// 先取消续约
	s.cancel()
	// 通知注册中心，让其去通知客户端准备退出
	if s.etcdManager != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := s.etcdManager.DeleteEndpoint(ctx, s.etcdKey)
		if err != nil {
			return err
		}
	}
	// 关闭客户端
	err := s.EtcdClient.Close()
	if err != nil {
		return err
	}
	s.Server.GracefulStop()
	return nil
}
