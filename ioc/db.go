package ioc

import (
	"Prove/webook/internal/repository/dao"
	"Prove/webook/pkg/gormx"
	"Prove/webook/pkg/logger"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
	"gorm.io/plugin/prometheus"
	"time"
)

func InitDB(l logger.LoggerV1) *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	// 设置默认值
	var cfg = Config{
		DSN: "root:root@tcp(localhost:13316)/webook_default",
	}
	err := viper.UnmarshalKey("db", &cfg)
	if err != nil {
		panic(err)
	}
	//dsn := viper.GetString("db.dsn")
	//db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: glogger.New(gormLoggerFunc(l.Debug), glogger.Config{
			// 慢查询阈值，只有执行时间超过该阈值，才会使用（50ms，100ms）
			// SQL 查询必然要求命中索引，最好就是走一次磁盘 IO，一次磁盘 IO 是不到 10ms
			SlowThreshold: time.Millisecond * 10,
			// 忽略记录未找到错误
			IgnoreRecordNotFoundError: true,
			// 确定数据库查询是否应该是参数化的,也就是用占位符代替了数据
			ParameterizedQueries: true,
			LogLevel:             glogger.Info,
		}),
	})
	if err != nil {
		panic(err)
	}

	err = db.Use(prometheus.New(prometheus.Config{
		DBName: "webook",
		// 每 15s 采集一次数据
		RefreshInterval: 15,
		// 监控数据库中正在运行的线程数
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.MySQL{
				VariableNames: []string{"Threads_running"},
			},
		},
	}))
	if err != nil {
		panic(err)
	}

	// 监控查询的执行时间
	cbs := gormx.NewCallbacks("geekbang_l0slakers",
		"webook", "gorm_query_time",
		"统计 GORM 的执行时间", map[string]string{"db": "webook"})
	cbs.RegisterAll(db)

	db.Use(tracing.NewPlugin(tracing.WithDBName("webook"),
		tracing.WithQueryFormatter(func(query string) string {
			l.Debug("", logger.String("query", query))
			return query
		})))

	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(msg string, args ...interface{}) {
	g(msg, logger.Field{Key: "args", Value: args})
}
