package braid

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/pojol/braid-go/mock"
	"github.com/pojol/braid-go/module/pubsub"
	"github.com/pojol/braid-go/modules/discoverconsul"
	"github.com/pojol/braid-go/modules/electorconsul"
	"github.com/pojol/braid-go/modules/linkerredis"
	"github.com/pojol/braid-go/modules/pubsubnsq"
	"github.com/pojol/braid-go/modules/zaplogger"
)

func TestMain(m *testing.M) {

	mock.Init()
	m.Run()
}

func TestPlugin(t *testing.T) {

	b, _ := NewService(
		"test_plugin",
	)

	b.Register(
		Module(LoggerZap),
		Module(PubsubNsq,
			pubsubnsq.WithLookupAddr([]string{mock.NSQLookupdAddr}),
			pubsubnsq.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}),
		),
		Module(linkerredis.Name, linkerredis.WithRedisAddr(mock.RedisAddr)),
		Module(discoverconsul.Name, discoverconsul.WithConsulAddr(mock.ConsulAddr)),
		Module(electorconsul.Name,
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

func TestMutiPubsub(t *testing.T) {

	b, _ := NewService(
		"test_plugin",
	)
	b.Register(
		Module(zaplogger.Name),
		Module(pubsubnsq.Name,
			pubsubnsq.WithLookupAddr([]string{mock.NSQLookupdAddr}),
			pubsubnsq.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}),
		),
	)

	var wg sync.WaitGroup
	done := make(chan struct{})

	Pubsub().RegistTopic("TestMutiPubsub", pubsub.ScopeProc)

	topic := Pubsub().GetTopic("TestMutiPubsub")
	c1 := topic.Sub("Normal")

	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func() {
			topic.Pub(&pubsub.Message{Body: []byte("msg")})
		}()
	}

	c1.Arrived(func(msg *pubsub.Message) {
		wg.Done()
	})

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// pass
		fmt.Println("done")
	case <-time.After(time.Second):
		// time out
		fmt.Println("timeout")
		t.FailNow()
	}
}
