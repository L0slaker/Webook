package dao

import (
	"context"
	"database/sql"
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
	Id int64 `gorm:"primaryKey;autoIncrement"`
	// 由于注册的情况，有邮箱注册，也有手机注册，选择其一会导致另一个为空，
	//所以我们要允许Email和Phone为空的情况。而我们设置了唯一索引，可能会引起冲突
	//Email      string `gorm:"unique"`
	//Phone      string `gorm:"unique"`
	// sql.NullString 唯一索引允许有多个空值，但不能有多个 ""
	Email      sql.NullString `gorm:"unique"`
	Phone      sql.NullString `gorm:"unique"`
	Password   string
	Nickname   string
	Birthday   string
	CreateTime int64
	UpdateTime int64
	//DeleteTime int64
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

func (dao *UserInfoDAO) FindByPhone(ctx context.Context, phone string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).First(&u, "phone = ?", phone).Error
	return u, err
}

func (dao *UserInfoDAO) FindById(ctx context.Context, id int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).First(&u, "`id` = ?", id).Error
	return u, err
}

func (dao *UserInfoDAO) CompleteInfo(ctx context.Context, u *User) error {
	res := dao.db.WithContext(ctx).Model(&u).Updates(User{
		Nickname:   u.Nickname,
		Birthday:   u.Birthday,
		UpdateTime: time.Now().UnixMilli(),
	})

	return res.Error
}
