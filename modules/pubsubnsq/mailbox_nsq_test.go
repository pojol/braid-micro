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

	"github.com/pojol/braid-go/mock"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/pubsub"
	"github.com/pojol/braid-go/modules/moduleparm"
	"github.com/pojol/braid-go/modules/zaplogger"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {

	mock.Init()

	os.Exit(m.Run())
}

func TestClusterBroadcast(t *testing.T) {

	log := module.GetBuilder(zaplogger.Name).Build("TestProcNotify").(logger.ILogger)
	mbb := module.GetBuilder(Name)
	mbb.AddModuleOption(WithLookupAddr([]string{}))
	mbb.AddModuleOption(WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}))
	mb := mbb.Build("TestProcNotify", moduleparm.WithLogger(log)).(pubsub.IPubsub)

	topic := "test.clusterBroadcast"

	mb.RegistTopic(topic, pubsub.ScopeCluster)

	channel1 := mb.GetTopic(topic).Sub("Normal_1")
	channel2 := mb.GetTopic(topic).Sub("Normal_2")

	var wg sync.WaitGroup
	done := make(chan struct{})
	wg.Add(2)

	channel1.Arrived(func(msg *pubsub.Message) {
		wg.Done()
	})
	channel2.Arrived(func(msg *pubsub.Message) {
		wg.Done()
	})

	go func() {
		wg.Wait()
		close(done)
	}()

	mb.GetTopic(topic).Pub(&pubsub.Message{Body: []byte("test msg")})

	select {
	case <-done:
		mb.RemoveTopic(topic)
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

	log := module.GetBuilder(zaplogger.Name).Build("TestClusterNotify").(logger.ILogger)
	mbb := module.GetBuilder(Name)
	mbb.AddModuleOption(WithLookupAddr([]string{}))
	mbb.AddModuleOption(WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}))
	mb := mbb.Build("TestClusterNotify", moduleparm.WithLogger(log)).(pubsub.IPubsub)

	var tick uint64

	topic := "test.clusterNotify"

	mb.RegistTopic(topic, pubsub.ScopeCluster)

	channel1 := mb.GetTopic(topic).Sub("Normal")
	channel2 := mb.GetTopic(topic).Sub("Normal")

	channel1.Arrived(func(msg *pubsub.Message) {
		atomic.AddUint64(&tick, 1)
	})
	channel2.Arrived(func(msg *pubsub.Message) {
		atomic.AddUint64(&tick, 1)
	})

	mb.GetTopic(topic).Pub(&pubsub.Message{Body: []byte("msg")})

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
	log := module.GetBuilder(zaplogger.Name).Build("BenchmarkClusterBroadcast").(logger.ILogger)
	mbb := module.GetBuilder(Name)
	mbb.AddModuleOption(WithLookupAddr([]string{}))
	mbb.AddModuleOption(WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}))
	mb := mbb.Build("BenchmarkClusterBroadcast", moduleparm.WithLogger(log)).(pubsub.IPubsub)

	topic := "benchmark.ClusterBroadcast"
	body := []byte("msg")

	mb.RegistTopic(topic, pubsub.ScopeCluster)

	c1 := mb.GetTopic(topic).Sub("Normal_1")
	c2 := mb.GetTopic(topic).Sub("Normal_2")

	c1.Arrived(func(msg *pubsub.Message) {
	})
	c2.Arrived(func(msg *pubsub.Message) {
	})

	b.SetParallelism(8)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mb.GetTopic(topic).Pub(&pubsub.Message{Body: body})
		}
	})

}

/*
//BenchmarkCompetition-8   	 3238792	       335 ns/op	      79 B/op	       2 allocs/op
func BenchmarkProcCompetition(b *testing.B) {
	mbb := mailbox.GetBuilder(Name)
	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	mb, _ := mbb.Build("BenchmarkProcCompetition", log)
	topic := "BenchmarkProcCompetition"
	body := []byte("msg")

	sub := mb.Sub(mailbox.Proc, topic)
	c1, _ := sub.Competition()
	c2, _ := sub.Competition()

	c1.OnArrived(func(msg pubsub.Message) error {
		return nil
	})

	c2.OnArrived(func(msg pubsub.Message) error {
		return nil
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mb.Pub(mailbox.Proc, topic, &pubsub.Message{Body: body})
	}
}

func BenchmarkProcCompetitionAsync(b *testing.B) {
	mbb := mailbox.GetBuilder(Name)
	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	mb, _ := mbb.Build("BenchmarkProcCompetitionAsync", log)
	topic := "BenchmarkProcCompetitionAsync"
	body := []byte("msg")

	sub := mb.Sub(mailbox.Proc, topic)
	c1, _ := sub.Competition()
	c2, _ := sub.Competition()

	c1.OnArrived(func(msg pubsub.Message) error {
		return nil
	})

	c2.OnArrived(func(msg pubsub.Message) error {
		return nil
	})

	b.SetParallelism(8)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mb.PubAsync(mailbox.Proc, topic, &pubsub.Message{Body: body})
		}
	})
}




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
