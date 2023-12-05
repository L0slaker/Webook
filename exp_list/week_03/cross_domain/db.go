package main

import (
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

func setupMysqlClient() *gorm.DB {
	dsn := mysqlConfig()

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库链接失败:", err)
	}

	//添加外键约束
	//自动建表
	db.AutoMigrate(&User{}, &Article{})
	//db.Exec("ALTER TABLE articles ADD CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE")
	return db
}

func mysqlConfig() string {
	username := "root"
	password := "AoNY12138"
	hostname := "127.0.0.1"
	port := "3306"
	databaseName := "login_register_db"

	dsn := username + ":" + password + "@tcp(" + hostname + ":" + port + ")/" + databaseName + "?charset=utf8mb4&parseTime=True&loc=Local"
	return dsn
}

func setupRedisClient() *redis.Client {
	// 1.创建客户端连接
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", //服务器地址和端口
		Password: "",               //密码
		DB:       0,                //数据库索引
	})

	return client
}
