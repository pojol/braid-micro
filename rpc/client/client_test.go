package client

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pojol/braid/log"
	"github.com/pojol/braid/mock"
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

	New("test", mock.ConsulAddr)
	time.Sleep(time.Millisecond * 200)

	conn, err := GetConn("test")
	if err != nil {
		t.Error(err)
		return
	}
	defer conn.Put()

	/*
		client := bproto.NewListenClient(conn.ClientConn)
		_, err = client.Routing(context.Background(), &bproto.RouteReq{})
		if err != nil {
			conn.Unhealthy()
		}
	*/
	//assert.Equal(t, err, nil)
}

func TestFindNode(t *testing.T) {
	New("test", mock.ConsulAddr)
	time.Sleep(time.Millisecond * 100)

	addr, _ := c.findNode("test", "")
	assert.Equal(t, addr, "")
}

func TestOpts(t *testing.T) {
	New("test", mock.ConsulAddr, WithTracing(), WithPoolInitNum(10), WithPoolCapacity(128), WithPoolIdle(100))

	assert.Equal(t, c.cfg.PoolInitNum, 10)
	assert.Equal(t, c.cfg.PoolCapacity, 128)
	assert.Equal(t, c.cfg.PoolIdle.Seconds(), float64(100))
}
