package discoverconsul

import (
	"testing"
	"time"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/mock"
	"github.com/pojol/braid/module/balancer"
	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/module/pubsub"
	"github.com/pojol/braid/plugin/balancerswrr"
	"github.com/pojol/braid/plugin/pubsubproc"
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
	defer r.Close()

	ps, _ := pubsub.GetBuilder(pubsubproc.PubsubName).Build()
	balancer.NewGroup(balancer.GetBuilder(balancerswrr.BalancerName), ps)

	m.Run()
}

func TestDiscover(t *testing.T) {

	b := discover.GetBuilder(Name)

	ps, _ := pubsub.GetBuilder(pubsubproc.PubsubName).Build()
	b.AddOption(WithProcPubsub(ps))
	b.AddOption(WithConsulAddress(mock.ConsulAddr))

	d, err := b.Build("test")
	assert.Equal(t, err, nil)

	d.Discover()

	time.Sleep(time.Second)
	d.Close()
}

func TestParmAddress(t *testing.T) {
	b := discover.GetBuilder(Name)

	ps, _ := pubsub.GetBuilder(pubsubproc.PubsubName).Build()
	b.AddOption(WithProcPubsub(ps))
	b.AddOption(WithConsulAddress("http://127.0.0.1:8500"))

	_, err := b.Build("test")
	assert.NotEqual(t, err, nil)
}
