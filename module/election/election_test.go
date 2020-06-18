package election

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/mock"
)

func TestElection(t *testing.T) {

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
