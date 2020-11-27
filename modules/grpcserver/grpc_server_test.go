package grpcserver

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pojol/braid/module/logger"
	"github.com/pojol/braid/module/rpc/server"
	"github.com/pojol/braid/modules/grpcclient/bproto"
	"github.com/pojol/braid/modules/zaplogger"
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

func TestNew(t *testing.T) {

	log, _ := logger.GetBuilder(zaplogger.Name).Build(logger.DEBUG)

	b := server.GetBuilder(Name)
	b.AddOption(WithListen(":14111"))
	s, _ := b.Build("TestNew", log)

	bproto.RegisterListenServer(s.Server().(*grpc.Server), &rpcServer{})

	s.Run()
	defer s.Close()
	time.Sleep(time.Millisecond * 10)

	conn, err := grpc.Dial(":14111", grpc.WithInsecure())
	if err != nil {
		t.Error(err)
	}
	rres := new(bproto.RouteRes)

	err = conn.Invoke(context.Background(), "/bproto.listen/routing", &bproto.RouteReq{
		Nod:     "normal",
		Service: "test",
		ReqBody: nil,
	}, rres)
	assert.Equal(t, err, nil)
	err = conn.Invoke(context.Background(), "/bproto.listen/routing", &bproto.RouteReq{
		Nod:     "normal",
		Service: "errtest",
		ReqBody: nil,
	}, rres)
	assert.NotEqual(t, err, nil)
}

func TestOpts(t *testing.T) {

	cfg := Parm{
		ListenAddr: ":14222",
	}

	op := WithListen(":1201")
	op(&cfg)
	assert.Equal(t, cfg.ListenAddr, ":1201")

	//top := WithTracing()
	//top(&cfg)
	//assert.Equal(t, cfg.isTracing, true)
}
