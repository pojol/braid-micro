package grpcclient

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pojol/braid-go/components/depends/blog"
	"github.com/pojol/braid-go/components/depends/bredis"
	"github.com/pojol/braid-go/components/internal/balancer"
	"github.com/pojol/braid-go/components/pubsubredis"
	"github.com/pojol/braid-go/components/rpcgrpc/grpcserver"
	"github.com/pojol/braid-go/components/rpcgrpc/proto"
	"github.com/pojol/braid-go/mock"
	"github.com/pojol/braid-go/module/meta"
	"github.com/redis/go-redis/v9"
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
	redisclient := bredis.BuildWithOption(&redis.Options{Addr: "127.0.0.1:6379"})

	ps := pubsubredis.BuildWithOption(
		meta.ServiceInfo{ID: "", Name: ""},
		log,
		redisclient,
	)

	grpcserver := grpcserver.BuildWithOption(
		meta.ServiceInfo{
			Name: "servergrpctest",
			ID:   uuid.New().String(),
		},
		log,
		grpcserver.WithListen(":1216"),
		grpcserver.RegisterHandler(func(srv *grpc.Server) {
			proto.RegisterListenServer(srv, &rpcServer{})
		}),
	)

	grpcserver.Init()
	grpcserver.Run()
	defer grpcserver.Close()

	// 伪造一个节点用于测试
	ps.GetTopic(meta.TopicDiscoverServiceUpdate).Pub(context.TODO(),
		meta.EncodeUpdateMsg(
			meta.TopicDiscoverServiceNodeAdd,
			meta.Node{
				ID:      "testnod",
				Name:    "testgrpcclient",
				Address: "http://localhost:1216",
			},
		))

	m.Run()
}

func TestInvoke(t *testing.T) {

	log := blog.BuildWithOption()
	redisclient := bredis.BuildWithOption(&redis.Options{Addr: "127.0.0.1:6379"})

	ps := pubsubredis.BuildWithOption(
		meta.ServiceInfo{ID: "", Name: ""},
		log,
		redisclient,
	)

	rpcc := BuildWithOption(
		meta.ServiceInfo{
			Name: "clientgrpctest",
			ID:   uuid.New().String(),
		},
		log,
		balancer.BuildWithOption(meta.ServiceInfo{}, log, ps),
		nil,
		ps,
	)

	time.Sleep(time.Second)
	res := &proto.RouteRes{}

	tc, cancel := context.WithTimeout(context.TODO(), time.Millisecond*200)
	defer cancel()

	rpcc.Invoke(tc, "testgrpcclient", "/bproto.listen/routing", "", &proto.RouteReq{
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
