package tracer

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// RedisTracer redis tracer
type RedisTracer struct {
	span opentracing.Span
	Cmd  string
}

// Begin 开始监听
func (r *RedisTracer) Begin(ctx context.Context) {
	gt := opentracing.GlobalTracer()
	parentSpan := opentracing.SpanFromContext(ctx)
	if parentSpan != nil {

		r.span = gt.StartSpan(r.Cmd, opentracing.ChildOf(parentSpan.Context()))
		ext.DBType.Set(r.span, "Redis")
	}

}

// End 结束监听
func (r *RedisTracer) End() {

	r.span.Finish()

}
