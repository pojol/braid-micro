package server

import (
	"context"
	"errors"
	"testing"

	"github.com/pojol/braid-go/module/rpc/proto"
	"github.com/pojol/braid-go/module/rpc/server"
	"github.com/stretchr/testify/assert"
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
		err = errors.New("routing err")
	}

	return out, err
}

/*
func TestNew(t *testing.T) {

	blog.New(blog.NewWithDefault())

	b := module.GetBuilder(Name)
	b.AddModuleOption(WithListen(":14111"))
	s := b.Build("TestNew").(server.IServer)
	s.Init()

	proto.RegisterListenServer(s.Server().(*grpc.Server), &rpcServer{})
	s.Run()
	defer s.Close()
	time.Sleep(time.Millisecond * 10)

	conn, err := grpc.Dial(":14111", grpc.WithInsecure())
	if err != nil {
		t.Error(err)
	}
	rres := new(proto.RouteRes)

	err = conn.Invoke(context.Background(), "/bproto.listen/routing", &proto.RouteReq{
		Nod:     "normal",
		Service: "test",
		ReqBody: nil,
	}, rres)
	assert.Equal(t, err, nil)
	err = conn.Invoke(context.Background(), "/bproto.listen/routing", &proto.RouteReq{
		Nod:     "normal",
		Service: "errtest",
		ReqBody: nil,
	}, rres)
	fmt.Println(err.Error())
	assert.NotEqual(t, err, nil)
}
*/

func TestOpts(t *testing.T) {

	cfg := server.Parm{
		ListenAddr: ":14222",
	}

	op := server.WithListen(":1201")
	op(&cfg)
	assert.Equal(t, cfg.ListenAddr, ":1201")

	//top := WithTracing()
	//top(&cfg)
	//assert.Equal(t, cfg.isTracing, true)

}
