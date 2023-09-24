package logger

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"io"
	"time"
)

type MiddlewareBuilder struct {
	allowReqBody  bool
	allowRespBody bool
	loggerFunc    func(ctx context.Context, al *AccessLog)
}

type AccessLog struct {
	Method   string
	Url      string
	ReqBody  string
	RespBody string
	Status   int
	Duration string
}

func NewMiddlewareBuilder(fn func(ctx context.Context, al *AccessLog)) *MiddlewareBuilder {
	return &MiddlewareBuilder{
		loggerFunc: fn,
	}
}

func (m *MiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		url := ctx.Request.URL.String()
		if len(url) > 1024 {
			url = url[:1024]
		}
		al := &AccessLog{
			Method: ctx.Request.Method,
			// url 可能会有长度问题：攻击者传入很长的URL，或者传入巨大的请求
			Url: url,
		}
		// 读取请求
		if m.allowReqBody && ctx.Request.Body != nil {
			// 这是一个很消耗 CPU 和 内存的操作，因为会引起复制
			//body, _ := io.ReadAll(ctx.Request.Body)
			body, _ := ctx.GetRawData()
			// io流读取出数据之后需要放回去，否则数据会丢失
			ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
			if len(body) > 1024 {
				body = body[:1024]
			}
			al.ReqBody = string(body)
		}

		// 读取请求
		if m.allowRespBody {
			ctx.Writer = &responseWriter{
				al:             al,
				ResponseWriter: ctx.Writer,
			}
		}

		// 防止ctx.Next()崩溃之后没有打印结果
		defer func() {
			al.Duration = time.Since(start).String()
			m.loggerFunc(ctx, al)
		}()

		// 执行到业务逻辑
		ctx.Next()
	}
}

func (m *MiddlewareBuilder) AllowReqBody() *MiddlewareBuilder {
	m.allowReqBody = true
	return m
}

func (m *MiddlewareBuilder) AllowRespBody() *MiddlewareBuilder {
	m.allowRespBody = true
	return m
}

type responseWriter struct {
	al *AccessLog
	gin.ResponseWriter
}

func (r *responseWriter) WriteHeader(statusCode int) {
	r.al.Status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseWriter) Write(data []byte) (int, error) {
	r.al.RespBody = string(data)
	return r.ResponseWriter.Write(data)
}

func (r *responseWriter) WriteString(data string) (int, error) {
	r.al.RespBody = data
	return r.ResponseWriter.WriteString(data)
}
