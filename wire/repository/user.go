package repository

import (
	"Prove/wire/repository/dao"
)

type UserInfoRepository struct {
	dao *dao.UserInfoDAO
	//cache *cache.UserCache
}

func NewUserInfoRepository(dao *dao.UserInfoDAO) *UserInfoRepository {
	return &UserInfoRepository{
		dao: dao,
		//cache: cache,
	}
}
