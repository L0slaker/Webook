package service

import (
	"Prove/webook/internal/domain"
	"Prove/webook/internal/repository"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
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

func (svc *UserService) Signup(ctx context.Context, u *domain.User) error {
	hashPwd, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashPwd)
	return svc.r.Create(ctx, u)
}

func (svc *UserService) Login(ctx context.Context, email, password string) (*domain.User, error) {
	u, err := svc.r.FindByEmail(ctx, email)
	if err == repository.ErrUserNotFound {
		return &domain.User{}, ErrUserDuplicateEmail
	}
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return &domain.User{}, ErrInvalidUserOrPassword
	}
	return u, err
}

func (svc *UserService) Edit(ctx context.Context, u *domain.User) error {
	return svc.r.CompleteInfo(ctx, u)
}

func (svc *UserService) FindOrCreate(ctx *gin.Context, phone string) (*domain.User, error) {
	u, err := svc.r.FindByPhone(ctx, phone)
	// 判断用户是否存在,存在即返回
	if err != repository.ErrUserNotFound {
		return u, err
	}
	// 不存在，需要创建
	err = svc.r.Create(ctx, &domain.User{
		Phone: phone,
	})
	if err != nil {
		return u, err
	}
	// 创建后再找一次 ID，但是可能会遇到主从延迟的问题
	return svc.r.FindByPhone(ctx, phone)
}

func (svc *UserService) Profile(ctx context.Context, id int64) (*domain.User, error) {
	return svc.r.FindById(ctx, id)
}
