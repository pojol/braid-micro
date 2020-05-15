package discover

import (
	"testing"
	"time"

	"github.com/pojol/braid/cache/redis"
	"github.com/pojol/braid/log"
	"github.com/pojol/braid/mock"
	"github.com/pojol/braid/service/balancer"
)

func TestDiscover(t *testing.T) {

	mock.Init()

	l := log.New("test")
	l.Init()

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
	err := ba.Init()
	if err != nil {
		t.Error(err)
	}

	d := New("test", mock.ConsulAddr, WithInterval(100))
	d.Init()

	d.Run()

	time.Sleep(time.Second)
	d.Close()
}
