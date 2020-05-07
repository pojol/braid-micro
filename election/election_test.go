package election

import (
	"testing"
	"time"

	"github.com/pojol/braid/log"
	"github.com/pojol/braid/mock"
)

func TestElection(t *testing.T) {

	mock.Init()

	l := log.New()
	l.Init(log.Config{
		Path:   "test",
		Suffex: ".log",
		Mode:   "debug",
	})

	e := New()
	e.Init(Config{
		Address:           mock.ConsulAddr,
		Name:              "test",
		RefushSessionTick: 500 * time.Millisecond,
		LockTick:          200 * time.Millisecond,
	})

	e.Run()
	time.Sleep(time.Second)
	e.Close()
}
