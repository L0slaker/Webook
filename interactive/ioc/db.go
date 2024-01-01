package ioc

import (
	"Prove/webook/interactive/repository/dao"
	"Prove/webook/pkg/logger"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		//Logger: glogger.New(gormLoggerFunc(l.Debug), glogger.Config{
		//	SlowThreshold:             time.Millisecond * 100,
		//	IgnoreRecordNotFoundError: true,
		//	ParameterizedQueries:      true,
		//	LogLevel:                  glogger.Info,
		//}),
	})
	if err != nil {
		panic(err)
	}

	//err = db.Use(prometheus.New(prometheus.Config{
	//	DBName:          "webook",
	//	RefreshInterval: 15,
	//	MetricsCollector: []prometheus.MetricsCollector{
	//		&prometheus.MySQL{
	//			VariableNames: []string{"Threads_running"},
	//		},
	//	},
	//}))
	//if err != nil {
	//	panic(err)
	//}

	// 监控查询的执行时间
	//cbs := gormx.NewCallbacks("geekbang_l0slakers",
	//	"webook", "gorm_query_time",
	//	"统计 GORM 的执行时间", map[string]string{"db": "webook"})
	//cbs.RegisterAll(db)

	//db.Use(tracing.NewPlugin(tracing.WithDBName("webook"),
	//	tracing.WithQueryFormatter(func(query string) string {
	//		l.Debug("", logger.String("query", query))
	//		return query
	//	})))

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
