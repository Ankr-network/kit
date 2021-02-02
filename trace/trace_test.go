package trace

import (
	"context"
	"testing"
	"time"
)

func TestTrace(t *testing.T) {
	closer := InitJaegerWithConfig()
	defer closer.Close()

	ctx, span := NewContextWithSpanName(context.Background(), "ctx-test")
	defer span.Finish()
	testctx1(ctx)
}

func testctx1(ctx context.Context) {
	span, ctx := StartSpanFromContext(ctx, "ctx-test-1")
	defer span.Finish()
	time.Sleep(time.Second)
	testctx11(ctx)
}

func testctx11(ctx context.Context) {
	_, ctx = StartSpanFromContext(ctx, "ctx-test-1-1")
	span := SpanFromContext(ctx)
	defer span.Finish()
	time.Sleep(time.Millisecond * 100)
}
