//go:build wireinject

// 让 wire 注入代码
package wire

import (
	"Prove/wire/repository"
	"Prove/wire/repository/dao"
	"github.com/google/wire"
)

func InitRepository() *repository.UserInfoRepository {
	// 只需要在这里声明需要的各种东西，组装构造和顺序都能自动编排
	// 传入各个组件的初始化方法
	wire.Build(repository.NewUserInfoRepository,
		dao.NewUserInfoDAO, InitDB,
	)
	// 这里随意返回就行
	return new(repository.UserInfoRepository)
}
