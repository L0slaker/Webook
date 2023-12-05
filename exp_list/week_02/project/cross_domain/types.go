package main

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	gorm.Model
	Username string    `gorm:"unique" json:"username"`
	Email    string    `gorm:"unique" json:"email"`
	Password string    `json:"password"`
	Articles []Article `json:"articles"`
}

type Article struct {
	gorm.Model
	Title   string `json:"title"`
	Content string `json:"content"`
	UserId  uint   `json:"user_id"` // UserId 为外键，与 User 建立一对多关系
}

type Session struct {
	SessionID string    `json:"session_id"`
	UserId    uint      `json:"user_id"`
	ExpiredAt time.Time // 过期时间
}
