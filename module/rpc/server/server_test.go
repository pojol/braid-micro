package server

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/module/rpc/client/bproto"
	"google.golang.org/grpc"
)

type rpcServer struct {
	bproto.ListenServer
}

func (rs *rpcServer) Routing(ctx context.Context, req *bproto.RouteReq) (*bproto.RouteRes, error) {
	out := new(bproto.RouteRes)
	var err error
	fmt.Println("pong")

	if req.Service == "test" {
		err = nil
	} else {
		err = errors.New("err")
	}

	return out, err
}

func TestNew(t *testing.T) {

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

	s := New("normal", WithListen(":14111"))

	bproto.RegisterListenServer(Get(), &rpcServer{})

	s.Run()
	time.Sleep(time.Millisecond * 10)

	conn, err := grpc.Dial("localhost:14111", grpc.WithInsecure())
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

	s.Close()
}

func TestOpts(t *testing.T) {

	New("testopt", WithTracing())
	assert.Equal(t, server.cfg.Tracing, true)

	New("testopt", WithListen(":1201"))
	assert.Equal(t, server.cfg.ListenAddress, ":1201")
}
