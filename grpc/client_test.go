package grpc

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	// 建立到地址 ":8090" 的 gRPC 服务器连接，使用不安全的传输凭证
	conn, err := grpc.Dial(":8090",
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := NewUserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := client.GetById(ctx, &GetByIdReq{
		Id: 12,
	})
	assert.NoError(t, err)
	t.Log(resp.User)
}
