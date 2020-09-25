package grpcclient

import (
	"context"
	"testing"
	"time"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/mock"
	"github.com/pojol/braid/module/balancer"
	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/module/pubsub"
	"github.com/pojol/braid/module/rpc/client"
	"github.com/pojol/braid/plugin/balancerswrr"
	"github.com/pojol/braid/plugin/discoverconsul"
	"github.com/pojol/braid/plugin/grpcclient/bproto"
	"github.com/pojol/braid/plugin/pubsubproc"
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

	db := discover.GetBuilder(discoverconsul.Name)

	ps, _ := pubsub.GetBuilder(pubsubproc.PubsubName).Build("TestMain")
	db.AddOption(discoverconsul.WithProcPubsub(ps))
	db.AddOption(discoverconsul.WithConsulAddr(mock.ConsulAddr))

	discv, err := db.Build("test")
	if err != nil {
		panic(err)
	}

	discv.Discover()
	defer discv.Close()

	bb := balancer.GetBuilder(balancerswrr.Name)
	bb.AddOption(balancerswrr.WithProcPubsub(ps))
	balancer.NewGroup(bb)

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
