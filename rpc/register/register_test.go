package register

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pojol/braid/log"
	"github.com/pojol/braid/rpc/dispatcher/bproto"
	"google.golang.org/grpc"
)

func TestNew(t *testing.T) {

	l := log.New("test")
	err := l.Init()
	if err != nil {
		t.Error(err)
	}

	s := New("normal", WithListen(":14111"))
	err = s.Init()
	if err != nil {
		t.Error(err)
	}

	s.Regist("test", func(ctx context.Context, in []byte) (out []byte, err error) {
		fmt.Println("pong")
		return nil, nil
	})

	s.Run()
	time.Sleep(time.Millisecond * 10)

	conn, err := grpc.Dial("localhost:14111", grpc.WithInsecure())
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

	wtr := New("testopt", WithTracing())
	assert.Equal(t, wtr.cfg.Tracing, true)

	wlr := New("testopt", WithListen(":1201"))
	assert.Equal(t, wlr.cfg.ListenAddress, ":1201")
}
