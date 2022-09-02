package module

import (
	"github.com/pojol/braid-go/depend/consul"
	"github.com/pojol/braid-go/depend/pubsub"
	"github.com/pojol/braid-go/depend/redis"
	"github.com/pojol/braid-go/depend/tracer"
	"github.com/pojol/braid-go/module/discover"
	"github.com/pojol/braid-go/module/elector"
	iclient "github.com/pojol/braid-go/module/internal/client"
	idiscover "github.com/pojol/braid-go/module/internal/discover"
	iserver "github.com/pojol/braid-go/module/internal/server"
	"github.com/pojol/braid-go/module/linkcache"
	"github.com/pojol/braid-go/module/rpc/client"
	"github.com/pojol/braid-go/module/rpc/server"
)

type BraidDepend struct {
	Itracer tracer.ITracer
	CClient *consul.Client
}

type Depend func(*BraidDepend)

func LoggerDepend() Depend {
	return func(d *BraidDepend) {

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

func ConsulDepend() Depend {
	return func(d *BraidDepend) {

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
		c.Idiscover = idiscover.Build(
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
