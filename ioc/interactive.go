package ioc

import (
	interv1 "Prove/webook/api/proto/gen/inter/v1"
	"Prove/webook/interactive/service"
	"Prove/webook/internal/web/client"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitInteractiveGRPCClient(svc service.InteractiveService) interv1.InteractiveServiceClient {
	type Config struct {
		Addr      string
		Secure    bool
		Threshold int32
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.inter", &cfg)
	if err != nil {
		panic(err)
	}
	var opts []grpc.DialOption
	if cfg.Secure {
		// 上面要加载证书之类的东西，启用 HTTPS
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	conn, err := grpc.Dial(cfg.Addr, opts...)
	if err != nil {
		panic(err)
	}

	remote := interv1.NewInteractiveServiceClient(conn)
	local := client.NewInteractiveServiceAdapter(svc)
	res := client.NewGreyScaleInteractiveServiceClient(remote, local)
	// 监听配置文件变更，重新加载
	viper.OnConfigChange(func(in fsnotify.Event) {
		var cfg Config
		err = viper.UnmarshalKey("grpc.client.inter", &cfg)
		if err != nil {
			// 输出日志
		}
		res.UpdateThreshold(cfg.Threshold)
	})
	return res
}
