package consuldiscover

import (
	"testing"
	"time"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/mock"
	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/plugin/balancer"
	"github.com/pojol/braid/plugin/balancer/swrrbalancer"
	_ "github.com/pojol/braid/plugin/balancer/swrrbalancer"
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
	defer r.Close()

	balancer.NewGroup(balancer.GetBuilder(swrrbalancer.BalancerName))

	m.Run()
}

func TestDiscover(t *testing.T) {

	b := discover.GetBuilder(DiscoverName)
	b.SetCfg(Cfg{
		Name:     "test",
		Interval: time.Second * 2,
		Address:  mock.ConsulAddr,
	})
	d := b.Build()

	d.Discover()

	time.Sleep(time.Second)
	d.Close()
}
