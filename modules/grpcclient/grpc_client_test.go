package grpcclient

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pojol/braid-go/mock"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/balancer"
	"github.com/pojol/braid-go/module/discover"
	"github.com/pojol/braid-go/module/linkcache"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/pubsub"
	"github.com/pojol/braid-go/module/rpc/client"
	"github.com/pojol/braid-go/module/rpc/server"
	"github.com/pojol/braid-go/modules/balancernormal"
	"github.com/pojol/braid-go/modules/discoverconsul"
	"github.com/pojol/braid-go/modules/grpcclient/bproto"
	"github.com/pojol/braid-go/modules/grpcserver"
	"github.com/pojol/braid-go/modules/linkerredis"
	"github.com/pojol/braid-go/modules/moduleparm"
	"github.com/pojol/braid-go/modules/pubsubnsq"
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

	log := module.GetBuilder(zaplogger.Name).Build("TestProcNotify").(logger.ILogger)
	mbb := module.GetBuilder(pubsubnsq.Name)
	mbb.AddModuleOption(pubsubnsq.WithLookupAddr([]string{}))
	mbb.AddModuleOption(pubsubnsq.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}))
	mb := mbb.Build("TestMain", moduleparm.WithLogger(log)).(pubsub.IPubsub)

	discoverB := module.GetBuilder(discoverconsul.Name)
	discoverB.AddModuleOption(discoverconsul.WithConsulAddr(mock.ConsulAddr))

	sb := module.GetBuilder(grpcserver.Name)
	sb.AddModuleOption(grpcserver.WithListen(":1216"))
	s := sb.Build("testgrpcclient", moduleparm.WithLogger(log)).(server.IServer)

	s.Init()
	bproto.RegisterListenServer(s.Server().(*grpc.Server), &rpcServer{})
	s.Run()

	// 伪造一个节点用于测试
	mb.GetTopic(discover.ServiceUpdate).Pub(discover.EncodeUpdateMsg(
		discover.EventAddService,
		discover.Node{
			ID:      "testnod",
			Name:    "testgrpcclient",
			Address: "http://localhost:1216",
			Weight:  100,
		},
	))

	m.Run()
}

func TestInvoke(t *testing.T) {

	log := module.GetBuilder(zaplogger.Name).Build("TestInvoke").(logger.ILogger)
	mbb := module.GetBuilder(pubsubnsq.Name)
	mbb.AddModuleOption(pubsubnsq.WithLookupAddr([]string{}))
	mbb.AddModuleOption(pubsubnsq.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}))
	mb := mbb.Build("TestInvoke", moduleparm.WithLogger(log)).(pubsub.IPubsub)

	bgb := module.GetBuilder(balancernormal.Name)
	b := bgb.Build("TestInvoke",
		moduleparm.WithLogger(log),
		moduleparm.WithPubsub(mb)).(balancer.IBalancer)

	clientB := module.GetBuilder(Name)
	cb := clientB.Build("TestInvoke",
		moduleparm.WithLogger(log),
		moduleparm.WithPubsub(mb),
		moduleparm.WithBalancer(b),
	).(client.IClient)

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

	log := module.GetBuilder(zaplogger.Name).Build("TestInvokeByLink").(logger.ILogger)
	mbb := module.GetBuilder(pubsubnsq.Name)
	mbb.AddModuleOption(pubsubnsq.WithLookupAddr([]string{}))
	mbb.AddModuleOption(pubsubnsq.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}))
	mb := mbb.Build("TestInvokeByLink", moduleparm.WithLogger(log)).(pubsub.IPubsub)

	lcb := module.GetBuilder(linkerredis.Name)
	lc := lcb.Build("TestInvokeByLink", moduleparm.WithPubsub(mb), moduleparm.WithLogger(log)).(linkcache.ILinkCache)

	bgb := module.GetBuilder(balancernormal.Name)
	b := bgb.Build("TestInvokeByLink",
		moduleparm.WithLogger(log),
		moduleparm.WithPubsub(mb)).(balancer.IBalancer)

	clientB := module.GetBuilder(Name)
	cb := clientB.Build("TestInvokeByLink",
		moduleparm.WithLogger(log),
		moduleparm.WithPubsub(mb),
		moduleparm.WithBalancer(b),
		moduleparm.WithLinkcache(lc),
	).(client.IClient)

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

	log := module.GetBuilder(zaplogger.Name).Build("TestInvokeByLink").(logger.ILogger)
	mbb := module.GetBuilder(pubsubnsq.Name)
	mbb.AddModuleOption(pubsubnsq.WithLookupAddr([]string{}))
	mbb.AddModuleOption(pubsubnsq.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}))
	mb := mbb.Build("TestInvokeByLink", moduleparm.WithLogger(log)).(pubsub.IPubsub)

	bgb := module.GetBuilder(balancernormal.Name)
	b := bgb.Build("TestInvokeByLink",
		moduleparm.WithLogger(log),
		moduleparm.WithPubsub(mb)).(balancer.IBalancer)

	clientB := module.GetBuilder(Name)
	clientB.AddModuleOption(WithPoolInitNum(100))
	clientB.AddModuleOption(WithPoolIdle(120))
	clientB.AddModuleOption(WithPoolCapacity(101))

	cb := clientB.Build("TestInvokeByLink",
		moduleparm.WithLogger(log),
		moduleparm.WithPubsub(mb),
		moduleparm.WithBalancer(b),
	).(*grpcClient)

	assert.Equal(t, cb.parm.PoolInitNum, 100)
	assert.Equal(t, cb.parm.PoolCapacity, 101)
	assert.Equal(t, cb.parm.PoolIdle, time.Second*120)

}
