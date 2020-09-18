package braid

import (
	"testing"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/mock"
	"github.com/pojol/braid/plugin/balancerswrr"
	"github.com/pojol/braid/plugin/discoverconsul"
)

func TestMain(m *testing.M) {

	mock.Init()
	l := log.New(log.Config{
		Mode:   log.DebugMode,
		Path:   "testNormal",
		Suffex: ".log",
	}, log.WithSys(log.Config{
		Mode:   log.DebugMode,
		Path:   "testSys",
		Suffex: ".sys",
	}))
	defer l.Close()

	m.Run()
}

func TestPlugin(t *testing.T) {

	b := New("testPlugin")
	b.RegistPlugin(
		Balancer(balancerswrr.Name),
		Discover(
			discoverconsul.Name,
			discoverconsul.WithConsulAddress(mock.ConsulAddr),
		),
	)

	b.Run()
	defer b.Close()
}

func TestWithClient(t *testing.T) {
	/*
		b := New("test")
		b.RegistPlugin(DiscoverByConsul(mock.ConsulAddr, discoverconsul.WithInterval(time.Second*3)),
			BalancerBySwrr(),
			GRPCClient(grpcclient.WithPoolCapacity(128)))

		b.Run()
		defer b.Close()

		Client().Invoke(context.TODO(), "targeNodeName", "/proto.node/method", "", nil, nil)
	*/
}
