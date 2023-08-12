package repository

import (
	"Prove/webook/internal/domain"
	"Prove/webook/internal/repository/dao"
	"context"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound       = dao.ErrDataNotFound
)

type UserInfoRepository struct {
	dao *dao.UserInfoDAO
}

func NewUserInfoRepository(dao *dao.UserInfoDAO) *UserInfoRepository {
	return &UserInfoRepository{
		dao: dao,
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
	return &domain.User{
		Id:         u.Id,
		Email:      u.Email,
		Password:   u.Password,
		Nickname:   u.Nickname,
		Birthday:   u.Birthday,
		CreateTime: u.CreateTime,
		UpdateTime: u.UpdateTime,
	}, err
}

func (ur *UserInfoRepository) FindById(ctx context.Context, id int64) (*domain.User, error) {
	u, err := ur.dao.FindById(ctx, id)
	return &domain.User{
		Id:         u.Id,
		Email:      u.Email,
		Password:   u.Password,
		Nickname:   u.Nickname,
		Birthday:   u.Birthday,
		CreateTime: u.CreateTime,
		UpdateTime: u.UpdateTime,
	}, err
}

func (ur *UserInfoRepository) CompleteInfo(ctx context.Context, u *domain.User) error {
	return ur.dao.CompleteInfo(ctx, &dao.User{
		Id:         u.Id,
		Nickname:   u.Nickname,
		Birthday:   u.Birthday,
		UpdateTime: u.UpdateTime,
	})
}
