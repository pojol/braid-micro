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

	c := New(mock.ConsulAddr)
	time.Sleep(time.Millisecond * 200)

	addr, _ := c.findNode(context.Background(), "test", "test", "")
	assert.Equal(t, addr, "")

	c.Call(context.Background(), "", "", nil, []byte{})
}

func TestInitNum(t *testing.T) {
	mock.Init()
	l := log.New("test")
	l.Init()

	New(mock.ConsulAddr)
}
