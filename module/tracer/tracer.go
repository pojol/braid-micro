package tracer

import (
	"errors"
	"io"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
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

// New n
func New(name string, jaegerAddress string, opts ...Option) *Tracer {

	const (
		defaultProbabilistic = 1
		defaultSlowRequest   = time.Millisecond * 100
		defaultSlowSpan      = time.Millisecond * 10
	)

	tracer = &Tracer{
		cfg: tconfig{
			Endpoint:      jaegerAddress,
			Probabilistic: defaultProbabilistic,
			Name:          name,
			SlowRequest:   defaultSlowRequest,
			SlowSpan:      defaultSlowSpan,
		},
	}

	for _, opt := range opts {
		opt(tracer)
	}

	return tracer
}

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

// Init 初始化
func (t *Tracer) Init() error {

	var tracer opentracing.Tracer
	var closer io.Closer
	var r, m config.Option

	jcfg := jaegerCfg.Configuration{
		Sampler: &jaegerCfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegerCfg.ReporterConfig{
			LogSpans:          false,
			CollectorEndpoint: t.cfg.Endpoint,
		},
		ServiceName: t.cfg.Name,
	}

	sender, err := newTransport(jcfg.Reporter)
	if err != nil {
		return err
	}

	r = jaegerCfg.Reporter(NewSlowReporter(sender, nil, t.cfg.Probabilistic))
	m = jaegerCfg.Metrics(metrics.NullFactory)

	tracer, closer, err = jcfg.NewTracer(r, m)
	if err != nil {
		return err
	}

	opentracing.SetGlobalTracer(tracer)
	t.tracing = tracer
	t.closer = closer
	return nil
}
