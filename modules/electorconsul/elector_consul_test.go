package electorconsul

import (
	"testing"
	"time"

	"github.com/pojol/braid-go/mock"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/mailbox"
	"github.com/pojol/braid-go/modules/mailboxnsq"
	"github.com/pojol/braid-go/modules/zaplogger"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	mock.Init()
	m.Run()
}

func TestElection(t *testing.T) {

	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	mb, _ := mailbox.GetBuilder(mailboxnsq.Name).Build("TestDiscover", log)

	eb := module.GetBuilder(Name)
	eb.AddOption(WithConsulAddr(mock.ConsulAddr))

	e, _ := eb.Build("test_elector_with_consul", mb, log)

	e.Run()
	time.Sleep(time.Second)
	e.Close()
}

func TestParm(t *testing.T) {

	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	mb, _ := mailbox.GetBuilder(mailboxnsq.Name).Build("TestDiscover", log)

	eb := module.GetBuilder(Name)
	eb.AddOption(WithConsulAddr(mock.ConsulAddr))
	eb.AddOption(WithLockTick(time.Second))
	eb.AddOption(WithSessionTick(time.Second))

	e, _ := eb.Build("test_elector_with_consul", mb, log)

	ec := e.(*consulElection)
	assert.Equal(t, ec.parm.ConsulAddr, mock.ConsulAddr)
	assert.Equal(t, ec.parm.LockTick, time.Second)
	assert.Equal(t, ec.parm.RefushSessionTick, time.Second)
}
