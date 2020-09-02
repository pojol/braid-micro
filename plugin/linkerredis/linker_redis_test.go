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

func TestTarget(t *testing.T) {

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

	ps, _ := pubsub.GetBuilder(pubsubnsq.PubsubName).Build()
	eb := elector.GetBuilder(electorconsul.ElectionName)
	eb.SetCfg(electorconsul.Cfg{
		Address:           mock.ConsulAddr,
		Name:              "test",
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

	r.Del(LinkerPrefix + "base_child_" + "127.0.0.1")
	r.Del(LinkerTokenPool)

	num, err := lk.Num("127.0.0.1")
	assert.Equal(t, num, 0)
	assert.Equal(t, err, nil)

	err = lk.Link("xxx", "127.0.0.1")
	assert.Equal(t, err, nil)

	addr, err := lk.Target("xxx")
	assert.Equal(t, err, nil)
	assert.Equal(t, addr, "127.0.0.1")
}
