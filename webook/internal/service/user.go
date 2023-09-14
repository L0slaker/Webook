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
	ErrUserDuplicate         = repository.ErrUserDuplicate
	ErrInvalidUserOrPassword = errors.New("邮箱或密码不正确")
)

type UserAndService interface {
	Signup(ctx context.Context, u *domain.User) error
	Login(ctx context.Context, email, password string) (*domain.User, error)
	Edit(ctx context.Context, u *domain.User) error
	FindOrCreate(ctx context.Context, phone string) (*domain.User, error)
	FindOrCreateByWechat(ctx *gin.Context, wechatInfo domain.WechatInfo) (*domain.User, error)
	Profile(ctx context.Context, id int64) (*domain.User, error)
}

type UserService struct {
	r repository.UserRepository
}

func NewUserService(r repository.UserRepository) UserAndService {
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
		return &domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return &domain.User{}, err
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

func (svc *UserService) FindOrCreate(ctx context.Context, phone string) (*domain.User, error) {
	u, err := svc.r.FindByPhone(ctx, phone)
	// 判断用户是否存在,存在即返回
	if err != repository.ErrUserNotFound {
		return u, err
	}

	// 1.理论上可以直接创建用户，如果不成功就说明已经注册过了。但是需要考虑数据库的承受能力，
	// 如果有大量的创建请求直接打进数据库，很可能导致数据库崩溃
	// 2.在系统资源不足时，出发降级之后，就不执行慢路径了
	// 数据库创建就是慢路径，资源不足时，就不允许用户注册了
	//if ctx.Value("降级") == true{
	//	return &domain.User{},errors.New("系统降级了")
	//}

	// 不存在，需要创建
	err = svc.r.Create(ctx, &domain.User{
		Phone: phone,
	})
	if err != nil && err != repository.ErrUserDuplicate {
		return u, err
	}
	// 创建后再找一次 ID，但是可能会遇到主从延迟的问题
	return svc.r.FindByPhone(ctx, phone)
}

func (svc *UserService) Profile(ctx context.Context, id int64) (*domain.User, error) {
	return svc.r.FindById(ctx, id)
}

func (svc *UserService) FindOrCreateByWechat(ctx *gin.Context, wechatInfo domain.WechatInfo) (*domain.User, error) {
	u, err := svc.r.FindByWechat(ctx, wechatInfo.OpenId)
	// 判断用户是否存在,存在即返回
	if err != repository.ErrUserNotFound {
		return u, err
	}

	user := &domain.User{
		WechatInfo: wechatInfo,
	}
	// 1.理论上可以直接创建用户，如果不成功就说明已经注册过了。但是需要考虑数据库的承受能力，
	// 如果有大量的创建请求直接打进数据库，很可能导致数据库崩溃
	// 2.在系统资源不足时，出发降级之后，就不执行慢路径了
	// 数据库创建就是慢路径，资源不足时，就不允许用户注册了
	//if ctx.Value("降级") == true{
	//	return &domain.User{},errors.New("系统降级了")
	//}

	// 不存在，需要创建
	err = svc.r.Create(ctx, user)
	if err != nil && err != repository.ErrUserDuplicate {
		return u, err
	}
	// 创建后再找一次 ID，但是可能会遇到主从延迟的问题
	return svc.r.FindByWechat(ctx, wechatInfo.OpenId)
}
