package opentelemetry

import (
	"Prove/webook/internal/service/sms"
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type OTELSMSService struct {
	svc    sms.Service
	tracer trace.Tracer
}

func NewOTELSMSService(svc sms.Service) sms.Service {
	tp := otel.GetTracerProvider()
	tracer := tp.Tracer("Prove/webook/internal/service/sms/opentelemetry")
	return &OTELSMSService{
		svc:    svc,
		tracer: tracer,
	}
}

func (o *OTELSMSService) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	// 创建一个名叫 "sms_send"+biz 的客户端 span
	ctx, span := o.tracer.Start(ctx, "sms_send"+biz,
		trace.WithSpanKind(trace.SpanKindClient))
	// 关闭 span，并记录栈信息
	defer span.End(trace.WithStackTrace(true))
	span.AddEvent("发送短信")
	err := o.svc.Send(ctx, biz, args, numbers...)
	if err != nil {
		span.RecordError(err)
	}
	return err
}
