package braid

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/pojol/braid/mock"
	"github.com/pojol/braid/module/mailbox"
	"github.com/pojol/braid/modules/discoverconsul"
	"github.com/pojol/braid/modules/electorconsul"
	"github.com/pojol/braid/modules/grpcserver"
	"github.com/pojol/braid/modules/linkerredis"
	"github.com/pojol/braid/modules/mailboxnsq"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {

	mock.Init()
	m.Run()
}

func TestPlugin(t *testing.T) {

	b, _ := New(
		"test_plugin",
		mailboxnsq.WithLookupAddr([]string{mock.NSQLookupdAddr}),
		mailboxnsq.WithNsqdAddr([]string{mock.NsqdAddr}),
	)

	b.RegistModule(
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

	b.Init()
	b.Run()
	defer b.Close()
}

func TestWithClient(t *testing.T) {
	/*
		b := New("test")
		b.RegistModule(DiscoverByConsul(mock.ConsulAddr, discoverconsul.WithInterval(time.Second*3)),
			BalancerBySwrr(),
			GRPCClient(grpcclient.WithPoolCapacity(128)))

		b.Run()
		defer b.Close()

		Client().Invoke(context.TODO(), "targeNodeName", "/proto.node/method", "", nil, nil)
	*/
}

func TestServerInterface(t *testing.T) {
	s := GetServer()
	assert.Equal(t, s, nil)

	b, _ := New("testserverinterface")
	b.RegistModule(Server(
		grpcserver.Name,
		grpcserver.WithListen(":14222")))

	s = GetServer()
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
	c1.OnArrived(func(msg mailbox.Message) error {
		wg.Done()
		return nil
	})

	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func() {
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
