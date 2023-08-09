package service

import (
	"Prove/webook/internal/domain"
	"Prove/webook/internal/repository"
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserDuplicateEmail    = repository.ErrUserDuplicateEmail
	ErrInvalidUserOrPassword = errors.New("邮箱或密码不正确")
)

type UserService struct {
	r *repository.UserInfoRepository
}

func NewUserService(r *repository.UserInfoRepository) *UserService {
	return &UserService{
		r: r,
	}
}

func (svc *UserService) Signup(ctx context.Context, u domain.User) error {
	hashPwd, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashPwd)
	return svc.r.Create(ctx, u)
}

func (svc *UserService) Login(ctx context.Context, email, password string) (domain.User, error) {
	u, err := svc.r.FindByEmail(ctx, email)
	if err == repository.ErrUserNotFound {
		return domain.User{}, ErrUserDuplicateEmail
	}
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, err
}

func (svc *UserService) Edit(ctx context.Context, u domain.User) (domain.User, error) {
	user, err := svc.r.FindById(ctx, u.Id)
	if err == repository.ErrUserNotFound {
		return domain.User{}, err
	}
	if err = svc.r.CompleteInfo(ctx, user); err != nil {
		return domain.User{}, err
	}
	return user, err
}

func (svc *UserService) Profile(ctx context.Context, id int64) (domain.User, error) {
	return svc.r.FindById(ctx, id)
}
