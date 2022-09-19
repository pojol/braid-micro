package module

import (
	"github.com/pojol/braid-go/depend"
	"github.com/pojol/braid-go/module/discover"
	"github.com/pojol/braid-go/module/elector"
	"github.com/pojol/braid-go/module/internal/clientgrpc"
	"github.com/pojol/braid-go/module/internal/discoverconsul"
	"github.com/pojol/braid-go/module/internal/electorconsul"
	"github.com/pojol/braid-go/module/internal/linkcacheredis"
	"github.com/pojol/braid-go/module/internal/pubsubnsq"
	"github.com/pojol/braid-go/module/internal/servergrpc"
	"github.com/pojol/braid-go/module/linkcache"
	"github.com/pojol/braid-go/module/pubsub"
	"github.com/pojol/braid-go/module/rpc/client"
	"github.com/pojol/braid-go/module/rpc/server"
)

type BraidModule struct {
	ServiceName string

	Depends *depend.BraidDepend

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

		c.Iclient = clientgrpc.BuildWithOption(
			c.ServiceName,
			c.Depends.Logger,
			c.Ipubsub,
			c.Ilinkcache,
			opts...,
		)

	}
}

func Server(opts ...server.Option) Module {
	return func(c *BraidModule) {
		c.Iserver = servergrpc.BuildWithOption(
			c.ServiceName,
			c.Depends.Logger,
			opts...,
		)
	}
}

func Discover(opts ...discover.Option) Module {
	return func(c *BraidModule) {
		c.Idiscover = discoverconsul.BuildWithOption(
			c.ServiceName,
			c.Depends.Logger,
			c.Ipubsub,
			c.Depends.ConsulClient,
			opts...,
		)
	}
}

func LinkCache(opts ...linkcache.Option) Module {
	return func(c *BraidModule) {
		c.Ilinkcache = linkcacheredis.BuildWithOption(
			c.ServiceName,
			c.Depends.Logger,
			c.Ipubsub,
			c.Depends.RedisClient,
			opts...,
		)
	}
}

func Elector(opts ...elector.Option) Module {
	return func(c *BraidModule) {
		c.Ielector = electorconsul.BuildWithOption(
			c.ServiceName,
			c.Depends.Logger,
			c.Ipubsub,
			c.Depends.ConsulClient,
			opts...,
		)
	}
}

func Pubsub(opts ...pubsub.Option) Module {
	return func(c *BraidModule) {
		c.Ipubsub = pubsubnsq.BuildWithOption(
			c.ServiceName,
			c.Depends.Logger,
			opts...,
		)
	}
}
