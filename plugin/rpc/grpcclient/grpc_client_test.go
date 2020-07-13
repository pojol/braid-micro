package grpcclient

import (
	"context"
	"testing"
	"time"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/mock"
	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/module/rpc/client"
	"github.com/pojol/braid/plugin/balancer"
	"github.com/pojol/braid/plugin/balancer/swrrbalancer"
	"github.com/pojol/braid/plugin/discover/consuldiscover"
	"github.com/pojol/braid/plugin/rpc/grpcclient/bproto"
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

	db := discover.GetBuilder(consuldiscover.DiscoverName)
	db.SetCfg(consuldiscover.Cfg{
		Name:     "test",
		Tag:      "braid",
		Interval: time.Second * 2,
		Address:  mock.ConsulAddr,
	})
	discv := db.Build()
	discv.Discover()
	defer discv.Close()

	balancer.NewGroup(balancer.GetBuilder(swrrbalancer.BalancerName))

	m.Run()
}

func TestCaller(t *testing.T) {

	b := client.GetBuilder(ClientName)
	b.SetCfg(Config{
		Name:         "test",
		PoolInitNum:  8,
		PoolCapacity: 32,
		PoolIdle:     120,
	})
	c := b.Build(nil, false)

	time.Sleep(time.Millisecond * 200)
	res := &bproto.RouteRes{}

	nodeName := "gateway"
	serviceName := "service"

	tc, cancel := context.WithTimeout(context.TODO(), time.Millisecond*200)
	defer cancel()

	c.Invoke(tc, nodeName, "/bproto.listen/routing", "", &bproto.RouteReq{
		Nod:     nodeName,
		Service: serviceName,
		ReqBody: []byte{},
	}, res)

}
