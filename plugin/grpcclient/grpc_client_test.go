package grpcclient

import (
	"context"
	"testing"
	"time"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/mock"
	"github.com/pojol/braid/module"
	"github.com/pojol/braid/module/balancer"
	"github.com/pojol/braid/module/mailbox"
	"github.com/pojol/braid/module/rpc/client"
	"github.com/pojol/braid/plugin/balancerswrr"
	"github.com/pojol/braid/plugin/discoverconsul"
	"github.com/pojol/braid/plugin/grpcclient/bproto"
	"github.com/pojol/braid/plugin/mailboxnsq"
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

	mbb := mailbox.GetBuilder(mailboxnsq.Name)
	mbb.AddOption(mailboxnsq.WithLookupAddr([]string{mock.NSQLookupdAddr}))
	mbb.AddOption(mailboxnsq.WithNsqdAddr([]string{mock.NsqdAddr}))
	mb, err := mbb.Build("TestMain")

	db := module.GetBuilder(discoverconsul.Name)
	db.AddOption(discoverconsul.WithConsulAddr(mock.ConsulAddr))

	discv, err := db.Build("test", mb)
	if err != nil {
		panic(err)
	}

	discv.Run()
	defer discv.Close()

	bb := module.GetBuilder(balancerswrr.Name)
	balancer.NewGroup(bb, mb)

	m.Run()
}

func TestCaller(t *testing.T) {
	b := client.GetBuilder(Name)
	cb, _ := b.Build("TestCaller")

	time.Sleep(time.Millisecond * 200)
	res := &bproto.RouteRes{}

	nodeName := "gateway"
	serviceName := "service"

	tc, cancel := context.WithTimeout(context.TODO(), time.Millisecond*200)
	defer cancel()

	cb.Invoke(tc, nodeName, "/bproto.listen/routing", "", &bproto.RouteReq{
		Nod:     nodeName,
		Service: serviceName,
		ReqBody: []byte{},
	}, res)

}
