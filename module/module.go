package module

import (
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
			"",
			c.Ipubsub,
			c.Ilinkcache,
			opts...,
		)

	}
}

func Server(opts ...server.Option) Module {
	return func(c *BraidModule) {
		c.Iserver = servergrpc.BuildWithOption(
			"",
			opts...,
		)
	}
}

func Discover(opts ...discover.Option) Module {
	return func(c *BraidModule) {
		c.Idiscover = discoverconsul.BuildWithOption(
			"",
			c.Ipubsub,
			opts...,
		)
	}
}

func LinkCache(opts ...linkcache.Option) Module {
	return func(c *BraidModule) {
		c.Ilinkcache = linkcacheredis.BuildWithOption(
			"",
			opts...,
		)
	}
}

func Elector(opts ...elector.Option) Module {
	return func(c *BraidModule) {
		c.Ielector = electorconsul.BuildWithOption(
			"",
			opts...,
		)
	}
}

func Pubsub(opts ...pubsub.Option) Module {
	return func(c *BraidModule) {
		c.Ipubsub = pubsubnsq.BuildWithOption(
			"",
			opts...,
		)
	}
}
