package tracer

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pojol/braid/3rd/log"
	"github.com/uber/jaeger-client-go"
)

// HTTPTracer http request tracer
type HTTPTracer struct {
	span      opentracing.Span
	requestID string
	beginTime time.Time
}

// Begin 开始捕捉
func (t *HTTPTracer) Begin(ectx echo.Context) {

	tr := opentracing.GlobalTracer()
	req := ectx.Request()
	t.beginTime = time.Now()

	if ctx, err := tr.Extract(opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header)); err != nil {
		t.span = tr.StartSpan(optionHTTPRequest)
	} else {
		t.span = tr.StartSpan(optionHTTPRequest, ext.RPCServerOption(ctx))
	}

	ext.HTTPMethod.Set(t.span, req.Method)
	ext.HTTPUrl.Set(t.span, req.URL.String())

	rc := opentracing.ContextWithSpan(req.Context(), t.span)

	// Set request ID for context.
	if sc, ok := t.span.Context().(jaeger.SpanContext); ok {
		t.requestID = sc.TraceID().String()
		rc = context.WithValue(rc, RequestKey, t.requestID)
	}

	req = req.WithContext(opentracing.ContextWithSpan(rc, t.span))
	ectx.SetRequest(req)
}

// End 结束捕捉
func (t *HTTPTracer) End(ectx echo.Context) {

	status := ectx.Response().Status
	committed := ectx.Response().Committed

	ext.HTTPStatusCode.Set(t.span, uint16(status))
	if status >= http.StatusInternalServerError || !committed {
		ext.Error.Set(t.span, true)
	}

	executionTime := time.Now().Sub(t.beginTime)
	if executionTime > tracer.cfg.SlowRequest {
		log.SysSlow(ectx.Path(), t.requestID, int(executionTime), "slow request")
	}

	t.span.Finish()
}
