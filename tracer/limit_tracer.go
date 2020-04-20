package tracer

import (
	"github.com/labstack/echo/v4"
	opentracing "github.com/opentracing/opentracing-go"
)

// LimitTracer token request limit
type LimitTracer struct {
	span opentracing.Span
}

// Begin 开始捕捉
func (t *LimitTracer) Begin(ectx echo.Context) {
	var rootCtx opentracing.SpanContext

	rootSpan := opentracing.SpanFromContext(ectx.Request().Context())
	if rootSpan != nil {
		rootCtx = rootSpan.Context()
	}
	t.span = opentracing.GlobalTracer().StartSpan(
		opRequestLimit,
		opentracing.ChildOf(rootCtx),
	)
}

// End 结束捕捉
func (t *LimitTracer) End(ectx echo.Context) {

	t.span.Finish()
}
