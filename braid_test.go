package braid

import (
	"testing"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/pojol/braid-go/depend"
	"github.com/pojol/braid-go/depend/bconsul"
	"github.com/pojol/braid-go/depend/blog"
	"github.com/pojol/braid-go/depend/bredis"
	"github.com/pojol/braid-go/depend/btracer"
	"github.com/pojol/braid-go/mock"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/elector"
	"github.com/pojol/braid-go/module/linkcache"
	"github.com/pojol/braid-go/module/pubsub"
	"github.com/pojol/braid-go/module/rpc/client"
	"github.com/pojol/braid-go/module/rpc/server"
)

func TestMain(m *testing.M) {

	mock.Init()
	m.Run()

}

func TestInit(t *testing.T) {

	b, _ := NewService(
		"test_init",
	)

	b.RegisterDepend(
		depend.Logger(blog.BuildWithOption()),
		depend.Redis(bredis.BuildWithOption(bredis.WithAddr(mock.RedisAddr))),
		depend.Tracer(
			btracer.WithHTTP(mock.JaegerAddr),
			btracer.WithProbabilistic(1),
		),
		depend.Consul(
			bconsul.BuildWithOption(bconsul.WithAddress([]string{mock.ConsulAddr})),
		),
	)

	b.RegisterModule(
		module.Pubsub(
			pubsub.WithLookupAddr([]string{mock.NSQLookupdAddr}),
			pubsub.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}),
		),
		module.Client(
			client.AppendInterceptors(grpc_prometheus.UnaryClientInterceptor),
		),
		module.Server(
			server.WithListen(":14222"),
			server.AppendInterceptors(grpc_prometheus.UnaryServerInterceptor),
		),
		module.Discover(),
		module.Elector(
			elector.WithLockTick(3*time.Second)),
		module.LinkCache(
			linkcache.WithMode(linkcache.LinkerRedisModeLocal),
		),
	)

	b.Init()
	b.Run()
	defer b.Close()
}
