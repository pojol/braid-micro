package modules

import (
	"github.com/pojol/braid-go/depend"
	"github.com/pojol/braid-go/module"
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

	mods []module.IModule

	IClient    client.IClient
	IServer    server.IServer
	Ilinkcache linkcache.ILinkCache
	Ipubsub    pubsub.IPubsub
}

type Module func(*BraidModule)

func (m *BraidModule) Mods() []module.IModule {
	return m.mods
}

func Client(opts ...client.Option) Module {
	return func(c *BraidModule) {
		c.IClient = clientgrpc.BuildWithOption(
			c.ServiceName,
			c.Depends.Logger,
			c.Ipubsub,
			c.Ilinkcache,
			opts...,
		)
		c.mods = append(c.mods, c.IClient)
	}
}

func Server(opts ...server.Option) Module {
	return func(c *BraidModule) {
		c.IServer = servergrpc.BuildWithOption(
			c.ServiceName,
			c.Depends.Logger,
			opts...,
		)
		c.mods = append(c.mods, c.IServer)
	}
}

func Discover(opts ...discover.Option) Module {
	return func(c *BraidModule) {
		c.mods = append(c.mods, discoverconsul.BuildWithOption(
			c.ServiceName,
			c.Depends.Logger,
			c.Ipubsub,
			c.Depends.ConsulClient,
			opts...,
		))
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
		c.mods = append(c.mods, c.Ilinkcache)
	}
}

func Elector(opts ...elector.Option) Module {
	return func(c *BraidModule) {
		c.mods = append(c.mods, electorconsul.BuildWithOption(
			c.ServiceName,
			c.Depends.Logger,
			c.Ipubsub,
			c.Depends.ConsulClient,
			opts...,
		))
	}
}

func Pubsub(opts ...pubsub.NsqOption) Module {
	return func(c *BraidModule) {
		c.Ipubsub = pubsubnsq.BuildWithOption(
			c.ServiceName,
			c.Depends.Logger,
			opts...,
		)
	}
}
