package depend

import (
	"github.com/pojol/braid-go/depend/bconsul"
	"github.com/pojol/braid-go/depend/blog"
	"github.com/pojol/braid-go/depend/bredis"
	"github.com/pojol/braid-go/depend/btracer"
)

type BraidDepend struct {
	Tracer       btracer.ITracer
	ConsulClient *bconsul.Client
	Logger       *blog.Logger
	RedisClient  *bredis.Client
}

type Depend func(*BraidDepend)

func Logger(log *blog.Logger) Depend {
	return func(d *BraidDepend) {
		d.Logger = log
	}
}

func Redis(client *bredis.Client) Depend {
	return func(d *BraidDepend) {
		d.RedisClient = client
	}
}

func Tracer(opts ...btracer.Option) Depend {
	return func(d *BraidDepend) {
		d.Tracer = btracer.BuildWithOption(opts...)
	}
}

func Consul(client *bconsul.Client) Depend {
	return func(d *BraidDepend) {
		d.ConsulClient = client
	}
}
