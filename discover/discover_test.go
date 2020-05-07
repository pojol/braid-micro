package discover

import (
	"testing"
	"time"

	"github.com/pojol/braid/balancer"
	"github.com/pojol/braid/cache/redis"
	"github.com/pojol/braid/log"
	"github.com/pojol/braid/mock"
)

func TestDiscover(t *testing.T) {

	mock.Init()

	l := log.New()
	l.Init(log.Config{
		Path:   "test",
		Suffex: ".log",
		Mode:   "debug",
	})

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

	ba := balancer.New()
	err := ba.Init(balancer.SelectorCfg{})
	if err != nil {
		t.Error(err)
	}

	d := New()
	d.Init(Config{
		Interval:      100,
		ConsulAddress: mock.ConsulAddr,
	})

	d.Run()

	time.Sleep(time.Second)
	d.Close()
}
