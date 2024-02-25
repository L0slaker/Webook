package ioc

import (
	grpc2 "Prove/webook/interactive/grpc"
	"Prove/webook/pkg/grpcx"
	"Prove/webook/pkg/logger"
	"github.com/spf13/viper"
	"go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

func InitGRPCxServer(interServer *grpc2.InteractiveServiceServer,
	etcdClient *clientv3.Client, l logger.LoggerV1) *grpcx.Server {
	type Config struct {
		Port     int    `yaml:"port"`
		EtcdAddr string `yaml:"etcdAddr"`
		EtcdTTL  int64  `yaml:"etcdTTL"`
	}

	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}

	server := grpc.NewServer()
	interServer.Register(server)

	return &grpcx.Server{
		Server:     server,
		Port:       cfg.Port,
		Name:       "interactive",
		L:          l,
		EtcdClient: etcdClient,
		EtcdTTL:    cfg.EtcdTTL,
	}
}
