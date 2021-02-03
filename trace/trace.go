package trace

import (
	"context"
	"fmt"
	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"io"
)

var (
	tracer opentracing.Tracer
	closer io.Closer
)

func InitJaegerWithConfig() {
	conf := MustLoadConfig()
	InitJaeger(conf)
	return
}

func InitJaeger(config *Config) {
	cfg := config.ToTraceConfiguration()
	var (
		err error
	)
	tracer, closer, err = cfg.NewTracer(jaegercfg.Logger(jaeger.StdLogger))

	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	opentracing.SetGlobalTracer(tracer)
	return
}

func Close() {
	if closer != nil {
		closer.Close()
	}
}

func GlobalTracer() opentracing.Tracer {
	return tracer
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


