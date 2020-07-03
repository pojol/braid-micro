package client

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/mock"
	"github.com/pojol/braid/module/rpc/client/bproto"
)

func TestMain(m *testing.M) {
	mock.Init()
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

	m.Run()
}

func TestCaller(t *testing.T) {

	c := New("test", mock.ConsulAddr)
	c.Discover()
	defer c.Close()

	time.Sleep(time.Millisecond * 200)
	res := &bproto.RouteRes{}

	nodeName := "gateway"
	serviceName := "service"

	tc, cancel := context.WithTimeout(context.TODO(), time.Millisecond*200)
	Invoke(tc, nodeName, "/bproto.listen/routing", "", &bproto.RouteReq{
		Nod:     nodeName,
		Service: serviceName,
		ReqBody: []byte{},
	}, res)

	cancel()
}

func TestOpts(t *testing.T) {
	New("test", mock.ConsulAddr, WithTracing(), WithPoolInitNum(10), WithPoolCapacity(128), WithPoolIdle(100))

	assert.Equal(t, c.cfg.PoolInitNum, 10)
	assert.Equal(t, c.cfg.PoolCapacity, 128)
	assert.Equal(t, c.cfg.PoolIdle.Seconds(), float64(100))
}
