package trace

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTrace(t *testing.T) {
	defer InitJaegerWithConfig().Close()

	span, ctx := NewSpanFromContext(context.Background(), "ctx-test")
	testctx1(ctx)
	span.Finish()
	testCtx(ctx, t)
}

func testCtx(ctx context.Context, t *testing.T) {
	span, ctx := StartSpanFromContext(ctx, "testCtx")
	defer span.Finish()

	t.Log(ctx)
	traceId := TraceIDFromContext(ctx)
	assert.NotEqual(t, traceId, "")
	t.Log(traceId)
	span, ctx = SpanContextWithTeaceId(traceId, "test")
	span = SpanFromContext(ctx)

	time.Sleep(time.Second)
	span.Finish()

	traceId = TraceIDFromContext(nil)
	assert.Equal(t, traceId, "")
	traceId = TraceIDFromContext(context.Background())
	assert.Equal(t, traceId, "")
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

func TestCtxAndSpan(t *testing.T) {
	defer InitJaegerWithConfig().Close()

	span, ctx := NewSpanFromContext(context.Background(), "test")
	t.Log(ctx)
	t.Log(span.Context())

	span, ctx = StartSpanFromContext(ctx, "test2")
	t.Log(span.Context())
	t.Log(ctx)
}
