package ioc

import (
	"Prove/webook/internal/repository/dao"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
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
	db, err := gorm.Open(mysql.Open(cfg.DSN))
	if err != nil {
		panic(err)
	}
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}
