package linkerredis

import (
	"testing"
	"time"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/mock"
	"github.com/pojol/braid/module/elector"
	"github.com/pojol/braid/module/linker"
	"github.com/pojol/braid/module/pubsub"
	"github.com/pojol/braid/plugin/electorconsul"
	"github.com/pojol/braid/plugin/pubsubnsq"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
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

	m.Run()
}

func TestLinkerTarget(t *testing.T) {
	LinkerRedisPrefix = "testlinkertarget"

	psb := pubsub.GetBuilder(pubsubnsq.PubsubName)
	psb.SetCfg(pubsubnsq.NsqConfig{
		Addres:       []string{mock.NsqdAddr},
		LookupAddres: []string{mock.NSQLookupdAddr},
	})
	ps, _ := psb.Build()
	eb := elector.GetBuilder(electorconsul.ElectionName)
	eb.SetCfg(electorconsul.Cfg{
		Address:           mock.ConsulAddr,
		Name:              "testlinkertarget",
		LockTick:          time.Second,
		RefushSessionTick: time.Second,
	})
	e, _ := eb.Build()
	defer e.Close()

	b := linker.GetBuilder(LinkerName)
	b.SetCfg(Config{
		ServiceName: "base",
	})
	lk := b.Build(e, ps)
	err := lk.Link("mail", "token01", "127.0.0.1")
	assert.Equal(t, err, nil)
	err = lk.Link("social", "token01", "127.0.0.2")
	assert.Equal(t, err, nil)

	addr, err := lk.Target("mail", "token01")
	assert.Equal(t, err, nil)
	assert.Equal(t, addr, "127.0.0.1")

	addr, err = lk.Target("mail", "token02")
	assert.Equal(t, addr, "")

	num, err := lk.Num("mail", "127.0.0.1")
	assert.Equal(t, err, nil)
	assert.Equal(t, num, 1)

	ps.Pub(LinkerTopicUnlink, &pubsub.Message{
		Body: []byte("token01"),
	})
	ps.Pub(LinkerTopicUnlink, &pubsub.Message{
		Body: []byte("token02"),
	})
	time.Sleep(time.Millisecond * 500)

	num, err = lk.Num("mail", "127.0.0.1")
	assert.Equal(t, num, 0)
}

func TestLinkerDown(t *testing.T) {
	LinkerRedisPrefix = "testlinkerdown"

	psb := pubsub.GetBuilder(pubsubnsq.PubsubName)
	psb.SetCfg(pubsubnsq.NsqConfig{
		Addres:       []string{mock.NsqdAddr},
		LookupAddres: []string{mock.NSQLookupdAddr},
	})
	ps, _ := psb.Build()
	eb := elector.GetBuilder(electorconsul.ElectionName)
	eb.SetCfg(electorconsul.Cfg{
		Address:           mock.ConsulAddr,
		Name:              "testlinkerdown",
		LockTick:          time.Second,
		RefushSessionTick: time.Second,
	})
	e, _ := eb.Build()
	defer e.Close()

	b := linker.GetBuilder(LinkerName)
	b.SetCfg(Config{
		ServiceName: "base",
	})
	lk := b.Build(e, ps)

	err := lk.Link("mail", "token01", "127.0.0.1")
	assert.Equal(t, err, nil)
	err = lk.Link("mail", "token02", "127.0.0.1")
	assert.Equal(t, err, nil)
	err = lk.Link("mail", "token03", "127.0.0.2")
	assert.Equal(t, err, nil)

	err = lk.Down("mail", "127.0.0.1")
	assert.Equal(t, err, nil)

	addr, _ := lk.Target("mail", "token01")
	assert.Equal(t, addr, "")

	lk.Unlink("token01")
	lk.Unlink("token02")
	lk.Unlink("token03")
}
