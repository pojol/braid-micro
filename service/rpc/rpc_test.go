package rpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pojol/braid/log"
	"github.com/pojol/braid/mock"
	"github.com/pojol/braid/service/balancer"
)

func TestCaller(t *testing.T) {

	mock.Init()
	l := log.New()
	l.Init(log.Config{
		Path:   "test",
		Suffex: ".log",
		Mode:   "debug",
	})

	_ = balancer.New()

	c := New()
	c.Init(Config{
		ConsulAddress: mock.ConsulAddr,
		PoolInitNum:   8,
		PoolCapacity:  32,
		PoolIdle:      time.Second * 120,
		Tracing:       false,
	})

	c.Run()
	time.Sleep(time.Millisecond * 200)

	addr, _ := c.findNode(context.Background(), "test", "test", "")
	assert.Equal(t, addr, "")

	c.Call(context.Background(), "", "", "", []byte{})
}

func TestInitNum(t *testing.T) {
	mock.Init()
	l := log.New()
	l.Init(log.Config{
		Path:   "test",
		Suffex: ".log",
		Mode:   "debug",
	})

	c := New()
	c.Init(Config{})
}
