package jaegertracing

import (
	"context"
	"errors"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pojol/braid/module/tracer"
)

const (
	// RedisSpan redis span
	RedisSpan = "tracer_span_redis"
)

// RedisTracer redis tracer
type RedisTracer struct {
	span    opentracing.Span
	tracing opentracing.Tracer

	starting bool

	Cmd string
}

func createRedisSpanFactory() tracer.SpanFactory {
	return func(tracing interface{}) (tracer.ISpan, error) {

		t, ok := tracing.(opentracing.Tracer)
		if !ok {
			return nil, errors.New("")
		}

		rt := &RedisTracer{
			tracing: t,
		}

		return rt, nil
	}
}

// Begin 开始监听
func (r *RedisTracer) Begin(ctx interface{}) {

	redisctx, ok := ctx.(context.Context)
	if !ok {
		return
	}

	parentSpan := opentracing.SpanFromContext(redisctx)
	if parentSpan != nil {
		r.span = r.tracing.StartSpan(r.Cmd, opentracing.ChildOf(parentSpan.Context()))
		ext.DBType.Set(r.span, "Redis")
	}

	r.starting = true
}

// End 结束监听
func (r *RedisTracer) End(ctx interface{}) {

	if !r.starting {
		return
	}

	if r.span != nil {
		r.span.Finish()
	}

}
