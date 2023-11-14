package web

import (
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"time"
)

type ObservabilityHandler struct {
}

func (o *ObservabilityHandler) RegisterRoutes(server *gin.Engine) {
	group := server.Group("test")
	group.GET("/metric", func(ctx *gin.Context) {
		// 取一个随机数
		sleep := rand.Int31n(1000)
		time.Sleep(time.Millisecond * time.Duration(sleep))
		ctx.String(http.StatusOK, "ok!")
	})
}
