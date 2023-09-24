package repository

import (
	"Prove/webook/internal/domain"
	"Prove/webook/internal/repository/cache"
	"Prove/webook/internal/repository/dao"
	"context"
	"database/sql"
	"time"
)

var (
	ErrUserDuplicate = dao.ErrUserDuplicate
	ErrUserNotFound  = dao.ErrDataNotFound
)

type UserRepository interface {
	Create(ctx context.Context, u *domain.User) error
	FindById(ctx context.Context, id int64) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByPhone(ctx context.Context, phone string) (*domain.User, error)
	CompleteInfo(ctx context.Context, u *domain.User) error
	FindByWechat(ctx context.Context, openId string) (*domain.User, error)
}

type UserInfoRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewUserInfoRepository(dao dao.UserDAO, c cache.UserCache) UserRepository {
	return &UserInfoRepository{
		dao:   dao,
		cache: c,
	}
}

func (ur *UserInfoRepository) Create(ctx context.Context, u *domain.User) error {
	ue := ur.domainToEntity(u)
	return ur.dao.Insert(ctx, ue)
}

func (ur *UserInfoRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	ue, err := ur.dao.FindByEmail(ctx, email)
	if err != nil {
		return &domain.User{}, err
	}
	u := ur.entityToDomain(ue)
	return u, nil
}

func (ur *UserInfoRepository) FindByPhone(ctx context.Context, phone string) (*domain.User, error) {
	ue, err := ur.dao.FindByPhone(ctx, phone)
	if err != nil {
		return &domain.User{}, err
	}
	u := ur.entityToDomain(ue)
	return u, nil
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
	u = ur.entityToDomain(ue)

	_ = ur.cache.Set(ctx, u)
	//if err != nil {
	//	// 考虑打日志，做监控
	//}
	//go func() {
	//
	//}()
	return u, nil
	// redis崩溃，是否转移到数据库查找？可能会导致数据库崩溃
	// 1.加载，但需要为数据库兜底，考虑使用限流（由于redis集群已崩，考虑使用单机限流）
	// 2.不加载，会降低用户体验
}

func (ur *UserInfoRepository) FindByWechat(ctx context.Context, openId string) (*domain.User, error) {
	ue, err := ur.dao.FindByWechat(ctx, openId)
	if err != nil {
		return &domain.User{}, err
	}
	u := ur.entityToDomain(ue)
	return u, nil
}

func (ur *UserInfoRepository) CompleteInfo(ctx context.Context, u *domain.User) error {
	return ur.dao.CompleteInfo(ctx, dao.User{
		Id:         u.Id,
		Nickname:   u.Nickname,
		Birthday:   u.Birthday,
		UpdateTime: u.Utime.UnixMilli(),
	})
}

func (ur *UserInfoRepository) entityToDomain(u dao.User) *domain.User {
	return &domain.User{
		Id:    u.Id,
		Email: u.Email.String,
		Phone: u.Phone.String,
		WechatInfo: domain.WechatInfo{
			UnionId: u.WechatUnionId.String,
			OpenId:  u.WechatOpenId.String,
		},
		Password: u.Password,
		Nickname: u.Nickname,
		Birthday: u.Birthday,
		Ctime:    time.UnixMilli(u.CreateTime),
		Utime:    time.UnixMilli(u.UpdateTime),
	}
}

func (ur *UserInfoRepository) domainToEntity(u *domain.User) dao.User {
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			// 确实有手机号
			Valid: u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			// 确实有邮箱
			Valid: u.Phone != "",
		},
		WechatOpenId: sql.NullString{
			String: u.WechatInfo.OpenId,
			Valid:  u.WechatInfo.OpenId != "",
		},
		WechatUnionId: sql.NullString{
			String: u.WechatInfo.UnionId,
			Valid:  u.WechatInfo.UnionId != "",
		},
		Password:   u.Password,
		Nickname:   u.Nickname,
		Birthday:   u.Birthday,
		CreateTime: u.Ctime.UnixMilli(),
		UpdateTime: u.Utime.UnixMilli(),
	}
}
