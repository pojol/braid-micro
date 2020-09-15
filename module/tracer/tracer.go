package tracer

import (
	"errors"
	"fmt"
	"io"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegerCfg "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/transport"
	"github.com/uber/jaeger-lib/metrics"
)

type (
	// Tracer tracer struct
	Tracer struct {
		closer  io.Closer
		tracing opentracing.Tracer
		cfg     tconfig
	}
)

const (
	optionHTTPRequest = "Request"
	opRequestLimit    = "RequestLimit"
)

// TraceKey 主键类型
type TraceKey int

const (
	// RequestKey 请求的键值
	RequestKey TraceKey = 1000 + iota
)

var (

	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("Convert linker config")

	tracer *Tracer
)

func newTransport(rc *jaegerCfg.ReporterConfig) (jaeger.Transport, error) {
	switch {
	case rc.CollectorEndpoint != "":
		httpOptions := []transport.HTTPOption{transport.HTTPBatchSize(1), transport.HTTPHeaders(rc.HTTPHeaders)}
		if rc.User != "" && rc.Password != "" {
			httpOptions = append(httpOptions, transport.HTTPBasicAuth(rc.User, rc.Password))
		}
		return transport.NewHTTPTransport(rc.CollectorEndpoint, httpOptions...), nil
	default:
		return jaeger.NewUDPTransport(rc.LocalAgentHostPort, 0)
	}
}

// New 创建 jaeger traing
func New(name string, protoOpt Option, opts ...Option) (*Tracer, error) {

	const (
		defaultProbabilistic = 1
		defaultSlowRequest   = time.Millisecond * 100
		defaultSlowSpan      = time.Millisecond * 10
	)

	tracer = &Tracer{
		cfg: tconfig{
			Probabilistic: defaultProbabilistic,
			Name:          name,
			SlowRequest:   defaultSlowRequest,
			SlowSpan:      defaultSlowSpan,
		},
	}

	protoOpt(tracer)

	for _, opt := range opts {
		opt(tracer)
	}

	jcfg := jaegerCfg.Configuration{
		Sampler: &jaegerCfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegerCfg.ReporterConfig{
			LogSpans:           true,
			CollectorEndpoint:  tracer.cfg.CollectorEndpoint, //with http
			LocalAgentHostPort: tracer.cfg.LocalAgentHostPort,
		},
		ServiceName: tracer.cfg.Name,
	}

	sender, err := newTransport(jcfg.Reporter)
	if err != nil {
		fmt.Println("new transport err", err)
		return nil, err
	}

	r := jaegerCfg.Reporter(NewSlowReporter(sender, nil, tracer.cfg.Probabilistic))
	m := jaegerCfg.Metrics(metrics.NullFactory)

	jtracing, closer, err := jcfg.NewTracer(r, m)
	if err != nil {
		fmt.Println("new tracer err", err)
		return nil, err
	}

	tracer.tracing = jtracing
	tracer.closer = closer
	opentracing.SetGlobalTracer(tracer.tracing)

	return tracer, nil
}

// Close 关闭tracing
func (t *Tracer) Close() {
	t.closer.Close()
}
