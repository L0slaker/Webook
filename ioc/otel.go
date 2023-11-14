package ioc

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"time"
)

func InitOTEL() func(ctx context.Context) {
	res, err := newResource("webook", "v0.0.1")
	if err != nil {
		panic(err)
	}
	prop := newPropagator()
	// 在客户端和服务端之间传递 tracing 的相关信息
	otel.SetTextMapPropagator(prop)

	// 初始化 trace provider
	tp, err := newTraceProvider(res)
	if err != nil {
		panic(err)
	}
	otel.SetTracerProvider(tp)
	return func(ctx context.Context) {
		tp.Shutdown(ctx)
	}
}

// newResource
// 该函数用于创建一个 OpenTelemetry 资源。
// 使用 resource.Merge 将默认资源与包含服务名称和版本的新资源合并。
// 返回资源和一个可能的错误。
func newResource(serviceName, serviceVersion string) (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
		))
}

// newTraceProvider
// 该函数初始化一个新的 OpenTelemetry 追踪器提供程序。
// 创建一个 Zipkin 导出器，配置为将跨度发送到 "http://localhost:9411/api/v2/spans"。
// 使用 trace.WithBatcher 选项设置了一个一秒的批处理超时。
// 使用 trace.WithResource 选项设置了追踪器提供程序的资源。
// 返回追踪器提供程序和一个可能的错误。
func newTraceProvider(res *resource.Resource) (*trace.TracerProvider, error) {
	exporter, err := zipkin.New("http://localhost:9411/api/v2/spans")
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter,
			// Default is 5s. Set to 1s for demonstrative purposes.
			trace.WithBatchTimeout(time.Second)),
		trace.WithResource(res),
	)
	return traceProvider, nil
}

// newPropagator
// 该函数创建一个新的 OpenTelemetry 传播器，用于在客户端和服务端之间传递追踪信息。
// 返回一个复合文本映射传播器，包括 propagation.TraceContext{} 和 propagation.Baggage{}。
func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}
