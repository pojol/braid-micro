package election

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pojol/braid/log"
	"github.com/pojol/braid/mock"
)

func TestElection(t *testing.T) {

	mock.Init()

	l := log.New("test")
	l.Init()

	e, _ := New("test", mock.ConsulAddr)

	e.Run()
	time.Sleep(time.Second)
	e.Close()
}

func TestOpts(t *testing.T) {
	mock.Init()

	New("test", mock.ConsulAddr, WithLockTick(1000), WithRefushTick(1000))
	assert.Equal(t, e.cfg.LockTick.Milliseconds(), int64(1000))
	assert.Equal(t, e.cfg.RefushSessionTick.Milliseconds(), int64(1000))
}
