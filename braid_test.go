package braid

import (
	"testing"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/pojol/braid-go/depend/pubsub"
	"github.com/pojol/braid-go/depend/tracer"
	"github.com/pojol/braid-go/mock"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/elector"
	"github.com/pojol/braid-go/module/linkcache"
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
		module.LoggerDepend(),
		module.RedisDepend(),
		module.TracerDepend(
			tracer.WithHTTP(mock.JaegerAddr),
			tracer.WithProbabilistic(1),
		),
	)

	b.RegisterModule(
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
		module.Pubsub(
			pubsub.WithLookupAddr([]string{mock.NSQLookupdAddr}),
			pubsub.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}),
		),
	)

	b.Init()
	b.Run()
	defer b.Close()
}
