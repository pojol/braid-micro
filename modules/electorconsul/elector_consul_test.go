package electorconsul

import (
	"testing"
	"time"

	"github.com/pojol/braid-go/mock"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/elector"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/pubsub"
	"github.com/pojol/braid-go/modules/moduleparm"
	"github.com/pojol/braid-go/modules/pubsubnsq"
	"github.com/pojol/braid-go/modules/zaplogger"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	mock.Init()
	m.Run()
}

func TestElection(t *testing.T) {

	log := module.GetBuilder(zaplogger.Name).Build("TestElection").(logger.ILogger)
	mb := module.GetBuilder(pubsubnsq.Name).Build("TestElection", moduleparm.WithLogger(log)).(pubsub.IPubsub)

	eb := module.GetBuilder(Name)
	eb.AddModuleOption(WithConsulAddr(mock.ConsulAddr))

	e := eb.Build("test_elector_with_consul",
		moduleparm.WithLogger(log),
		moduleparm.WithPubsub(mb)).(elector.IElector)

	e.Run()
	time.Sleep(time.Second)
	e.Close()
}

func TestParm(t *testing.T) {

	log := module.GetBuilder(zaplogger.Name).Build("TestParm").(logger.ILogger)
	mb := module.GetBuilder(pubsubnsq.Name).Build("TestParm", moduleparm.WithLogger(log)).(pubsub.IPubsub)

	eb := module.GetBuilder(Name)
	eb.AddModuleOption(WithConsulAddr(mock.ConsulAddr))
	eb.AddModuleOption(WithLockTick(time.Second))
	eb.AddModuleOption(WithSessionTick(time.Second))

	e := eb.Build("test_elector_with_consul",
		moduleparm.WithLogger(log),
		moduleparm.WithPubsub(mb)).(*consulElection)

	assert.Equal(t, e.parm.ConsulAddr, mock.ConsulAddr)
	assert.Equal(t, e.parm.LockTick, time.Second)
	assert.Equal(t, e.parm.RefushSessionTick, time.Second)
}
