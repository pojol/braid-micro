package module

import (
	"github.com/pojol/braid-go/depend/blog"
	"github.com/pojol/braid-go/depend/consul"
	"github.com/pojol/braid-go/depend/redis"
	"github.com/pojol/braid-go/depend/tracer"
	"github.com/pojol/braid-go/module/discover"
	"github.com/pojol/braid-go/module/elector"
	iclient "github.com/pojol/braid-go/module/internal/client"
	idiscover "github.com/pojol/braid-go/module/internal/discover_consul"
	iserver "github.com/pojol/braid-go/module/internal/server"
	"github.com/pojol/braid-go/module/linkcache"
	"github.com/pojol/braid-go/module/pubsub"
	"github.com/pojol/braid-go/module/rpc/client"
	"github.com/pojol/braid-go/module/rpc/server"
)

type BraidDepend struct {
	Itracer tracer.ITracer
	CClient *consul.Client
	Logger  *blog.Logger
}

type Depend func(*BraidDepend)

func LoggerDepend(opts ...blog.Option) Depend {
	return func(d *BraidDepend) {
		d.Logger = blog.BuildWithOption(opts...)
	}
}

func RedisDepend(opts ...redis.Option) Depend {
	return func(d *BraidDepend) {

	}
}

func TracerDepend(opts ...tracer.Option) Depend {
	return func(d *BraidDepend) {

	}
}

func ConsulDepend(opts ...consul.Option) Depend {
	return func(d *BraidDepend) {
		d.CClient = consul.BuildWithOption(opts...)
	}
}

////////////////////////////////////////////////////////////////////////////////

type BraidModule struct {
	Iclient client.IClient
	Iserver server.IServer

	Idiscover  discover.IDiscover
	Ielector   elector.IElector
	Ilinkcache linkcache.ILinkCache

	Ipubsub pubsub.IPubsub
}

type Module func(*BraidModule)

func Client(opts ...client.Option) Module {
	return func(c *BraidModule) {

		c.Iclient = iclient.BuildWithOption(
			"",
			c.Ipubsub,
			c.Ilinkcache,
			opts...,
		)

	}
}

func Server(opts ...server.Option) Module {
	return func(c *BraidModule) {
		c.Iserver = iserver.BuildWithOption(
			"",
			opts...,
		)
	}
}

func Discover(opts ...discover.Option) Module {
	return func(c *BraidModule) {
		c.Idiscover = idiscover.BuildWithOption(
			"",
			c.Ipubsub,
			opts...,
		)
	}
}

func LinkCache(opts ...linkcache.Option) Module {
	return func(c *BraidModule) {

	}
}

func Elector(opts ...elector.Option) Module {
	return func(c *BraidModule) {

	}
}

func Pubsub(opts ...pubsub.Option) Module {
	return func(c *BraidModule) {

	}
}
