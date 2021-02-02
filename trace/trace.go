package trace

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"google.golang.org/grpc"
	"io"
)

var (
	tracer opentracing.Tracer
	closer io.Closer
)

func InitJaegerWithConfig() io.Closer {
	conf := MustLoadConfig()
	return InitJaeger(conf)
}

func InitJaeger(config *Config) io.Closer {
	cfg := config.ToTraceConfiguration()
	var (
		err error
	)
	tracer, closer, err = cfg.NewTracer(jaegercfg.Logger(jaeger.StdLogger))

	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	opentracing.SetGlobalTracer(tracer)
	return closer
}

func GlobalTracer() opentracing.Tracer {
	return tracer
}

func TraceSpanServerInterceptor() grpc.UnaryServerInterceptor {
	return otgrpc.OpenTracingServerInterceptor(tracer)
}

func TraceSpanClientInterceptor() grpc.UnaryClientInterceptor {
	return otgrpc.OpenTracingClientInterceptor(tracer)
}

func StartSpan(name string) opentracing.Span {
	return tracer.StartSpan(name)
}

func NewContextWithSpanName(ctx context.Context, spanName string) (newCtx context.Context, span opentracing.Span) {
	span = StartSpan(spanName)
	newCtx = opentracing.ContextWithSpan(ctx, span)
	return
}

func SpanFromContext(ctx context.Context) opentracing.Span {
	return opentracing.SpanFromContext(ctx)
}

func StartSpanFromContext(ctx context.Context, opentionName string) (opentracing.Span, context.Context) {
	return opentracing.StartSpanFromContext(ctx, opentionName)
}
