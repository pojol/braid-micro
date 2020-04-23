package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pojol/braid/caller/brpc"
	"github.com/pojol/braid/log"
	"google.golang.org/grpc"
)

func TestNew(t *testing.T) {

	l := log.New()
	l.Init(log.Config{
		Path:   "test",
		Suffex: ".log",
		Mode:   "debug",
	})

	s := New()
	err := s.Init(Config{
		Tracing: false,
		Name:    "test",
	})
	if err != nil {
		t.Error(err)
	}

	s.Regist("test", func(ctx context.Context, in []byte) (out []byte, err error) {
		fmt.Println("pong")
		return nil, nil
	})

	s.Run()

	conn, err := grpc.Dial("localhost"+ListenAddress, grpc.WithInsecure())
	rres := new(brpc.RouteRes)

	err = conn.Invoke(context.Background(), "/brpc.gateway/routing", &brpc.RouteReq{
		Box:     "normal",
		Service: "test",
		ReqBody: nil,
	}, rres)
	assert.Equal(t, err, nil)
	err = conn.Invoke(context.Background(), "/brpc.gateway/routing", &brpc.RouteReq{
		Box:     "normal",
		Service: "errtest",
		ReqBody: nil,
	}, rres)
	assert.NotEqual(t, err, nil)

	s.Close()
}
