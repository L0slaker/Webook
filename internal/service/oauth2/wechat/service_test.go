//go:build e2e

package wechat

import (
	"context"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func Test_service_e2e_VerifyCode(t *testing.T) {
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("没有找到环境变量 WECHAT_APP_ID")
	}
	appSecret, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("没有找到环境变量 WECHAT_APP_SECRET")
	}
	svc := NewService(appId, appSecret)
	res, err := svc.VerifyCode(context.Background(), "081u101w3n1uk13FiK1w3RXxyS0u1011", "7c7kwSHQeDUC2e6oiWN58i")
	require.NoError(t, err)
	t.Log(res)
}
