package cron_job

import (
	"github.com/robfig/cron/v3"
	"log"
	"testing"
	"time"
)

type myJob struct {
}

func (m myJob) Run() {
	log.Println("I am coming")
}

func TestCronJob(t *testing.T) {
	// Second | Minute | Hour | Dom | Month | Dow | Descriptor
	// 秒 | 分 | 时 | 日期 | 月 | 星期 | 年
	expr := cron.New(cron.WithSeconds())
	//expr.AddJob("@every 1s", myJob{})
	expr.AddFunc("@every 3s", func() {
		t.Log("开始长任务")
		time.Sleep(time.Second * 12)
		t.Log("结束长任务")
	})
	expr.Start()
	// 模拟运行10s
	time.Sleep(time.Second * 10)
	stop := expr.Stop()
	t.Log("已经发出停止信号")
	<-stop.Done()
	t.Log("彻底结束")
}
