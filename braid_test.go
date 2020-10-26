package braid

import (
	"testing"
	"time"

	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/mock"
	"github.com/pojol/braid/plugin/balancerswrr"
	"github.com/pojol/braid/plugin/discoverconsul"
	"github.com/pojol/braid/plugin/electorconsul"
	"github.com/pojol/braid/plugin/linkerredis"
	"github.com/pojol/braid/plugin/mailboxnsq"
)

func TestMain(m *testing.M) {

	mock.Init()

	c := redis.New()
	c.Init(redis.Config{
		Address:        mock.RedisAddr,
		ReadTimeOut:    time.Millisecond * time.Duration(5000),
		WriteTimeOut:   time.Millisecond * time.Duration(5000),
		ConnectTimeOut: time.Millisecond * time.Duration(2000),
		IdleTimeout:    time.Millisecond * time.Duration(0),
		MaxIdle:        16,
		MaxActive:      128,
	})
	defer c.Close()

	m.Run()
}

func TestPlugin(t *testing.T) {

	b, _ := New(
		"test_plugin",
		mailboxnsq.WithLookupAddr([]string{mock.NSQLookupdAddr}),
		mailboxnsq.WithNsqdAddr([]string{mock.NsqdAddr}),
	)

	b.RegistPlugin(
		Balancer(balancerswrr.Name),
		LinkCache(linkerredis.Name, linkerredis.WithRedisAddr(mock.RedisAddr)),
		Discover(
			discoverconsul.Name,
			discoverconsul.WithConsulAddr(mock.ConsulAddr),
		),
		Elector(
			electorconsul.Name,
			electorconsul.WithConsulAddr(mock.ConsulAddr),
			electorconsul.WithLockTick(time.Second*2),
		),
	)

	b.Run()
	defer b.Close()
}

func TestWithClient(t *testing.T) {
	/*
		b := New("test")
		b.RegistPlugin(DiscoverByConsul(mock.ConsulAddr, discoverconsul.WithInterval(time.Second*3)),
			BalancerBySwrr(),
			GRPCClient(grpcclient.WithPoolCapacity(128)))

		b.Run()
		defer b.Close()

		Client().Invoke(context.TODO(), "targeNodeName", "/proto.node/method", "", nil, nil)
	*/
}
