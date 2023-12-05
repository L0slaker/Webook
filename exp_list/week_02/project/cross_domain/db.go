package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

func ConnectDB() *gorm.DB {
	dsn := ConfigDatabase()

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库链接失败:", err)
	}

	//添加外键约束
	//自动建表
	db.AutoMigrate(&User{}, &Article{}, &Session{})
	//db.Exec("ALTER TABLE articles ADD CONSTRAINT fk_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE")
	return db
}

func ConfigDatabase() string {
	username := "root"
	password := "AoNY12138"
	hostname := "127.0.0.1"
	port := "3306"
	databaseName := "login_register_db"

	dsn := username + ":" + password + "@tcp(" + hostname + ":" + port + ")/" + databaseName + "?charset=utf8mb4&parseTime=True&loc=Local"
	return dsn
}
