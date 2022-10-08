package clientgrpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pojol/braid-go/depend/blog"
	"github.com/pojol/braid-go/mock"
	"github.com/pojol/braid-go/module/discover"
	"github.com/pojol/braid-go/module/internal/pubsubnsq"
	"github.com/pojol/braid-go/module/internal/servergrpc"
	"github.com/pojol/braid-go/module/pubsub"
	"github.com/pojol/braid-go/module/rpc/proto"
	"github.com/pojol/braid-go/module/rpc/server"
	"github.com/pojol/braid-go/service"
	"google.golang.org/grpc"
)

type rpcServer struct {
	proto.ListenServer
}

func (rs *rpcServer) Routing(ctx context.Context, req *proto.RouteReq) (*proto.RouteRes, error) {
	out := new(proto.RouteRes)
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

	log := blog.BuildWithOption()

	ps := pubsubnsq.BuildWithOption("", log, pubsub.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}))

	s := servergrpc.BuildWithOption("", log, server.WithListen(":1216")).(server.IServer)

	s.Init()
	proto.RegisterListenServer(s.Server().(*grpc.Server), &rpcServer{})
	s.Run()

	// 伪造一个节点用于测试
	ps.LocalTopic(discover.TopicServiceUpdate).Pub(discover.EncodeUpdateMsg(
		discover.EventAddService,
		service.Node{
			ID:      "testnod",
			Name:    "testgrpcclient",
			Address: "http://localhost:1216",
			Weight:  100,
		},
	))

	m.Run()
}

func TestInvoke(t *testing.T) {

	log := blog.BuildWithOption()

	ps := pubsubnsq.BuildWithOption("TestInvoke", log, pubsub.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}))

	clientB := BuildWithOption("TestInvoke", log, ps, nil)

	time.Sleep(time.Second)
	res := &proto.RouteRes{}

	tc, cancel := context.WithTimeout(context.TODO(), time.Millisecond*200)
	defer cancel()

	clientB.Invoke(tc, "testgrpcclient", "/bproto.listen/routing", "", &proto.RouteReq{
		Nod:     "testgrpcclient",
		Service: "test",
		ReqBody: []byte{},
	}, res)

}

/*
func TestInvokeByLink(t *testing.T) {

	mbb := module.GetBuilder(pubsubnsq.Name)
	mbb.AddModuleOption(pubsubnsq.WithLookupAddr([]string{}))
	mbb.AddModuleOption(pubsubnsq.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}))
	mb := mbb.Build("TestInvokeByLink").(pubsub.IPubsub)

	lcb := module.GetBuilder(linkerredis.Name)
	lc := lcb.Build("TestInvokeByLink", moduleparm.WithPubsub(mb)).(linkcache.ILinkCache)

	bgb := module.GetBuilder(balancernormal.Name)
	b := bgb.Build("TestInvokeByLink",
		moduleparm.WithPubsub(mb)).(balancer.IBalancer)

	clientB := module.GetBuilder(Name)
	cb := clientB.Build("TestInvokeByLink",
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
*/

/*
func TestParm(t *testing.T) {

	mbb := module.GetBuilder(pubsubnsq.Name)
	mbb.AddModuleOption(pubsubnsq.WithLookupAddr([]string{}))
	mbb.AddModuleOption(pubsubnsq.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}))
	mb := mbb.Build("TestInvokeByLink").(pubsub.IPubsub)

	bgb := module.GetBuilder(balancernormal.Name)
	b := bgb.Build("TestInvokeByLink",
		moduleparm.WithPubsub(mb)).(balancer.IBalancer)

	clientB := module.GetBuilder(Name)
	clientB.AddModuleOption(WithPoolInitNum(100))
	clientB.AddModuleOption(WithPoolIdle(120))
	clientB.AddModuleOption(WithPoolCapacity(101))

	cb := clientB.Build("TestInvokeByLink",
		moduleparm.WithPubsub(mb),
		moduleparm.WithBalancer(b),
	).(*grpcClient)

	assert.Equal(t, cb.parm.PoolInitNum, 100)
	assert.Equal(t, cb.parm.PoolCapacity, 101)
	assert.Equal(t, cb.parm.PoolIdle, time.Second*120)

}

*/
