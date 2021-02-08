package trace

import (
	"context"
	"fmt"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	jaeger "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"io"
)

var (
	tracer       opentracing.Tracer
	closer       io.Closer
	noopTrace    = &opentracing.NoopTracer{}
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

// Deprecated: Please close through the closer returned by initjaeger
func Close() {
	if closer != nil {
		closer.Close()
	}
}

func GlobalTracer() opentracing.Tracer {
	if tracer == nil {
		return noopTrace
	}
	return tracer
}

// get span
func StartSpan(name string) opentracing.Span {
	if tracer == nil {
		return noopTrace.StartSpan(name)
	}
	return tracer.StartSpan(name)
}

// first, context does not contain TraceId
// returns new context and span
// Deprecated: please use NewSpanFromContext
func NewContextWithSpanName(ctx context.Context, spanName string) (newCtx context.Context, span opentracing.Span) {
	if tracer == nil {
		return ctx, noopTrace.StartSpan(spanName)
	}
	span = StartSpan(spanName)
	newCtx = opentracing.ContextWithSpan(ctx, span)
	return
}

func NewSpanFromContext(ctx context.Context, spanName string) (span opentracing.Span, newCtx context.Context) {
	if tracer == nil {
		return noopTrace.StartSpan(spanName), ctx
	}
	span = StartSpan(spanName)
	newCtx = opentracing.ContextWithSpan(ctx, span)
	return
}

// get span by ctx
// return span
func SpanFromContext(ctx context.Context) opentracing.Span {
	if tracer == nil {
		return noopTrace.StartSpan("")
	}
	return opentracing.SpanFromContext(ctx)
}

// generate new ctx and span by ctx
// return new span and ctx
func StartSpanFromContext(ctx context.Context, opentionName string) (opentracing.Span, context.Context) {
	if tracer == nil {
		return noopTrace.StartSpan(opentionName), ctx
	}
	return opentracing.StartSpanFromContext(ctx, opentionName)
}

// generate new ctx by traceId
// return new span and context
func SpanContextWithTeaceId(traceId string, spanName string) (opentracing.Span, context.Context) {
	carrier := opentracing.HTTPHeadersCarrier{}
	carrier.Set("Uber-Trace-Id", traceId)
	if tracer == nil {
		return NewSpanFromContext(context.Background(), spanName)
	}
	wireContext, err := tracer.Extract(opentracing.HTTPHeaders, carrier)
	if err != nil {
		return NewSpanFromContext(context.Background(), spanName)
	}

	// traceId context => serverSpan
	serverSpan := opentracing.StartSpan(
		spanName,
		ext.RPCServerOption(wireContext))
	return serverSpan, opentracing.ContextWithSpan(context.Background(), serverSpan)
}

// get traceId by ctx
// return traceId
func TraceIDFromContext(ctx context.Context) (traceId string) {
	if ctx == nil {
		return ""
	}
	carrier := opentracing.HTTPHeadersCarrier{}

	// get span by context
	span := opentracing.SpanFromContext(ctx)
	if span == opentracing.Span(nil) {
		return ""
	}
	if tracer == nil {
		return ""
	}
	err := tracer.Inject(span.Context(), opentracing.HTTPHeaders, carrier)
	if err != nil {
		return ""
	}
	if v, ok := carrier["Uber-Trace-Id"]; ok {
		if len(v) == 1 {
			return v[0]
		}
	}
	return ""
}
