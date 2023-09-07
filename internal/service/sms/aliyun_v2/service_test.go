package aliyun_v2

import (
	"context"
	"github.com/ecodeclub/ekit"
	"os"
	"testing"
)

func TestSendSMS(t *testing.T) {
	accessKeyID := "<Your Access Key ID>"
	accessKeySecret := "<Your Access Key Secret>"

	// 设置环境变量
	os.Setenv("ALIBABA_CLOUD_ACCESS_KEY_ID", accessKeyID)
	os.Setenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET", accessKeySecret)

	accessKeyId := "LTAI5tPd2puB2DMpFKyNupGP"
	accessSecret := "HHCb1QjkxWjJ2bIeL5tqwsJKMIOxHr"
	signName := "阿里云短信测试"
	tplId := "SMS_154950909"
	args := []string{"567344"}
	numbers := []string{"1", "3", "5", "0", "9", "5", "1", "6", "5", "2", "0"}

	svc := NewService(ekit.ToPtr[string](accessKeyId), ekit.ToPtr[string](accessSecret), signName)
	err := svc.Send(context.Background(), tplId, args, numbers...)
	if err != nil {
		t.Errorf("Error sending SMS: %v", err)
	}
}
