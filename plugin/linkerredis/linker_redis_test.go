package linkerredis

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/mock"
	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/module/elector"
	"github.com/pojol/braid/module/linkcache"
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
	LinkerRedisPrefix = "testlinkertarget_"

	psb := pubsub.GetBuilder(pubsubnsq.Name)
	psb.AddOption(pubsubnsq.WithLookupAddr([]string{mock.NSQLookupdAddr}))
	psb.AddOption(pubsubnsq.WithNsqdAddr([]string{mock.NsqdAddr}))
	ps, _ := psb.Build("gate")
	eb := elector.GetBuilder(electorconsul.Name)
	eb.AddOption(electorconsul.WithConsulAddr(mock.ConsulAddr))
	e, _ := eb.Build("testlinkertarget")
	defer e.Close()

	b := linkcache.GetBuilder(Name)
	b.AddOption(WithElector(e))
	b.AddOption(WithClusterPubsub(ps))

	lk, err := b.Build("gate")
	assert.Equal(t, err, nil)

	nods := []discover.Node{
		{
			ID:      "a001",
			Name:    "base",
			Address: "127.0.0.1:12001",
		},
		{
			ID:      "a002",
			Name:    "login",
			Address: "127.0.0.1:13001",
		},
	}

	err = lk.Link("token01", nods[0])
	assert.Equal(t, err, nil)

	err = lk.Link("token01", nods[1])
	assert.Equal(t, err, nil)

	err = lk.Link("token02", nods[0])
	assert.Equal(t, err, nil)

	addr, err := lk.Target("token01", "base")
	assert.Equal(t, err, nil)
	assert.Equal(t, addr, "127.0.0.1:12001")

	num, err := lk.Num(nods[0])
	assert.Equal(t, err, nil)
	assert.Equal(t, num, 2)

	lk.Unlink("token01")
	lk.Unlink("token02")

	for _, v := range nods {
		lk.Down(v)
	}

	time.Sleep(time.Millisecond * 500)
}

func BenchmarkLink(b *testing.B) {
	LinkerRedisPrefix = "benchmarklink"

	psb := pubsub.GetBuilder(pubsubnsq.Name)
	psb.AddOption(pubsubnsq.WithLookupAddr([]string{mock.NSQLookupdAddr}))
	psb.AddOption(pubsubnsq.WithNsqdAddr([]string{mock.NsqdAddr}))
	ps, _ := psb.Build("TestLinkerTarget")
	eb := elector.GetBuilder(electorconsul.Name)
	eb.AddOption(electorconsul.WithConsulAddr(mock.ConsulAddr))
	e, _ := eb.Build("testlinkertarget")
	defer e.Close()

	lb := linkcache.GetBuilder(Name)
	lb.AddOption(WithElector(e))
	lb.AddOption(WithClusterPubsub(ps))

	lk, err := lb.Build("gate")
	assert.Equal(b, err, nil)
	rand.Seed(time.Now().UnixNano())

	baseTargets := []discover.Node{
		{
			ID:      "a001",
			Name:    "base",
			Address: "127.0.0.1:12001",
		},
		{
			ID:      "a002",
			Name:    "base",
			Address: "127.0.0.1:12002",
		},
		{
			ID:      "a003",
			Name:    "base",
			Address: "127.0.0.1:12003",
		},
	}

	loginTargets := []discover.Node{
		{
			ID:      "b001",
			Name:    "login",
			Address: "127.0.0.1:13001",
		},
		{
			ID:      "b002",
			Name:    "login",
			Address: "127.0.0.1:13001",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lk.Link("token"+strconv.Itoa(i), baseTargets[rand.Intn(len(baseTargets))])
		lk.Link("token"+strconv.Itoa(i), loginTargets[rand.Intn(len(loginTargets))])
	}
}
