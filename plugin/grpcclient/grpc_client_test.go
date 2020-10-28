package grpcclient

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pojol/braid/mock"
	"github.com/pojol/braid/module"
	"github.com/pojol/braid/module/balancer"
	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/module/linkcache"
	"github.com/pojol/braid/module/logger"
	"github.com/pojol/braid/module/mailbox"
	"github.com/pojol/braid/module/rpc/client"
	"github.com/pojol/braid/module/rpc/server"
	"github.com/pojol/braid/plugin/balancerswrr"
	"github.com/pojol/braid/plugin/discoverconsul"
	"github.com/pojol/braid/plugin/grpcclient/bproto"
	"github.com/pojol/braid/plugin/grpcserver"
	"github.com/pojol/braid/plugin/linkerredis"
	"github.com/pojol/braid/plugin/mailboxnsq"
	"github.com/pojol/braid/plugin/zaplogger"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type rpcServer struct {
	bproto.ListenServer
}

func (rs *rpcServer) Routing(ctx context.Context, req *bproto.RouteReq) (*bproto.RouteRes, error) {
	out := new(bproto.RouteRes)
	var err error

	if req.Service == "test" {
		err = nil
	} else {
		err = errors.New("err")
	}

	return out, err
}

func TestMain(m *testing.M) {
	mock.Init()

	mbb := mailbox.GetBuilder(mailboxnsq.Name)
	mbb.AddOption(mailboxnsq.WithLookupAddr([]string{mock.NSQLookupdAddr}))
	mbb.AddOption(mailboxnsq.WithNsqdAddr([]string{mock.NsqdAddr}))
	mb, _ := mbb.Build("TestMain")

	log, _ := logger.GetBuilder(zaplogger.Name).Build(logger.DEBUG)

	db := module.GetBuilder(discoverconsul.Name)
	db.AddOption(discoverconsul.WithConsulAddr(mock.ConsulAddr))

	bb := module.GetBuilder(balancerswrr.Name)
	balancer.NewGroup(bb, mb, log)
	balancer.Get("testgrpcclient")

	sb := server.GetBuilder(grpcserver.Name)
	sb.AddOption(grpcserver.WithListen(":1216"))
	s, _ := sb.Build("testgrpcclient", log)
	bproto.RegisterListenServer(s.Server().(*grpc.Server), &rpcServer{})
	s.Run()

	// 伪造一个节点用于测试
	mb.ProcPub(discover.AddService, mailbox.NewMessage(discover.Node{
		ID:      "testnod",
		Name:    "testgrpcclient",
		Address: "http://localhost:1216",
		Weight:  100,
	}))

	m.Run()
}

func TestInvoke(t *testing.T) {
	b := client.GetBuilder(Name)
	log, _ := logger.GetBuilder(zaplogger.Name).Build(logger.DEBUG)
	cb, _ := b.Build("TestInvoke", log)
	defer cb.Close()

	time.Sleep(time.Second)
	res := &bproto.RouteRes{}

	tc, cancel := context.WithTimeout(context.TODO(), time.Millisecond*200)
	defer cancel()

	cb.Invoke(tc, "testgrpcclient", "/bproto.listen/routing", "", &bproto.RouteReq{
		Nod:     "testgrpcclient",
		Service: "test",
		ReqBody: []byte{},
	}, res)

}

func TestInvokeByLink(t *testing.T) {
	b := client.GetBuilder(Name)
	log, _ := logger.GetBuilder(zaplogger.Name).Build(logger.DEBUG)

	mbb := mailbox.GetBuilder(mailboxnsq.Name)
	mbb.AddOption(mailboxnsq.WithLookupAddr([]string{mock.NSQLookupdAddr}))
	mbb.AddOption(mailboxnsq.WithNsqdAddr([]string{mock.NsqdAddr}))
	mb, _ := mbb.Build("TestInvokeByLink")

	lb := module.GetBuilder(linkerredis.Name)
	lb.AddOption(linkerredis.WithRedisAddr(mock.RedisAddr))
	lc, _ := lb.Build("TestInvokeByLink", mb, log)

	b.AddOption(LinkCache(lc.(linkcache.ILinkCache)))
	b.AddOption(Tracing())
	cb, _ := b.Build("TestInvokeByLink", log)
	defer cb.Close()

	time.Sleep(time.Second)
	res := &bproto.RouteRes{}

	tc, cancel := context.WithTimeout(context.TODO(), time.Millisecond*200)
	defer cancel()

	cb.Invoke(tc, "testgrpcclient", "/bproto.listen/routing", "", &bproto.RouteReq{
		Nod:     "testgrpcclient",
		Service: "test",
		ReqBody: []byte{},
	}, res)

	cb.Invoke(tc, "testgrpcclient", "/bproto.listen/routing", "testtoken", &bproto.RouteReq{
		Nod:     "testgrpcclient",
		Service: "test",
		ReqBody: []byte{},
	}, res)
}

func TestParm(t *testing.T) {
	b := client.GetBuilder(Name)
	b.AddOption(WithPoolInitNum(100))
	b.AddOption(WithPoolCapacity(101))
	b.AddOption(WithPoolIdle(120))
	b.AddOption(Tracing())
	b.AddOption(LinkCache(nil))

	log, _ := logger.GetBuilder(zaplogger.Name).Build(logger.DEBUG)
	cb, _ := b.Build("TestCaller", log)
	gc := cb.(*grpcClient)

	assert.Equal(t, gc.parm.PoolInitNum, 100)
	assert.Equal(t, gc.parm.PoolCapacity, 101)
	assert.Equal(t, gc.parm.PoolIdle, time.Second*120)
	assert.Equal(t, gc.parm.isTracing, true)
	assert.Equal(t, gc.parm.byLink, true)

}
