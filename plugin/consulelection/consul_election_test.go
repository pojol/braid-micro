package consulelection

import (
	"testing"
	"time"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/mock"
	"github.com/pojol/braid/module/election"
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

	e := election.GetBuilder(ElectionName).Build(Cfg{
		Address:           mock.ConsulAddr,
		Name:              "test",
		LockTick:          time.Second,
		RefushSessionTick: time.Second,
	})

	e.Run()
	time.Sleep(time.Second)
	e.Close()
}
