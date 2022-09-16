package pubsubnsq

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pojol/braid-go/depend/blog"
	"github.com/pojol/braid-go/mock"
	"github.com/pojol/braid-go/module/pubsub"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {

	mock.Init()

	os.Exit(m.Run())
}

func TestClusterBroadcast(t *testing.T) {

	mb := BuildWithOption(
		"TestClusterBroadcast",
		blog.BuildWithOption(),
		pubsub.WithLookupAddr([]string{}),
		pubsub.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}),
	)

	topic := "test.clusterBroadcast"

	channel1 := mb.ClusterTopic(topic).Sub("Normal_1")
	channel2 := mb.ClusterTopic(topic).Sub("Normal_2")

	msgcnt := 10000

	var wg1, wg2 sync.WaitGroup
	done := make(chan struct{})

	wg1.Add(msgcnt)
	wg2.Add(msgcnt)

	channel1.Arrived(func(msg *pubsub.Message) {
		wg1.Done()
	})
	channel2.Arrived(func(msg *pubsub.Message) {
		wg2.Done()
	})

	go func() {
		wg1.Wait()
		wg2.Wait()
		close(done)
	}()

	for i := 0; i < msgcnt; i++ {
		mb.ClusterTopic(topic).Pub(&pubsub.Message{Body: []byte("test msg")})
	}

	select {
	case <-done:
		mb.RmvTopic(topic)
		// pass
	case <-time.After(time.Second * 5):
		fmt.Println("timeout")

		res, _ := http.Get("http://127.0.0.1:4151/stats")
		byt, _ := ioutil.ReadAll(res.Body)
		fmt.Println(string(byt))
		res.Body.Close()

		t.FailNow()
	}

}

func TestClusterNotify(t *testing.T) {

	mb := BuildWithOption(
		"TestClusterNotify",
		blog.BuildWithOption(),
		pubsub.WithLookupAddr([]string{}),
		pubsub.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}),
	)

	var tick uint64

	topic := "test.clusterNotify"

	channel1 := mb.ClusterTopic(topic).Sub("Normal")
	channel2 := mb.ClusterTopic(topic).Sub("Normal")

	channel1.Arrived(func(msg *pubsub.Message) {
		atomic.AddUint64(&tick, 1)
	})
	channel2.Arrived(func(msg *pubsub.Message) {
		atomic.AddUint64(&tick, 1)
	})

	mb.ClusterTopic(topic).Pub(&pubsub.Message{Body: []byte("msg")})

	for {
		<-time.After(time.Second * 3)
		assert.Equal(t, atomic.LoadUint64(&tick), uint64(1))
		break
	}

}

func TestClusterMutiNSQD(t *testing.T) {
	// 非常规测试，需要开启多个nsqd进程

	/*
		b := mailbox.GetBuilder(Name)
		log, _ := logger.GetBuilder(zaplogger.Name).Build()
		b.AddOption(WithLookupAddr([]string{"127.0.0.1:4161"}))
		b.AddOption(WithNsqdAddr([]string{"127.0.0.1:4150", "127.0.0.1:4152"}))
		b.AddOption(WithNsqdHTTPAddr([]string{"127.0.0.1:4151", "127.0.0.1:4153"}))

		topic := "test.clusterMutiNsqd"
		mb, _ := b.Build("TestClusterMutiNSQD", log)
		mb.RegistTopic(topic, mailbox.ScopeCluster)

		mb.GetTopic(topic).Sub("consumer_1").Arrived(func(msg *pubsub.Message) {
			fmt.Println("consumer a receive", string(msg.Body))
		})
		mb.GetTopic(topic).Sub("consumer_1").Arrived(func(msg *pubsub.Message) {
			fmt.Println("consumer b receive", string(msg.Body))
		})

		for i := 0; i < 10; i++ {
			mb.GetTopic(topic).Pub(&pubsub.Message{Body: []byte(strconv.Itoa(i))})
		}

		for {
			<-time.After(time.Second * 2)
			t.FailNow()
		}
	*/
}

func BenchmarkClusterBroadcast(b *testing.B) {

	mb := BuildWithOption(
		"BenchmarkClusterBroadcast",
		blog.BuildWithOption(),
		pubsub.WithLookupAddr([]string{}),
		pubsub.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}),
	)

	topic := "benchmark.ClusterBroadcast"
	body := []byte("msg")

	c1 := mb.ClusterTopic(topic).Sub("Normal_1")
	c2 := mb.ClusterTopic(topic).Sub("Normal_2")

	c1.Arrived(func(msg *pubsub.Message) {
	})
	c2.Arrived(func(msg *pubsub.Message) {
	})

	b.SetParallelism(8)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mb.ClusterTopic(topic).Pub(&pubsub.Message{Body: body})
		}
	})

}

/*
func TestClusterMailboxParm(t *testing.T) {
	b := mailbox.GetBuilder(Name)
	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	b.AddOption(WithChannel("parm"))
	b.AddOption(WithLookupAddr([]string{mock.NSQLookupdAddr}))
	b.AddOption(WithNsqdAddr([]string{mock.NsqdAddr}))

	mb, err := b.Build("cluster", log)
	assert.Equal(t, err, nil)
	cm := mb.(*nsqMailbox)

	assert.Equal(t, cm.parm.Channel, "parm")
	assert.Equal(t, cm.parm.LookupAddress, []string{mock.NSQLookupdAddr})
	assert.Equal(t, cm.parm.Address, []string{mock.NsqdAddr})
}
*/
