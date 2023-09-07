package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestGormUserDAO_Insert(t *testing.T) {
	testCases := []struct {
		name    string
		ctx     context.Context
		user    User
		mock    func(t *testing.T) *sql.DB
		wantErr error
	}{
		{
			name: "邮箱冲突！",
			ctx:  context.Background(),
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectExec("INSERT INTO `users` .*").
					WillReturnError(&mysql.MySQLError{
						Number: 1062,
					})
				return mockDB
			},
			user:    User{},
			wantErr: ErrUserDuplicate,
		},
		{
			name: "数据库错误！",
			ctx:  context.Background(),
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectExec("INSERT INTO `users` .*").
					WillReturnError(errors.New("数据库错误！"))
				return mockDB
			},
			user:    User{},
			wantErr: errors.New("数据库错误！"),
		},
		{
			name: "插入成功！",
			ctx:  context.Background(),
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				require.NoError(t, err)
				res := sqlmock.NewResult(3, 1)
				// 增删改
				mock.ExpectExec("INSERT INTO `users` .*").
					WillReturnResult(res)
				// 查
				//mock.ExpectQuery()
				return mockDB
			},
			user: User{
				Email: sql.NullString{
					String: "l0slakers@gmail.com",
					Valid:  true,
				},
			},
			wantErr: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			db, err := gorm.Open(gormMysql.New(gormMysql.Config{
				Conn: tc.mock(t),
				// 如果为 false ，则GORM在初始化时，会先调用 show version
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				// 如果为 true ，则不允许 Ping数据库
				DisableAutomaticPing: true,
				// 如果为 false ，则即使是单一语句，也会开启事务
				SkipDefaultTransaction: true,
			})
			d := NewUserInfoDAO(db)
			err = d.Insert(tc.ctx, tc.user)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestGormUserDAO_FindByEmail(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name     string
		ctx      context.Context
		email    string
		mock     func(t *testing.T) *sql.DB
		wantUser User
		wantErr  error
	}{
		{
			name:  "查询失败！",
			ctx:   context.Background(),
			email: "l0slakers@gmail.com",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				require.NoError(t, err)
				rows := sqlmock.NewRows([]string{
					"id", "email", "phone", "password", "nickname", "birthday", "create_time", "update_time",
				})
				mock.ExpectQuery("SELECT .*").
					WithArgs("l0slakers@gmail.com").
					WillReturnRows(rows)
				return mockDB
			},
			wantUser: User{},
			wantErr:  ErrDataNotFound,
		},
		{
			name:  "查询成功",
			ctx:   context.Background(),
			email: "l0slakers@gmail.com",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				require.NoError(t, err)
				rows := sqlmock.NewRows([]string{
					"id", "email", "phone", "password", "nickname", "birthday", "create_time", "update_time",
				}).AddRow(
					int64(123),
					"l0slakers@gmail.com",
					"13377818900",
					"$2a$10$K0T3cJ5hAbFIAhJiRcd1durGTO7/E5pn7nYPmk6f9bTkixxMMtEmm",
					"l0slakers",
					"2000-12-14",
					now.UnixMilli(),
					now.UnixMilli(),
				)
				mock.ExpectQuery("SELECT .*").
					WithArgs("l0slakers@gmail.com").
					WillReturnRows(rows)
				return mockDB
			},
			wantUser: User{
				Id: 123,
				Email: sql.NullString{
					String: "l0slakers@gmail.com",
					Valid:  true,
				},
				Phone: sql.NullString{
					String: "13377818900",
					Valid:  true,
				},
				Password:   "$2a$10$K0T3cJ5hAbFIAhJiRcd1durGTO7/E5pn7nYPmk6f9bTkixxMMtEmm",
				Nickname:   "l0slakers",
				Birthday:   "2000-12-14",
				CreateTime: now.UnixMilli(),
				UpdateTime: now.UnixMilli(),
			},
			wantErr: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := gorm.Open(gormMysql.New(gormMysql.Config{
				Conn:                      tc.mock(t),
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				DisableAutomaticPing:   true,
				SkipDefaultTransaction: true,
			})
			d := NewUserInfoDAO(db)
			require.NoError(t, err)
			user, err := d.FindByEmail(tc.ctx, tc.email)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}

func TestGormUserDAO_FindByPhone(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name     string
		ctx      context.Context
		phone    string
		mock     func(t *testing.T) *sql.DB
		wantUser User
		wantErr  error
	}{
		{
			name:  "查询失败！",
			ctx:   context.Background(),
			phone: "13377818900",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				require.NoError(t, err)
				rows := sqlmock.NewRows([]string{
					"id", "email", "phone", "password", "nickname", "birthday", "create_time", "update_time",
				})
				mock.ExpectQuery("SELECT .*").
					WithArgs("13377818900").
					WillReturnRows(rows)
				return mockDB
			},
			wantUser: User{},
			wantErr:  ErrDataNotFound,
		},
		{
			name:  "查询成功",
			ctx:   context.Background(),
			phone: "13377818900",
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				require.NoError(t, err)
				rows := sqlmock.NewRows([]string{
					"id", "email", "phone", "password", "nickname", "birthday", "create_time", "update_time",
				}).AddRow(
					int64(123),
					"l0slakers@gmail.com",
					"13377818900",
					"$2a$10$K0T3cJ5hAbFIAhJiRcd1durGTO7/E5pn7nYPmk6f9bTkixxMMtEmm",
					"l0slakers",
					"2000-12-14",
					now.UnixMilli(),
					now.UnixMilli(),
				)
				mock.ExpectQuery("SELECT .*").
					WithArgs("13377818900").
					WillReturnRows(rows)
				return mockDB
			},
			wantUser: User{
				Id: 123,
				Email: sql.NullString{
					String: "l0slakers@gmail.com",
					Valid:  true,
				},
				Phone: sql.NullString{
					String: "13377818900",
					Valid:  true,
				},
				Password:   "$2a$10$K0T3cJ5hAbFIAhJiRcd1durGTO7/E5pn7nYPmk6f9bTkixxMMtEmm",
				Nickname:   "l0slakers",
				Birthday:   "2000-12-14",
				CreateTime: now.UnixMilli(),
				UpdateTime: now.UnixMilli(),
			},
			wantErr: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := gorm.Open(gormMysql.New(gormMysql.Config{
				Conn:                      tc.mock(t),
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				DisableAutomaticPing:   true,
				SkipDefaultTransaction: true,
			})
			d := NewUserInfoDAO(db)
			require.NoError(t, err)
			user, err := d.FindByPhone(tc.ctx, tc.phone)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}

func TestGormUserDAO_FindById(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name     string
		ctx      context.Context
		id       int64
		mock     func(t *testing.T) *sql.DB
		wantUser User
		wantErr  error
	}{
		{
			name: "查询失败！",
			ctx:  context.Background(),
			id:   123,
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				require.NoError(t, err)
				rows := sqlmock.NewRows([]string{
					"id", "email", "phone", "password", "nickname", "birthday", "create_time", "update_time",
				})
				mock.ExpectQuery("SELECT .*").
					WithArgs(123).
					WillReturnRows(rows)
				return mockDB
			},
			wantUser: User{},
			wantErr:  ErrDataNotFound,
		},
		{
			name: "查询成功",
			ctx:  context.Background(),
			id:   123,
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				require.NoError(t, err)
				rows := sqlmock.NewRows([]string{
					"id", "email", "phone", "password", "nickname", "birthday", "create_time", "update_time",
				}).AddRow(
					int64(123),
					"l0slakers@gmail.com",
					"13377818900",
					"$2a$10$K0T3cJ5hAbFIAhJiRcd1durGTO7/E5pn7nYPmk6f9bTkixxMMtEmm",
					"l0slakers",
					"2000-12-14",
					now.UnixMilli(),
					now.UnixMilli(),
				)
				mock.ExpectQuery("SELECT .*").
					WithArgs(123).
					WillReturnRows(rows)
				return mockDB
			},
			wantUser: User{
				Id: 123,
				Email: sql.NullString{
					String: "l0slakers@gmail.com",
					Valid:  true,
				},
				Phone: sql.NullString{
					String: "13377818900",
					Valid:  true,
				},
				Password:   "$2a$10$K0T3cJ5hAbFIAhJiRcd1durGTO7/E5pn7nYPmk6f9bTkixxMMtEmm",
				Nickname:   "l0slakers",
				Birthday:   "2000-12-14",
				CreateTime: now.UnixMilli(),
				UpdateTime: now.UnixMilli(),
			},
			wantErr: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := gorm.Open(gormMysql.New(gormMysql.Config{
				Conn:                      tc.mock(t),
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				DisableAutomaticPing:   true,
				SkipDefaultTransaction: true,
			})
			require.NoError(t, err)
			d := NewUserInfoDAO(db)
			user, err := d.FindById(tc.ctx, tc.id)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}

func TestGormUserDAO_CompleteInfo(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name    string
		mock    func(t *testing.T) *sql.DB
		ctx     context.Context
		user    User
		wantErr error
	}{
		{
			name: "更新失败！",
			ctx:  context.Background(),
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				require.NoError(t, err)
				mock.ExpectExec("UPDATE `users` .*").
					WillReturnError(errors.New("WHERE conditions required"))
				return mockDB
			},
			user:    User{},
			wantErr: errors.New("WHERE conditions required"),
		},
		{
			name: "更新成功！",
			ctx:  context.Background(),
			mock: func(t *testing.T) *sql.DB {
				mockDB, mock, err := sqlmock.New()
				require.NoError(t, err)
				res := sqlmock.NewResult(1, 3)
				mock.ExpectExec("UPDATE `users` .*").
					WillReturnResult(res)
				return mockDB
			},
			user: User{
				Id:         123,
				Nickname:   "Loslakers",
				Birthday:   "2020-12-24",
				UpdateTime: now.UnixMilli(),
			},
			wantErr: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, err := gorm.Open(gormMysql.New(gormMysql.Config{
				Conn:                      tc.mock(t),
				SkipInitializeWithVersion: true,
			}), &gorm.Config{
				DisableAutomaticPing:   true,
				SkipDefaultTransaction: true,
			})
			require.NoError(t, err)
			d := NewUserInfoDAO(db)
			err = d.CompleteInfo(tc.ctx, tc.user)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
