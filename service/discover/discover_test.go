package discover

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pojol/braid/cache/redis"
	"github.com/pojol/braid/log"
	"github.com/pojol/braid/mock"
	"github.com/pojol/braid/service/balancer"
)

func TestDiscover(t *testing.T) {

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

	balancer.New()
	d := New("test", mock.ConsulAddr, WithInterval(100))

	d.Run()

	time.Sleep(time.Second)
	d.Close()
}

func TestOpts(t *testing.T) {

	mock.Init()
	New("test", mock.ConsulAddr, WithInterval(100))
	assert.Equal(t, dc.cfg.Interval.Milliseconds(), int64(100))
}
