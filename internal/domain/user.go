package domain

import "time"

type User struct {
	Id    int64
	Email string
	Phone string
	// 尽量不要组合，万一有新的服务 DingdingInfo
	WechatInfo
	Password string
	Nickname string
	Birthday string
	Ctime    time.Time
	Utime    time.Time
}
