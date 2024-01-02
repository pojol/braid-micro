package braid

import (
	"testing"
	"time"

	"github.com/google/uuid"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/pojol/braid-go/components"
	"github.com/pojol/braid-go/components/electork8s"
	"github.com/pojol/braid-go/components/linkcacheredis"
	"github.com/pojol/braid-go/components/rpcgrpc/grpcclient"
	"github.com/pojol/braid-go/components/rpcgrpc/grpcserver"
	"github.com/pojol/braid-go/mock"
	"google.golang.org/grpc"
)

func TestMain(m *testing.M) {

	mock.Init()
	m.Run()

}

func TestInit(t *testing.T) {

	b, _ := NewService(
		"test_init",
		uuid.New().String(),
		&components.DefaultDirector{
			Opts: &components.DirectorOpts{
				ClientOpts: []grpcclient.Option{
					grpcclient.AppendUnaryInterceptors(grpc_prometheus.UnaryClientInterceptor),
				},
				ServerOpts: []grpcserver.Option{
					grpcserver.WithListen(":14222"),
					grpcserver.AppendUnaryInterceptors(grpc_prometheus.UnaryServerInterceptor),
					grpcserver.RegisterHandler(func(srv *grpc.Server) {
						// register grpc handler
					}),
				},
				ElectorOpts: []electork8s.Option{
					electork8s.WithRefreshTick(time.Second * 5),
				},
				LinkcacheOpts: []linkcacheredis.Option{
					linkcacheredis.WithMode(linkcacheredis.LinkerRedisModeLocal),
				},
			},
		},
	)

	b.Init()
	b.Run()
	b.Close()
}
