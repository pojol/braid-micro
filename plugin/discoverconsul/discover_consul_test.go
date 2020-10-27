package discoverconsul

import (
	"testing"
	"time"

	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/mock"
	"github.com/pojol/braid/module"
	"github.com/pojol/braid/module/balancer"
	"github.com/pojol/braid/module/logger"
	"github.com/pojol/braid/module/mailbox"
	"github.com/pojol/braid/plugin/balancerswrr"
	"github.com/pojol/braid/plugin/mailboxnsq"
	"github.com/pojol/braid/plugin/zaplogger"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	mock.Init()

	r := redis.New()
	r.Init(redis.Config{
		Address:        mock.RedisAddr,
		ReadTimeOut:    time.Millisecond * time.Duration(5000),
		WriteTimeOut:   time.Millisecond * time.Duration(5000),
		ConnectTimeOut: time.Millisecond * time.Duration(2000),
		IdleTimeout:    time.Millisecond * time.Duration(0),
		MaxIdle:        16,
		MaxActive:      128,
	})
	defer r.Close()

	mb, _ := mailbox.GetBuilder(mailboxnsq.Name).Build("TestDiscover")
	bb := module.GetBuilder(balancerswrr.Name)
	log, _ := logger.GetBuilder(zaplogger.Name).Build(logger.DEBUG)
	balancer.NewGroup(bb, mb, log)

	m.Run()
}

func TestDiscover(t *testing.T) {

	b := module.GetBuilder(Name)

	mb, err := mailbox.GetBuilder(mailboxnsq.Name).Build("TestDiscover")
	b.AddOption(WithConsulAddr(mock.ConsulAddr))
	log, err := logger.GetBuilder(zaplogger.Name).Build(logger.DEBUG)

	d, err := b.Build("test", mb, log)
	assert.Equal(t, err, nil)

	d.Run()

	time.Sleep(time.Second)
	d.Close()
}

func TestParm(t *testing.T) {
	b := module.GetBuilder(Name)

	mb, err := mailbox.GetBuilder(mailboxnsq.Name).Build("TestDiscover")
	b.AddOption(WithConsulAddr("http://127.0.0.1:8500"))
	b.AddOption(WithTag("TestParm"))
	b.AddOption(WithBlacklist([]string{"gate"}))
	b.AddOption(WithInterval(time.Second))
	log, err := logger.GetBuilder(zaplogger.Name).Build(logger.DEBUG)

	discv, err := b.Build("test", mb, log)
	assert.NotEqual(t, err, nil)

	cd := discv.(*consulDiscover)
	assert.Equal(t, cd.parm.Address, "http://127.0.0.1:8500")
	assert.Equal(t, cd.parm.Tag, "TestParm")
	assert.Equal(t, cd.parm.Blacklist, []string{"gate"})
	assert.Equal(t, cd.parm.Interval, time.Second)
}
