package repository

import (
	"Prove/webook/internal/domain"
	"Prove/webook/internal/repository/cache"
	"Prove/webook/internal/repository/dao"
	"context"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound       = dao.ErrDataNotFound
)

type UserInfoRepository struct {
	dao   *dao.UserInfoDAO
	cache *cache.UserCache
}

func NewUserInfoRepository(dao *dao.UserInfoDAO, c *cache.UserCache) *UserInfoRepository {
	return &UserInfoRepository{
		dao:   dao,
		cache: c,
	}
}

func (ur *UserInfoRepository) Create(ctx context.Context, u *domain.User) error {
	return ur.dao.Insert(ctx, &dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
}

func (ur *UserInfoRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, err := ur.dao.FindByEmail(ctx, email)
	if err != nil {
		return &domain.User{}, err
	}
	return &domain.User{
		Id:         u.Id,
		Email:      u.Email,
		Password:   u.Password,
		Nickname:   u.Nickname,
		Birthday:   u.Birthday,
		CreateTime: u.CreateTime,
		UpdateTime: u.UpdateTime,
	}, nil
}

func (ur *UserInfoRepository) FindById(ctx context.Context, id int64) (*domain.User, error) {
	u, err := ur.cache.Get(ctx, id)
	// 存在数据
	if err == nil {
		return u, nil
	}
	// 不存在数据，考虑从数据库加载
	//if err == cache.ErrKeyNotExist {
	//
	//}
	ue, err := ur.dao.FindById(ctx, id)
	if err != nil {
		return &domain.User{}, err
	}
	u = &domain.User{
		Id:         ue.Id,
		Email:      ue.Email,
		Password:   ue.Password,
		Nickname:   ue.Nickname,
		Birthday:   ue.Birthday,
		CreateTime: ue.CreateTime,
		UpdateTime: ue.UpdateTime,
	}

	go func() {
		_ = ur.cache.Set(ctx, u)
		//if err != nil {
		//	// 考虑打日志，做监控
		//}
	}()
	return u, nil
	// redis崩溃，是否转移到数据库查找？可能会导致数据库崩溃
	// 1.加载，但需要为数据库兜底，考虑使用限流（由于redis集群已崩，考虑使用单机限流）
	// 2.不加载，会降低用户体验
}

func (ur *UserInfoRepository) CompleteInfo(ctx context.Context, u *domain.User) error {
	return ur.dao.CompleteInfo(ctx, &dao.User{
		Id:         u.Id,
		Nickname:   u.Nickname,
		Birthday:   u.Birthday,
		UpdateTime: u.UpdateTime,
	})
}
