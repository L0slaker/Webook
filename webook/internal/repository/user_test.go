package repository

import (
	"Prove/webook/internal/domain"
	"Prove/webook/internal/repository/cache"
	cachemocks "Prove/webook/internal/repository/cache/mocks"
	"Prove/webook/internal/repository/dao"
	daomocks "Prove/webook/internal/repository/dao/mocks"
	"context"
	"database/sql"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestUserInfoRepository_FindById(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name     string
		mock     func(*gomock.Controller) (dao.UserDAO, cache.UserCache)
		ctx      context.Context
		id       int64
		wantUser *domain.User
		wantErr  error
	}{
		{
			name: "缓存命中，查询成功！",
			ctx:  context.Background(),
			id:   123,
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				userDao := daomocks.NewMockUserDAO(ctrl)
				userCache := cachemocks.NewMockUserCache(ctrl)
				userCache.EXPECT().Get(gomock.Any(), int64(123)).
					Return(&domain.User{
						Id:       123,
						Email:    "l0slakers@gmail.com",
						Phone:    "13399887069",
						Password: "$2a$10$K0T3cJ5hAbFIAhJiRcd1durGTO7/E5pn7nYPmk6f9bTkixxMMtEmm",
						Nickname: "l0slakers",
						Birthday: "2000-12-14",
						Ctime:    now.UnixMilli(),
						Utime:    now.UnixMilli(),
					}, nil)
				return userDao, userCache
			},
			wantUser: &domain.User{
				Id:       123,
				Email:    "l0slakers@gmail.com",
				Phone:    "13399887069",
				Password: "$2a$10$K0T3cJ5hAbFIAhJiRcd1durGTO7/E5pn7nYPmk6f9bTkixxMMtEmm",
				Nickname: "l0slakers",
				Birthday: "2000-12-14",
				Ctime:    now.UnixMilli(),
				Utime:    now.UnixMilli(),
			},
			wantErr: nil,
		},
		{
			name: "缓存未命中，查询也失败！",
			ctx:  context.Background(),
			id:   123,
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				userDao := daomocks.NewMockUserDAO(ctrl)
				userCache := cachemocks.NewMockUserCache(ctrl)
				userCache.EXPECT().Get(gomock.Any(), int64(123)).
					Return(&domain.User{}, cache.ErrKeyNotExist)
				userDao.EXPECT().FindById(gomock.Any(), int64(123)).
					Return(dao.User{}, ErrUserNotFound)
				return userDao, userCache
			},
			wantUser: &domain.User{},
			wantErr:  ErrUserNotFound,
		},
		{
			name: "缓存未命中，查询成功！",
			ctx:  context.Background(),
			id:   123,
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				userDao := daomocks.NewMockUserDAO(ctrl)
				userCache := cachemocks.NewMockUserCache(ctrl)
				userCache.EXPECT().Get(gomock.Any(), int64(123)).
					Return(&domain.User{}, cache.ErrKeyNotExist)
				userDao.EXPECT().FindById(gomock.Any(), int64(123)).
					Return(dao.User{
						Id: 123,
						Email: sql.NullString{
							String: "l0slakers@gmail.com",
							Valid:  true,
						},
						Phone: sql.NullString{
							String: "13399887069",
							Valid:  true,
						},
						Password:   "$2a$10$K0T3cJ5hAbFIAhJiRcd1durGTO7/E5pn7nYPmk6f9bTkixxMMtEmm",
						Nickname:   "l0slakers",
						Birthday:   "2000-12-14",
						CreateTime: now.UnixMilli(),
						UpdateTime: now.UnixMilli(),
					}, nil)
				userCache.EXPECT().Set(gomock.Any(), &domain.User{
					Id:       123,
					Email:    "l0slakers@gmail.com",
					Phone:    "13399887069",
					Password: "$2a$10$K0T3cJ5hAbFIAhJiRcd1durGTO7/E5pn7nYPmk6f9bTkixxMMtEmm",
					Nickname: "l0slakers",
					Birthday: "2000-12-14",
					Ctime:    now.UnixMilli(),
					Utime:    now.UnixMilli(),
				}).Return(nil)
				return userDao, userCache
			},
			wantUser: &domain.User{
				Id:       123,
				Email:    "l0slakers@gmail.com",
				Phone:    "13399887069",
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

			repo := NewUserInfoRepository(tc.mock(ctrl))
			user, err := repo.FindById(tc.ctx, tc.id)
			assert.Equal(t, tc.wantUser, user)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestUserInfoRepository_FindByEmail(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name     string
		ctx      context.Context
		email    string
		mock     func(*gomock.Controller) (dao.UserDAO, cache.UserCache)
		wantUser *domain.User
		wantErr  error
	}{
		{
			name:  "查询失败！",
			ctx:   context.Background(),
			email: "l0slakers@gmail.com",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				c := cachemocks.NewMockUserCache(ctrl)
				d := daomocks.NewMockUserDAO(ctrl)
				d.EXPECT().FindByEmail(gomock.Any(), "l0slakers@gmail.com").
					Return(dao.User{}, ErrUserNotFound)
				return d, c
			},
			wantUser: &domain.User{},
			wantErr:  ErrUserNotFound,
		},
		{
			name:  "查询成功！",
			ctx:   context.Background(),
			email: "l0slakers@gmail.com",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				c := cachemocks.NewMockUserCache(ctrl)
				d := daomocks.NewMockUserDAO(ctrl)
				d.EXPECT().FindByEmail(gomock.Any(), "l0slakers@gmail.com").
					Return(dao.User{
						Id: 123,
						Email: sql.NullString{
							String: "l0slakers@gmail.com",
							Valid:  true,
						},
						Phone: sql.NullString{
							String: "13399887069",
							Valid:  true,
						},
						Password:   "$2a$10$K0T3cJ5hAbFIAhJiRcd1durGTO7/E5pn7nYPmk6f9bTkixxMMtEmm",
						Nickname:   "l0slakers",
						Birthday:   "2000-12-14",
						CreateTime: now.UnixMilli(),
						UpdateTime: now.UnixMilli(),
					}, nil)
				return d, c
			},
			wantUser: &domain.User{
				Id:       123,
				Email:    "l0slakers@gmail.com",
				Phone:    "13399887069",
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

			repo := NewUserInfoRepository(tc.mock(ctrl))
			user, err := repo.FindByEmail(tc.ctx, tc.email)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}

func TestUserInfoRepository_FindByPhone(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name     string
		mock     func(*gomock.Controller) (dao.UserDAO, cache.UserCache)
		ctx      context.Context
		phone    string
		wantUser *domain.User
		wantErr  error
	}{
		{
			name:  "查询失败！",
			ctx:   context.Background(),
			phone: "13377839761",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				c := cachemocks.NewMockUserCache(ctrl)
				d := daomocks.NewMockUserDAO(ctrl)
				d.EXPECT().FindByPhone(gomock.Any(), "13377839761").
					Return(dao.User{}, ErrUserNotFound)
				return d, c
			},
			wantUser: &domain.User{},
			wantErr:  ErrUserNotFound,
		},
		{
			name:  "查询成功！",
			ctx:   context.Background(),
			phone: "13377839761",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				c := cachemocks.NewMockUserCache(ctrl)
				d := daomocks.NewMockUserDAO(ctrl)
				d.EXPECT().FindByPhone(gomock.Any(), "13377839761").
					Return(dao.User{
						Id: 123,
						Email: sql.NullString{
							String: "l0slakers@gmail.com",
							Valid:  true,
						},
						Phone: sql.NullString{
							String: "13399887069",
							Valid:  true,
						},
						Password:   "$2a$10$K0T3cJ5hAbFIAhJiRcd1durGTO7/E5pn7nYPmk6f9bTkixxMMtEmm",
						Nickname:   "l0slakers",
						Birthday:   "2000-12-14",
						CreateTime: now.UnixMilli(),
						UpdateTime: now.UnixMilli(),
					}, nil)
				return d, c
			},
			wantUser: &domain.User{
				Id:       123,
				Email:    "l0slakers@gmail.com",
				Phone:    "13399887069",
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

			repo := NewUserInfoRepository(tc.mock(ctrl))
			user, err := repo.FindByPhone(tc.ctx, tc.phone)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}

func TestUserInfoRepository_Create(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name    string
		mock    func(*gomock.Controller) (dao.UserDAO, cache.UserCache)
		ctx     context.Context
		user    *domain.User
		wantErr error
	}{
		{
			name: "创建成功！",
			ctx:  context.Background(),
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				d := daomocks.NewMockUserDAO(ctrl)
				c := cachemocks.NewMockUserCache(ctrl)
				d.EXPECT().Insert(gomock.Any(), dao.User{
					Id: 123,
					Email: sql.NullString{
						String: "l0slakers@gmail.com",
						Valid:  true,
					},
					Phone: sql.NullString{
						String: "13399887069",
						Valid:  true,
					},
					Password:   "$2a$10$K0T3cJ5hAbFIAhJiRcd1durGTO7/E5pn7nYPmk6f9bTkixxMMtEmm",
					Nickname:   "l0slakers",
					Birthday:   "2000-12-14",
					CreateTime: now.UnixMilli(),
					UpdateTime: now.UnixMilli(),
				}).Return(nil)
				return d, c
			},
			user: &domain.User{
				Id:       123,
				Email:    "l0slakers@gmail.com",
				Phone:    "13399887069",
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

			repo := NewUserInfoRepository(tc.mock(ctrl))
			err := repo.Create(tc.ctx, tc.user)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestUserInfoRepository_CompleteInfo(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name    string
		mock    func(*gomock.Controller) (dao.UserDAO, cache.UserCache)
		ctx     context.Context
		user    *domain.User
		wantErr error
	}{
		{
			name: "创建成功！",
			ctx:  context.Background(),
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				d := daomocks.NewMockUserDAO(ctrl)
				c := cachemocks.NewMockUserCache(ctrl)
				d.EXPECT().CompleteInfo(gomock.Any(), dao.User{
					Id:         123,
					Nickname:   "l0slakers",
					Birthday:   "2000-12-14",
					UpdateTime: now.UnixMilli(),
				}).Return(nil)
				return d, c
			},
			user: &domain.User{
				Id:       123,
				Nickname: "l0slakers",
				Birthday: "2000-12-14",
				Utime:    now.UnixMilli(),
			},
			wantErr: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := NewUserInfoRepository(tc.mock(ctrl))
			err := repo.CompleteInfo(tc.ctx, tc.user)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
