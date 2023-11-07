package main

import (
	"bytes"
	"fmt"
	"github.com/fsnotify/fsnotify"
	_ "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	//etcdctl --endpoints=127.0.0.1:12379 put /webook (Get-Content -Raw dev.yaml)
	//etcdctl --endpoints=127.0.0.1:12379 get /webook
	//initViperV3Remote()
	initViperV1()
	initLogger()

	keys := viper.AllKeys()
	settings := viper.AllSettings()
	fmt.Println("all keys: ", keys)
	fmt.Println("all settings: ", settings)

	app := InitWebServer()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}

	server := app.server
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "welcome!")
	})
	server.Run(":8080")
}

// 需要引入 	_ "github.com/spf13/viper/remote"
func initViperV3Remote() {
	// 通过 webook 和其他使用 etcd 的区别出来
	err := viper.AddRemoteProvider("etcd3",
		"http://127.0.0.1:12379", "/webook")
	if err != nil {
		panic(err)
	}
	viper.SetConfigType("yaml")

	// 监听配置文件变更
	err = viper.WatchRemoteConfig()
	if err != nil {
		panic(err)
	}

	err = viper.ReadRemoteConfig()
	if err != nil {
		panic(err)
	}
}

// initLogger 加载日志
func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	// 将自定义的日志记录器（logger）替换为全局的默认日志记录器
	zap.ReplaceGlobals(logger)
}

// initViperV1 加载配置
func initViperV1() {
	// 要在参数中指定 --config=config/dev.yaml
	cfile := pflag.String("config",
		"webook/config/dev.yaml", "配置文件路径")
	pflag.Parse()
	// 直接指定文件路径
	viper.SetConfigFile(*cfile)

	// 监听配置文件变更
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		// 有些设计不足，当修改配置文件时，只会提示被修改。而没有显示修改前后的变化
		// 比较好的设计，会在 in 里面告诉你变更前和变更后的数据
		// 更好的设计，会直接告诉你差异
		fmt.Println(in.Name, in.Op)
		fmt.Println(viper.GetString("db.dsn"))
	})
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func initViperV2() {
	// 读取的文件名为 dev
	viper.SetConfigName("dev")
	// 读取的类型为 yaml
	viper.SetConfigType("yaml")
	// 在当前目录的 config 子目录下
	viper.AddConfigPath("webook/config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func initViperV3() {
	// 直接指定文件路径
	viper.SetConfigFile("config/dev.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func initViperV4() {
	viper.SetConfigType("yaml")
	cfg := `
db.mysql:
  dsn: "root:root@tcp(localhost:13316)/webook"

redis:
  addr: "localhost:6379"
`
	err := viper.ReadConfig(bytes.NewReader([]byte(cfg)))
	if err != nil {
		panic(err)
	}
}
