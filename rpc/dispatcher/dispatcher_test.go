package dispatcher

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
	l := log.New("test")
	l.Init()

	_ = balancer.New()

	c := New("test", mock.ConsulAddr)
	time.Sleep(time.Millisecond * 200)

	addr, _ := c.findNode(context.Background(), "test", "test", "")
	assert.Equal(t, addr, "")

	Call(context.Background(), "", "", nil, []byte{})
}

func TestInitNum(t *testing.T) {
	mock.Init()
	l := log.New("test")
	l.Init()

	New("test", mock.ConsulAddr)
}

func TestOpts(t *testing.T) {
	mock.Init()
	New("test", mock.ConsulAddr, WithTracing(), WithPoolInitNum(10), WithPoolCapacity(128), WithPoolIdle(100))

	assert.Equal(t, r.cfg.PoolInitNum, 10)
	assert.Equal(t, r.cfg.PoolCapacity, 128)
	assert.Equal(t, r.cfg.PoolIdle.Seconds(), float64(100))
}
