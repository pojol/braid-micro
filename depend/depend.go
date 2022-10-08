package depend

import (
	"github.com/pojol/braid-go/depend/blog"
	"github.com/pojol/braid-go/depend/consul"
	"github.com/pojol/braid-go/depend/redis"
	"github.com/pojol/braid-go/depend/tracer"
)

type BraidDepend struct {
	Tracer       tracer.ITracer
	ConsulClient *consul.Client
	Logger       *blog.Logger
	RedisClient  *redis.Client
}

type Depend func(*BraidDepend)

func Logger(opts ...blog.Option) Depend {
	return func(d *BraidDepend) {
		d.Logger = blog.BuildWithOption(opts...)
	}
}

func Redis(opts ...redis.Option) Depend {
	return func(d *BraidDepend) {
		d.RedisClient = redis.BuildWithOption(opts...)
	}
}

func Tracer(opts ...tracer.Option) Depend {
	return func(d *BraidDepend) {
		d.Tracer = tracer.BuildWithOption(opts...)
	}
}

func Consul(opts ...consul.Option) Depend {
	return func(d *BraidDepend) {
		d.ConsulClient = consul.BuildWithOption(opts...)
	}
}
