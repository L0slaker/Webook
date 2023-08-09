package dao

import (
	"context"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var (
	ErrUserDuplicateEmail = errors.New("邮箱已被使用！")
	ErrDataNotFound       = gorm.ErrRecordNotFound
)

type User struct {
	Id         int64  `gorm:"primaryKey;autoIncrement"`
	Email      string `gorm:"unique"`
	Password   string
	Nickname   string
	Birthday   string
	CreateTime int64
	UpdateTime int64
	DeleteTime int64
}

type UserInfoDAO struct {
	db *gorm.DB
}

func NewUserInfoDAO(db *gorm.DB) *UserInfoDAO {
	return &UserInfoDAO{
		db: db,
	}
}

func InitTables(db *gorm.DB) error {
	return db.AutoMigrate(&User{})
}

func (dao *UserInfoDAO) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.CreateTime = now
	u.UpdateTime = now

	err := dao.db.WithContext(ctx).Create(&u).Error
	if e, ok := err.(*mysql.MySQLError); ok {
		const uniqueIndexErr uint16 = 1062
		// 检查错误编号是否表示唯一索引冲突
		if e.Number == uniqueIndexErr {
			return ErrUserDuplicateEmail
		}
	}
	return err
}

func (dao *UserInfoDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).First(&u, "email = ?", email).Error
	return u, err
}

func (dao *UserInfoDAO) FindById(ctx context.Context, id int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).First(&u, "id = ?", id).Error
	return u, err
}

func (dao *UserInfoDAO) CompleteInfo(ctx context.Context, u User) error {
	//1.找到对应的user
	user, err := dao.FindById(ctx, u.Id)
	if err != nil {
		return err
	}
	//2.更新user的nickname和birthday、updateTime字段
	user.Nickname = u.Nickname
	user.Birthday = u.Birthday
	now := time.Now().UnixMilli()
	user.UpdateTime = now

	//3.保存修改
	res := dao.db.Save(&user)
	return res.Error
}
