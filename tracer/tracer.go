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
	Tracer struct {
		closer  io.Closer
		tracing opentracing.Tracer
		cfg     Config
	}

	//Config 链路追踪配置
	Config struct {
		Endpoint      string        // jaeger 地址
		Probabilistic float64       // 采样率
		Name          string        // tracer name
		SlowRequest   int64         // 一旦request超出设置的SlowRequest（ms）时间，则一定会有一条slow日志
		SlowSpan      time.Duration // 一旦span超出设置的SlowSpan（ms）时间，则一定会有一条slow日志
	}
)

const (
	optionHTTPRequest = "Request"
	requestID         = "RequestID"
	opRequestLimit    = "RequestLimit"
)

var (
	// https://github.com/jaegertracing/jaeger-client-go/blob/master/config/config.go
	// DefaultTracerConfig 默认tracer配置
	defaultTracerConfig = Config{
		Endpoint: "http://localhost:14268/api/traces",
		// 采样率 0 ～ 1
		Probabilistic: 1,
		SlowRequest:   100,
		SlowSpan:      10,
	}

	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("Convert linker config")

	tracer *Tracer
)

func New() *Tracer {
	tracer = &Tracer{}
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

func (t *Tracer) Init(cfg interface{}) error {
	tCfg, ok := cfg.(Config)
	if !ok {
		return ErrConfigConvert
	}

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
			CollectorEndpoint: tCfg.Endpoint,
		},
		ServiceName: tCfg.Name,
	}

	sender, err := newTransport(jcfg.Reporter)
	if err != nil {
		return err
	}

	r = jaegerCfg.Reporter(NewSlowReporter(sender, nil, tCfg.Probabilistic))
	m = jaegerCfg.Metrics(metrics.NullFactory)

	tracer, closer, err = jcfg.NewTracer(r, m)
	if err != nil {
		return err
	}

	opentracing.SetGlobalTracer(tracer)
	t.tracing = tracer
	t.closer = closer
	t.cfg = tCfg
	return nil
}

func (t *Tracer) Run() {

}

func (t *Tracer) Close() {

}
