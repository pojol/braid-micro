package electorconsul

import (
	"testing"
	"time"

	"github.com/pojol/braid/mock"
	"github.com/pojol/braid/module"
	"github.com/pojol/braid/module/logger"
	"github.com/pojol/braid/module/mailbox"
	"github.com/pojol/braid/plugin/mailboxnsq"
	"github.com/pojol/braid/plugin/zaplogger"
)

func TestElection(t *testing.T) {

	mock.Init()

	mb, _ := mailbox.GetBuilder(mailboxnsq.Name).Build("TestDiscover")

	eb := module.GetBuilder(Name)
	eb.AddOption(WithConsulAddr(mock.ConsulAddr))

	log, _ := logger.GetBuilder(zaplogger.Name).Build(logger.DEBUG)

	e, _ := eb.Build("test_elector_with_consul", mb, log)

	e.Run()
	time.Sleep(time.Second)
	e.Close()
}
