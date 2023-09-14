package service

import (
	"Prove/webook/internal/domain"
	"Prove/webook/internal/repository"
	repomocks "Prove/webook/internal/repository/mocks"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"
)

func TestUserService_Signup(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(*gomock.Controller) repository.UserRepository
		user    *domain.User
		wantErr error
	}{
		{
			name: "注册成功！",
			user: &domain.User{
				Id:       123,
				Email:    "l0slakers@gmail.com",
				Password: "Abcd#1234",
			},
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				userRepo := repomocks.NewMockUserRepository(ctrl)
				userRepo.EXPECT().Create(gomock.Any(), gomock.Any()).
					Return(nil)
				return userRepo
			},
			wantErr: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewUserService(tc.mock(ctrl))
			err := svc.Signup(context.Background(), tc.user)
			fmt.Println("actual ERROR:", err)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestUserService_Login(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name     string
		mock     func(*gomock.Controller) repository.UserRepository
		email    string
		password string
		wantUser *domain.User
		wantErr  error
	}{
		{
			name:     "找不到该用户！",
			email:    "l0slakers@gmail.com",
			password: "Abcd#1234",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				userRepo := repomocks.NewMockUserRepository(ctrl)
				userRepo.EXPECT().FindByEmail(gomock.Any(), "l0slakers@gmail.com").
					Return(&domain.User{}, repository.ErrUserNotFound)
				return userRepo
			},
			wantUser: &domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
		{
			name:     "DB 错误！",
			email:    "l0slakers@gmail.com",
			password: "Abcd#1234",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				userRepo := repomocks.NewMockUserRepository(ctrl)
				userRepo.EXPECT().FindByEmail(gomock.Any(), "l0slakers@gmail.com").
					Return(&domain.User{}, errors.New("mock db 错误！"))
				return userRepo
			},
			wantUser: &domain.User{},
			wantErr:  errors.New("mock db 错误！"),
		},
		{
			name:  "密码不正确！",
			email: "l0slakers@gmail.com",
			// 这里的密码已经和加密的密码不同
			password: "Abcd#12347788",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				userRepo := repomocks.NewMockUserRepository(ctrl)
				userRepo.EXPECT().FindByEmail(gomock.Any(), "l0slakers@gmail.com").
					Return(&domain.User{
						Email:    "l0slakers@gmail.com",
						Password: "$2a$10$K0T3cJ5hAbFIAhJiRcd1durGTO7/E5pn7nYPmk6f9bTkixxMMtEmm",
						Phone:    "13355779876",
						Ctime:    now.UnixMilli(),
						Utime:    now.UnixMilli(),
					}, nil)
				return userRepo
			},
			wantUser: &domain.User{},
			wantErr:  ErrInvalidUserOrPassword,
		},
		{
			name:     "登陆成功！",
			email:    "l0slakers@gmail.com",
			password: "Abcd#1234",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				userRepo := repomocks.NewMockUserRepository(ctrl)
				userRepo.EXPECT().FindByEmail(gomock.Any(), "l0slakers@gmail.com").
					Return(&domain.User{
						Email:    "l0slakers@gmail.com",
						Password: "$2a$10$K0T3cJ5hAbFIAhJiRcd1durGTO7/E5pn7nYPmk6f9bTkixxMMtEmm",
						Phone:    "13355779876",
						Ctime:    now.UnixMilli(),
						Utime:    now.UnixMilli(),
					}, nil)
				return userRepo
			},
			wantUser: &domain.User{
				Email:    "l0slakers@gmail.com",
				Password: "$2a$10$K0T3cJ5hAbFIAhJiRcd1durGTO7/E5pn7nYPmk6f9bTkixxMMtEmm",
				Phone:    "13355779876",
				Ctime:    now.UnixMilli(),
				Utime:    now.UnixMilli(),
			},
			wantErr: nil,
		},
	}
	gin.SetMode(gin.ReleaseMode)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewUserService(tc.mock(ctrl))
			user, err := svc.Login(context.Background(), tc.email, tc.password)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}

func TestUserService_Edit(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(*gomock.Controller) repository.UserRepository
		user    *domain.User
		wantErr error
	}{
		{
			name: "更新成功！",
			user: &domain.User{
				Id:       123,
				Nickname: "l0slakers",
				Birthday: "2000-12-14",
			},
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				userRepo := repomocks.NewMockUserRepository(ctrl)
				userRepo.EXPECT().CompleteInfo(gomock.Any(), gomock.Any()).
					Return(nil)
				return userRepo
			},
			wantErr: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewUserService(tc.mock(ctrl))
			err := svc.Edit(context.Background(), tc.user)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestUserService_Profile(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name     string
		mock     func(*gomock.Controller) repository.UserRepository
		id       int64
		wantUser *domain.User
		wantErr  error
	}{
		{
			name: "查看个人信息",
			id:   123,
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				userRepo := repomocks.NewMockUserRepository(ctrl)
				userRepo.EXPECT().FindById(gomock.Any(), gomock.Any()).
					Return(&domain.User{
						Id:       123,
						Email:    "l0slakers@gmail.com",
						Phone:    "13378899456",
						Password: "$2a$10$K0T3cJ5hAbFIAhJiRcd1durGTO7/E5pn7nYPmk6f9bTkixxMMtEmm",
						Nickname: "l0slakers",
						Birthday: "2000-12-14",
						Ctime:    now.UnixMilli(),
						Utime:    now.UnixMilli(),
					}, nil)
				return userRepo
			},
			wantUser: &domain.User{
				Id:       123,
				Email:    "l0slakers@gmail.com",
				Phone:    "13378899456",
				Password: "$2a$10$K0T3cJ5hAbFIAhJiRcd1durGTO7/E5pn7nYPmk6f9bTkixxMMtEmm",
				Nickname: "l0slakers",
				Birthday: "2000-12-14",
				Ctime:    now.UnixMilli(),
				Utime:    now.UnixMilli(),
			},
			wantErr: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewUserService(tc.mock(ctrl))
			user, err := svc.Profile(context.Background(), tc.id)
			assert.Equal(t, tc.wantUser, user)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestUserService_FindOrCreate(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name     string
		mock     func(controller *gomock.Controller) repository.UserRepository
		phone    string
		wantUser *domain.User
		wantErr  error
	}{
		{
			name:  "用户存在",
			phone: "13355779876",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				userRepo := repomocks.NewMockUserRepository(ctrl)
				userRepo.EXPECT().FindByPhone(gomock.Any(), "13355779876").
					Return(&domain.User{
						Id:    123,
						Phone: "13355779876",
						Ctime: now.UnixMilli(),
						Utime: now.UnixMilli(),
					}, nil)
				return userRepo
			},
			wantUser: &domain.User{
				Id:    123,
				Phone: "13355779876",
				Ctime: now.UnixMilli(),
				Utime: now.UnixMilli(),
			},
			wantErr: nil,
		},
		{
			name:  "用户不存在，创建成功",
			phone: "13355779876",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				userRepo := repomocks.NewMockUserRepository(ctrl)
				userRepo.EXPECT().FindByPhone(gomock.Any(), "13355779876").
					Return(nil, repository.ErrUserNotFound)
				userRepo.EXPECT().Create(gomock.Any(), gomock.Any()).
					Return(nil)
				userRepo.EXPECT().FindByPhone(gomock.Any(), "13355779876").
					Return(&domain.User{
						Id:    456,
						Phone: "13355779876",
						Ctime: now.UnixMilli(),
						Utime: now.UnixMilli(),
					}, nil)
				return userRepo
			},
			wantUser: &domain.User{
				Id:    456,
				Phone: "13355779876",
				Ctime: now.UnixMilli(),
				Utime: now.UnixMilli(),
			},
			wantErr: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewUserService(tc.mock(ctrl))
			user, err := svc.FindOrCreate(context.Background(), tc.phone)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}

func TestEncrypt(t *testing.T) {
	password := "Abcd#1234"
	encrypted, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err == nil {
		t.Log(string(encrypted))
	}
	//err = bcrypt.CompareHashAndPassword(encrypted, []byte(password))
	//assert.NoError(t, err)
}
