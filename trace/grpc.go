package trace

import (
	"context"
	"encoding/base64"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"strings"
)

func TraceSpanServerInterceptor() grpc.UnaryServerInterceptor {
	if tracer == nil {
		return func(
			ctx context.Context,
			req interface{},
			info *grpc.UnaryServerInfo,
			handler grpc.UnaryHandler,
		) (resp interface{}, err error) {
			return handler(ctx, req)
		}
	} else {
		return otgrpc.OpenTracingServerInterceptor(tracer)
	}
}

func TraceSpanClientInterceptor() grpc.UnaryClientInterceptor {
	if tracer == nil {
		return func(
			ctx context.Context,
			method string, req, resp interface{},
			cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
		) (err error) {
			return invoker(ctx, method, req, resp, cc, opts...)
		}
	} else {
		return func(
			ctx context.Context,
			method string, req, resp interface{},
			cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
		) (err error) {
			span := opentracing.SpanFromContext(ctx)
			// Save current span context.
			md, ok := metadata.FromOutgoingContext(ctx)
			if !ok {
				md = metadata.Pairs()
			}
			if err = opentracing.GlobalTracer().Inject(
				span.Context(), opentracing.HTTPHeaders, metadataTextMap(md),
			); err != nil {
				log.Print(ctx, "Failed to inject trace span: %v", err)
			}
			return invoker(metadata.NewOutgoingContext(ctx, md), method, req, resp, cc, opts...)
		}
	}
}

const (
	binHeaderSuffix = "_bin"
)

// metadataTextMap extends a metadata.MD to be an opentracing textmap
type metadataTextMap metadata.MD

// Set is a opentracing.TextMapReader interface that extracts values.
func (m metadataTextMap) Set(key, val string) {
	// gRPC allows for complex binary values to be written.
	encodedKey, encodedVal := encodeKeyValue(key, val)
	// The metadata object is a multimap, and previous values may exist, but for opentracing headers, we do not append
	// we just override.
	m[encodedKey] = []string{encodedVal}
}

// ForeachKey is a opentracing.TextMapReader interface that extracts values.
func (m metadataTextMap) ForeachKey(callback func(key, val string) error) error {
	for k, vv := range m {
		for _, v := range vv {
			if err := callback(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

// encodeKeyValue encodes key and value qualified for transmission via gRPC.
// note: copy pasted from private values of grpc.metadata
func encodeKeyValue(k, v string) (string, string) {
	k = strings.ToLower(k)
	if strings.HasSuffix(k, binHeaderSuffix) {
		val := base64.StdEncoding.EncodeToString([]byte(v))
		v = val
	}
	return k, v
}
