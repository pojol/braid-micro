package components

import (
	"fmt"

	"github.com/pojol/braid-go/components/depends/bconsul"
	"github.com/pojol/braid-go/components/depends/bk8s"
	"github.com/pojol/braid-go/components/depends/blog"
	"github.com/pojol/braid-go/components/depends/bredis"
	"github.com/pojol/braid-go/components/discoverk8s"
	"github.com/pojol/braid-go/components/electork8s"
	"github.com/pojol/braid-go/components/internal/balancer"
	"github.com/pojol/braid-go/components/linkcacheredis"
	"github.com/pojol/braid-go/components/monitorredis"
	"github.com/pojol/braid-go/components/pubsubredis"
	"github.com/pojol/braid-go/components/rpcgrpc/grpcclient"
	"github.com/pojol/braid-go/components/rpcgrpc/grpcserver"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/meta"
	"github.com/redis/go-redis/v9"
)

type IDirector interface {
	Build() error

	SetServiceInfo(info meta.ServiceInfo)

	Init() error
	Run()
	Close()

	Logger() *blog.Logger

	Pubsub() module.IPubsub
	Client() module.IClient
}

type DirectorOpts struct {
	LogOpts       []blog.Option
	RedisCliOpts  *redis.Options
	ConsulCliOpts []bconsul.Option
	K8sCliOpts    []bk8s.Option

	ClientOpts    []grpcclient.Option
	ServerOpts    []grpcserver.Option
	ElectorOpts   []electork8s.Option
	LinkcacheOpts []linkcacheredis.Option
	DiscoverOpts  []discoverk8s.Option
}

type DefaultDirector struct {
	Opts *DirectorOpts

	info meta.ServiceInfo

	log *blog.Logger

	monitor module.IMonitor

	client module.IClient
	server module.IServer

	balancer balancer.IBalancer

	discovery module.IDiscover
	linkcache module.ILinkCache
	pubsub    module.IPubsub
	elector   module.IElector
}

type Component func(*DefaultDirector)

func WithLog(log *blog.Logger) Component {
	return func(c *DefaultDirector) {
		c.log = log
	}
}

func WithClient(client module.IClient) Component {
	return func(c *DefaultDirector) {
		c.client = client
	}
}

func (d *DefaultDirector) SetServiceInfo(info meta.ServiceInfo) {
	d.info = info
}

func (d *DefaultDirector) Build() error {

	d.log = blog.BuildWithOption(d.Opts.LogOpts...)
	var rediscli *redis.Client
	var k8scli *bk8s.Client

	if d.Opts.RedisCliOpts != nil {
		rediscli = bredis.BuildWithOption(d.Opts.RedisCliOpts)
	} else {
		rediscli = bredis.BuildWithDefault()
	}

	if len(d.Opts.K8sCliOpts) != 0 {
		k8scli = bk8s.BuildWithOption(d.Opts.K8sCliOpts...)
	} else {
		k8scli = bk8s.BuildWithOption(bk8s.WithConfigPath(""))
	}

	ps := pubsubredis.BuildWithOption(d.info, d.log, rediscli)

	discover := discoverk8s.BuildWithOption(
		d.info,
		d.log,
		k8scli,
		ps,
		d.Opts.DiscoverOpts...,
	)

	lc := linkcacheredis.BuildWithOption(d.info, d.log, ps, rediscli, d.Opts.LinkcacheOpts...)

	elector := electork8s.BuildWithOption(d.info, d.log, ps, k8scli, d.Opts.ElectorOpts...)

	// tmp
	e := monitorredis.BuildWithOption(d.log, rediscli)

	d.pubsub = ps
	d.linkcache = lc
	d.discovery = discover
	d.elector = elector
	d.balancer = balancer.BuildWithOption(d.info, d.log, ps)
	d.monitor = e

	d.client = grpcclient.BuildWithOption(
		d.info,
		d.log,
		d.balancer,
		lc,
		ps,
		d.Opts.ClientOpts...,
	)

	if len(d.Opts.ServerOpts) != 0 {
		d.server = grpcserver.BuildWithOption(d.info, d.log, d.Opts.ServerOpts...)
	}

	return nil
}

func (d *DefaultDirector) Init() error {

	var err error

	if d.server != nil {
		err = d.server.Init()
		if err != nil {
			return fmt.Errorf("server init err : %w", err)
		}
	}

	if d.client != nil {
		err = d.client.Init()
		if err != nil {
			return fmt.Errorf("client init err : %w", err)
		}
	}

	if d.discovery != nil {
		err = d.discovery.Init()
		if err != nil {
			return fmt.Errorf("discovery init err : %w", err)
		}
	}

	if d.elector != nil {
		err = d.elector.Init()
		if err != nil {
			return fmt.Errorf("elector init err : %w", err)
		}
	}

	if d.linkcache != nil {
		err = d.linkcache.Init()
		if err != nil {
			return fmt.Errorf("linkcache init err : %w", err)
		}
	}

	return nil
}

func (d *DefaultDirector) Run() {

	d.monitor.Run()

	if d.server != nil {
		d.server.Run()
	}

	if d.discovery != nil {
		d.discovery.Run()
	}

	if d.elector != nil {
		d.elector.Run()
	}

	if d.linkcache != nil {
		d.linkcache.Run()
	}
}

func (d *DefaultDirector) Close() {

	d.balancer.Close()

	if d.server != nil {
		d.server.Close()
	}

	if d.discovery != nil {
		d.discovery.Close()
	}

	if d.elector != nil {
		d.elector.Close()
	}

	if d.linkcache != nil {
		d.linkcache.Close()
	}

	if d.client != nil {
		d.client.Close()
	}
}

func (d *DefaultDirector) Logger() *blog.Logger {
	return d.log
}

func (d *DefaultDirector) Client() module.IClient {
	return d.client
}

func (d *DefaultDirector) Pubsub() module.IPubsub {
	return d.pubsub
}
