package braid

import (
	"context"
	"testing"
	"time"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/mock"
	"github.com/pojol/braid/plugin/discover/consuldiscover"
	"github.com/pojol/braid/plugin/rpc/grpcclient"
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

func TestWithClient(t *testing.T) {

	b := New("test")
	b.RegistPlugin(DiscoverByConsul(mock.ConsulAddr, consuldiscover.WithInterval(time.Second*3)),
		BalancerBySwrr(),
		RPCClient(grpcclient.WithPoolCapacity(128)))

	b.Run()
	defer b.Close()

	Client().Invoke(context.TODO(), "targeNodeName", "/proto.node/method", "", nil, nil)
}
