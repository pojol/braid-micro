package discover

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/internal/balancer"
	"github.com/pojol/braid/mock"
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

	bg := balancer.NewGroup()
	d := New("test", mock.ConsulAddr, bg, WithInterval(100))

	d.Run()

	time.Sleep(time.Second)
	d.Close()
}

func TestOpts(t *testing.T) {

	mock.Init()
	New("test", mock.ConsulAddr, nil, WithInterval(100))
	assert.Equal(t, dc.cfg.Interval.Milliseconds(), int64(100))
}
