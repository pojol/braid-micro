package moduleparm

import (
	"github.com/pojol/braid-go/module/balancer"
	"github.com/pojol/braid-go/module/linkcache"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/pubsub"
	"github.com/pojol/braid-go/module/tracer"
)

type BuildParm struct {
	Name string

	PS        pubsub.IPubsub
	Logger    logger.ILogger
	Tracer    tracer.ITracer
	Balancer  balancer.IBalancer
	Linkcache linkcache.ILinkCache
}

type Option func(*BuildParm)

func WithPubsub(ps pubsub.IPubsub) Option {
	return func(bp *BuildParm) {
		bp.PS = ps
	}
}

func WithLogger(log logger.ILogger) Option {
	return func(bp *BuildParm) {
		bp.Logger = log
	}
}

func WithTracer(trarcer tracer.ITracer) Option {
	return func(bp *BuildParm) {
		bp.Tracer = trarcer
	}
}

func WithBalancer(b balancer.IBalancer) Option {
	return func(bp *BuildParm) {
		bp.Balancer = b
	}
}

func WithLinkcache(lc linkcache.ILinkCache) Option {
	return func(bp *BuildParm) {
		bp.Linkcache = lc
	}
}
