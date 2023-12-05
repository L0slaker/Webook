package domain

import (
	"github.com/robfig/cron/v3"
	"time"
)

type Job struct {
	Id         int64
	Cfg        string
	Name       string       // 比如 ranking 任务
	Executor   string       // 任务所需的执行器
	Cron       string       // cron 表达式，表示任务调度时间规则
	CancelFunc func() error // 取消函数
}

var parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom |
	cron.Month | cron.Dow | cron.Descriptor)

func (j Job) NextTime() time.Time {
	s, _ := parser.Parse(j.Cron)
	return s.Next(time.Now())
}
