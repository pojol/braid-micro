package braid

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/pojol/braid/3rd/redis"
	"github.com/pojol/braid/mock"
	"github.com/pojol/braid/module/mailbox"
	"github.com/pojol/braid/plugin/discoverconsul"
	"github.com/pojol/braid/plugin/electorconsul"
	"github.com/pojol/braid/plugin/grpcserver"
	"github.com/pojol/braid/plugin/linkerredis"
	"github.com/pojol/braid/plugin/mailboxnsq"
	"github.com/stretchr/testify/assert"
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

func TestServerInterface(t *testing.T) {
	s := Server()
	assert.Equal(t, s, nil)

	b, _ := New("testserverinterface")
	b.RegistPlugin(GRPCServer(
		grpcserver.Name,
		grpcserver.WithListen(":14222")))

	s = Server()
	assert.NotEqual(t, s, nil)
}

func TestMutiMailBox(t *testing.T) {

	New(
		"test_plugin",
		mailboxnsq.WithLookupAddr([]string{mock.NSQLookupdAddr}),
		mailboxnsq.WithNsqdAddr([]string{mock.NsqdAddr}),
	)
	topic := "TestMutiSharedProc"

	var wg sync.WaitGroup
	done := make(chan struct{})

	sub := Mailbox().Sub(mailbox.Proc, topic)
	c1, _ := sub.Shared()
	c1.OnArrived(func(msg *mailbox.Message) error {
		wg.Done()
		return nil
	})

	for i := 0; i < 1000; i++ {
		go func() {
			wg.Add(1)
			Mailbox().Pub(mailbox.Proc, topic, &mailbox.Message{Body: []byte("msg")})
		}()
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// pass
		fmt.Println("done")
	case <-time.After(time.Millisecond * 500):
		// time out
		t.FailNow()
	}
}
