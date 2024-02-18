package ioc

import (
	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"

	"Prove/webook/interactive/repository/dao"
	"Prove/webook/pkg/ginx"
	"Prove/webook/pkg/gormx/connpool"
	"Prove/webook/pkg/logger"
	"Prove/webook/pkg/migrator/events"
	"Prove/webook/pkg/migrator/events/fixer"
	"Prove/webook/pkg/migrator/scheduler"
)

const (
	InteractiveTopic       = "migrator_interactives"
	UserLikeBizTopic       = "migrator_user_like_biz"
	UserCollectionBizTopic = "migrator_user_collection_biz"
	CollectionTopic        = "migrator_collection"
)

func InitFixDataConsumer(l logger.LoggerV1, src SrcDB, dst DstDB, client sarama.Client) *fixer.Consumer[dao.Interactive] {
	res, err := fixer.NewConsumer[dao.Interactive](client, l, src, dst, InteractiveTopic)
	if err != nil {
		panic(err)
	}
	return res
}

func InitMigradatorProducer(p sarama.SyncProducer) events.Producer {
	return events.NewSaramaProducer(p, InteractiveTopic)
}

// InitMigratorWeb 初始化，有多少张表就初始化多少个 scheduler
func InitMigratorWeb(l logger.LoggerV1, src SrcDB, dst DstDB,
	pool *connpool.DoubleWritePool, producer events.Producer) *ginx.Server {
	engine := gin.Default()
	ginx.InitCounter(prometheus.CounterOpts{
		Namespace: "geekbang_l0slakers",
		Subsystem: "webook_inter",
		Name:      "http_biz_code",
		Help:      "GIN 中 的 HTTP 请求",
		ConstLabels: map[string]string{
			"instance_id": "my-instance-1",
		},
	})
	interSch := scheduler.NewScheduler[dao.Interactive](l, src, dst, pool, producer)
	interSch.RegisterRoutes(engine.Group("/migrator/interactive"))

	//likeBizSch := scheduler.NewScheduler[dao.UserLikeBiz](l, src, dst, pool, producer)
	//collectBizSch := scheduler.NewScheduler[dao.UserCollectionBiz](l, src, dst, pool, producer)
	//collectSch := scheduler.NewScheduler[dao.Collection](l, src, dst, pool, producer)
	//likeBizSch.RegisterRoutes(engine.Group("/migrator/user_like_biz"))
	//collectBizSch.RegisterRoutes(engine.Group("/migrator/user_collect_biz"))
	//collectSch.RegisterRoutes(engine.Group("/migrator/collection"))

	addr := viper.GetString("migrator.web.addr")
	return &ginx.Server{
		Addr:   addr,
		Engine: engine,
	}
}

func InitDoubleWritePool(src SrcDB, dst DstDB, l logger.LoggerV1) *connpool.DoubleWritePool {
	return connpool.NewDoubleWritePool(src, dst, l)
}
