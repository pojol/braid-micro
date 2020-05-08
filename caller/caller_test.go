package caller

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pojol/braid/log"
	"github.com/pojol/braid/mock"
)

func TestCaller(t *testing.T) {

	mock.Init()
	l := log.New()
	l.Init(log.Config{
		Path:   "test",
		Suffex: ".log",
		Mode:   "debug",
	})

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

	addr, _ := c.getNodeWithCoordinate(context.Background(), "test", "func")
	assert.Equal(t, addr, "")

	addr, _ = c.findNode(context.Background(), "test", "test", "")
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
