package tracer

import "time"

//config 链路追踪配置
// https://github.com/jaegertracing/jaeger-client-go/blob/master/config/config.go
type tconfig struct {
	CollectorEndpoint  string        // jaeger http地址
	LocalAgentHostPort string        // jaeger udp地址
	Probabilistic      float64       // 采样率
	Name               string        // tracer name
	SlowRequest        time.Duration // 一旦request超出设置的SlowRequest（ms）时间，则一定会有一条slow日志
	SlowSpan           time.Duration // 一旦span超出设置的SlowSpan（ms）时间，则一定会有一条slow日志
}

// Option config wraps
type Option func(*Tracer)

// WithProbabilistic 采样率 0 ~ 1
func WithProbabilistic(probabilistic float64) Option {
	return func(t *Tracer) {
		t.cfg.Probabilistic = probabilistic
	}
}

// WithSlowRequest 单次请求阀值（超过会有慢日志
func WithSlowRequest(ms int) Option {
	return func(t *Tracer) {
		t.cfg.SlowRequest = time.Duration(ms) * time.Millisecond
	}
}

// WithSlowSpan 单次调用阀值（超过会有慢日志
func WithSlowSpan(ms int) Option {
	return func(t *Tracer) {
		t.cfg.SlowSpan = time.Duration(ms) * time.Millisecond
	}
}

// WithHTTP http://172.17.0.1:14268/api/traces
func WithHTTP(CollectorEndpoint string) Option {
	return func(t *Tracer) {
		t.cfg.CollectorEndpoint = CollectorEndpoint
	}
}

// WithUDP 172.17.0.1:6831
func WithUDP(LocalAgentHostPort string) Option {
	return func(t *Tracer) {
		t.cfg.LocalAgentHostPort = LocalAgentHostPort
	}
}
