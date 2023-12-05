package dao

import "gorm.io/gorm"

type UserInfoDAO struct {
	db *gorm.DB
}

func NewUserInfoDAO(db *gorm.DB) *UserInfoDAO {
	return &UserInfoDAO{
		db: db,
	}
}
