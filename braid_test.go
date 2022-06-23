package braid

import (
	"fmt"
	"sync"
	"testing"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/pojol/braid-go/depend/blog"
	"github.com/pojol/braid-go/depend/pubsub"
	"github.com/pojol/braid-go/depend/redis"
	"github.com/pojol/braid-go/depend/tracer"
	"github.com/pojol/braid-go/mock"
	"github.com/pojol/braid-go/module/elector"
	"github.com/pojol/braid-go/module/linkcache"
	"github.com/pojol/braid-go/rpc/client"
)

func TestMain(m *testing.M) {

	mock.Init()
	m.Run()
}

func TestInit(t *testing.T) {

	b, _ := NewService(
		"test_init",
	)

	b.RegisterDepend(
		blog.BuildWithNormal(),
		redis.BuildWithDefault(),
		pubsub.BuildWithOption(
			b.name,
			pubsub.WithLookupAddr([]string{mock.NSQLookupdAddr}),
			pubsub.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}),
		),
		tracer.BuildWithOption(
			b.name,
			tracer.WithHTTP(mock.JaegerAddr),
			tracer.WithProbabilistic(0.1),
		),
	)

	b.RegisterClient(
		client.AppendInterceptors(grpc_prometheus.UnaryClientInterceptor),
	)

	b.RegisterModule(
		b.Discover(),
		b.Elector(elector.WithLockTick(3*time.Second)),
		b.LinkCache(linkcache.WithMode(linkcache.LinkerRedisModeLocal)),
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

	b.Init()
	b.Run()

	var wg sync.WaitGroup
	done := make(chan struct{})

	Pubsub().GetTopic("TestMutiPubsub")

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
