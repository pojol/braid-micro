package linkerredis

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/mock"
	"github.com/pojol/braid/module"
	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/module/linkcache"
	"github.com/pojol/braid/module/logger"
	"github.com/pojol/braid/module/mailbox"
	"github.com/pojol/braid/plugin/electorconsul"
	"github.com/pojol/braid/plugin/mailboxnsq"
	"github.com/pojol/braid/plugin/zaplogger"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	mock.Init()

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

	mbb := mailbox.GetBuilder(mailboxnsq.Name)
	mbb.AddOption(mailboxnsq.WithLookupAddr([]string{mock.NSQLookupdAddr}))
	mbb.AddOption(mailboxnsq.WithNsqdAddr([]string{mock.NsqdAddr}))
	mb, _ := mbb.Build("testlinkertarget")

	log, _ := logger.GetBuilder(zaplogger.Name).Build()

	eb := module.GetBuilder(electorconsul.Name)
	eb.AddOption(electorconsul.WithConsulAddr(mock.ConsulAddr))
	e, _ := eb.Build("testlinkertarget", mb, log)
	defer e.Close()

	b := module.GetBuilder(Name)

	lk, err := b.Build("gate", mb, log)
	lc := lk.(linkcache.ILinkCache)
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

	err = lc.Link("token01", nods[0])
	assert.Equal(t, err, nil)

	err = lc.Link("token01", nods[1])
	assert.Equal(t, err, nil)

	err = lc.Link("token02", nods[0])
	assert.Equal(t, err, nil)

	addr, err := lc.Target("token01", "base")
	assert.Equal(t, err, nil)
	assert.Equal(t, addr, "127.0.0.1:12001")

	lc.Unlink("token01", "")
	lc.Unlink("token02", "")

	for _, v := range nods {
		lc.Down(v)
	}

	time.Sleep(time.Millisecond * 500)
}

func TestDispatch(t *testing.T) {

	LinkerRedisPrefix = "braid_linker-"

}

func BenchmarkLink(b *testing.B) {
	LinkerRedisPrefix = "benchmarklink"

	log, _ := logger.GetBuilder(zaplogger.Name).Build()

	mbb := mailbox.GetBuilder(mailboxnsq.Name)
	mbb.AddOption(mailboxnsq.WithLookupAddr([]string{mock.NSQLookupdAddr}))
	mbb.AddOption(mailboxnsq.WithNsqdAddr([]string{mock.NsqdAddr}))
	mb, _ := mbb.Build("benchmarklink")

	eb := module.GetBuilder(electorconsul.Name)
	eb.AddOption(electorconsul.WithConsulAddr(mock.ConsulAddr))
	e, _ := eb.Build("testlinkertarget", mb, log)
	defer e.Close()

	lb := module.GetBuilder(Name)

	lk, err := lb.Build("gate", mb, log)
	lc := lk.(linkcache.ILinkCache)
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
		lc.Link("token"+strconv.Itoa(i), baseTargets[rand.Intn(len(baseTargets))])
		lc.Link("token"+strconv.Itoa(i), loginTargets[rand.Intn(len(loginTargets))])
	}
}
