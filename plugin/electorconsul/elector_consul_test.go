package electorconsul

import (
	"testing"
	"time"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/mock"
	"github.com/pojol/braid/module/elector"
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

	eb := elector.GetBuilder(Name)
	eb.AddOption(WithConsulAddr(mock.ConsulAddr))
	e, _ := eb.Build("test_elector_with_consul")

	e.Run()
	time.Sleep(time.Second)
	e.IsMaster()
	e.Close()
}
