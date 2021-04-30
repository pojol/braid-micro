package discoverconsul

import (
	"testing"
	"time"

	"github.com/pojol/braid-go/3rd/redis"
	"github.com/pojol/braid-go/mock"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/mailbox"
	"github.com/pojol/braid-go/modules/balancergroupbase"
	"github.com/pojol/braid-go/modules/balancerswrr"
	"github.com/pojol/braid-go/modules/mailboxnsq"
	"github.com/pojol/braid-go/modules/zaplogger"
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

	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	mb, _ := mailbox.GetBuilder(mailboxnsq.Name).Build("TestDiscover", log)

	bgb := module.GetBuilder(balancergroupbase.Name)
	bgb.AddOption(balancergroupbase.WithStrategy([]string{balancerswrr.Name}))
	b, _ := bgb.Build("discover_consul_test", mb, log)

	b.Init()
	b.Run()
	defer b.Close()

	m.Run()
}

func TestDiscover(t *testing.T) {

	b := module.GetBuilder(Name)
	assert.Equal(t, b.Type(), module.TyDiscover)

	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	mb, err := mailbox.GetBuilder(mailboxnsq.Name).Build("TestDiscover", log)
	assert.Equal(t, err, nil)

	b.AddOption(WithConsulAddr(mock.ConsulAddr))
	b.AddOption(WithSyncServiceInterval(time.Millisecond * 100))
	b.AddOption(WithSyncServiceWeightInterval(time.Millisecond * 100))
	b.AddOption(WithBlacklist([]string{"gate"}))
	assert.Equal(t, err, nil)

	d, err := b.Build("test", mb, log)
	assert.Equal(t, err, nil)
	dc := d.(*consulDiscover)
	assert.Equal(t, dc.InBlacklist("gate"), true)
	assert.Equal(t, dc.InBlacklist("login"), false)

	d.Init()
	d.Run()

	time.Sleep(time.Second)
	d.Close()
}

func TestParm(t *testing.T) {
	b := module.GetBuilder(Name)

	log, err := logger.GetBuilder(zaplogger.Name).Build()
	mb, err := mailbox.GetBuilder(mailboxnsq.Name).Build("TestDiscover", log)
	assert.Equal(t, err, nil)

	b.AddOption(WithConsulAddr(mock.ConsulAddr))
	b.AddOption(WithTag("TestParm"))
	b.AddOption(WithBlacklist([]string{"gate"}))
	b.AddOption(WithSyncServiceInterval(time.Second))
	b.AddOption(WithSyncServiceWeightInterval(time.Second))

	assert.Equal(t, err, nil)

	discv, err := b.Build("test", mb, log)
	assert.Equal(t, err, nil)

	cd := discv.(*consulDiscover)
	assert.Equal(t, cd.parm.Address, mock.ConsulAddr)
	assert.Equal(t, cd.parm.Tag, "TestParm")
	assert.Equal(t, cd.parm.Blacklist, []string{"gate"})
	assert.Equal(t, cd.parm.SyncServicesInterval, time.Second)
	assert.Equal(t, cd.parm.SyncServiceWeightInterval, time.Second)
}
