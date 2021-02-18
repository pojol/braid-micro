package grpcclient

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pojol/braid-go/mock"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/discover"
	"github.com/pojol/braid-go/module/linkcache"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/mailbox"
	"github.com/pojol/braid-go/module/rpc/client"
	"github.com/pojol/braid-go/module/rpc/server"
	"github.com/pojol/braid-go/modules/discoverconsul"
	"github.com/pojol/braid-go/modules/grpcclient/bproto"
	"github.com/pojol/braid-go/modules/grpcserver"
	"github.com/pojol/braid-go/modules/linkerredis"
	"github.com/pojol/braid-go/modules/mailboxnsq"
	"github.com/pojol/braid-go/modules/zaplogger"
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

	log, _ := logger.GetBuilder(zaplogger.Name).Build()

	db := module.GetBuilder(discoverconsul.Name)
	db.AddOption(discoverconsul.WithConsulAddr(mock.ConsulAddr))

	sb := server.GetBuilder(grpcserver.Name)
	sb.AddOption(grpcserver.WithListen(":1216"))
	s, _ := sb.Build("testgrpcclient", log)
	s.Init()
	bproto.RegisterListenServer(s.Server().(*grpc.Server), &rpcServer{})
	s.Run()

	// 伪造一个节点用于测试
	mb.Pub(mailbox.Proc, discover.AddService, mailbox.NewMessage(discover.Node{
		ID:      "testnod",
		Name:    "testgrpcclient",
		Address: "http://localhost:1216",
		Weight:  100,
	}))

	m.Run()
}

func TestInvoke(t *testing.T) {
	mbb := mailbox.GetBuilder(mailboxnsq.Name)
	mbb.AddOption(mailboxnsq.WithLookupAddr([]string{mock.NSQLookupdAddr}))
	mbb.AddOption(mailboxnsq.WithNsqdAddr([]string{mock.NsqdAddr}))
	mb, _ := mbb.Build("TestInvoke")

	b := client.GetBuilder(Name)
	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	cb, _ := b.Build("TestInvoke", mb, log)

	cb.Init()
	cb.Run()
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
	log, _ := logger.GetBuilder(zaplogger.Name).Build()

	mbb := mailbox.GetBuilder(mailboxnsq.Name)
	mbb.AddOption(mailboxnsq.WithLookupAddr([]string{mock.NSQLookupdAddr}))
	mbb.AddOption(mailboxnsq.WithNsqdAddr([]string{mock.NsqdAddr}))
	mb, _ := mbb.Build("TestInvokeByLink")

	lb := module.GetBuilder(linkerredis.Name)
	lb.AddOption(linkerredis.WithRedisAddr(mock.RedisAddr))
	lc, _ := lb.Build("TestInvokeByLink", mb, log)

	b.AddOption(AutoLinkCache(lc.(linkcache.ILinkCache)))
	//b.AddOption(Tracing())
	cb, _ := b.Build("TestInvokeByLink", mb, log)

	cb.Init()
	cb.Run()
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

	mbb := mailbox.GetBuilder(mailboxnsq.Name)
	mbb.AddOption(mailboxnsq.WithLookupAddr([]string{mock.NSQLookupdAddr}))
	mbb.AddOption(mailboxnsq.WithNsqdAddr([]string{mock.NsqdAddr}))
	mb, _ := mbb.Build("TestParm")

	b := client.GetBuilder(Name)
	b.AddOption(WithPoolInitNum(100))
	b.AddOption(WithPoolCapacity(101))
	b.AddOption(WithPoolIdle(120))
	//b.AddOption(Tracing())
	b.AddOption(AutoLinkCache(nil))

	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	cb, _ := b.Build("TestCaller", mb, log)
	gc := cb.(*grpcClient)

	assert.Equal(t, gc.parm.PoolInitNum, 100)
	assert.Equal(t, gc.parm.PoolCapacity, 101)
	assert.Equal(t, gc.parm.PoolIdle, time.Second*120)
	//assert.Equal(t, gc.parm.isTracing, true)
	assert.Equal(t, gc.parm.byLink, true)

}
