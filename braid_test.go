package braid

import (
	"testing"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/pojol/braid-go/depend/blog"
	"github.com/pojol/braid-go/depend/pubsub"
	"github.com/pojol/braid-go/depend/redis"
	"github.com/pojol/braid-go/depend/tracer"
	"github.com/pojol/braid-go/mock"
	"github.com/pojol/braid-go/module/elector"
	"github.com/pojol/braid-go/module/linkcache"
	"github.com/pojol/braid-go/rpc/client"
	"github.com/pojol/braid-go/rpc/server"
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
		blog.BuildWithNormal(),
		redis.BuildWithDefault(),
		pubsub.BuildWithOption(
			b.name,
			pubsub.WithLookupAddr([]string{mock.NSQLookupdAddr}),
			pubsub.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}),
		),
		tracer.BuildWithOption(
			b.name,
			tracer.WithHTTP(mock.JaegerAddr),
			tracer.WithProbabilistic(1),
		),
	)

	b.RegisterClient(
		client.AppendInterceptors(grpc_prometheus.UnaryClientInterceptor),
	)
	b.RegisterServer(
		server.WithListen(":14222"),
		server.AppendInterceptors(grpc_prometheus.UnaryServerInterceptor),
	)

	b.RegisterModule(
		b.Discover(),
		b.Elector(elector.WithLockTick(3*time.Second)),
		b.LinkCache(linkcache.WithMode(linkcache.LinkerRedisModeLocal)),
	)

	b.Init()
	b.Run()
	defer b.Close()
}
