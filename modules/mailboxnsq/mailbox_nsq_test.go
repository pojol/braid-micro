package mailboxnsq

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
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/mailbox"
	"github.com/pojol/braid-go/modules/zaplogger"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {

	mock.Init()

	os.Exit(m.Run())
}

func TestClusterBroadcast(t *testing.T) {

	b := mailbox.GetBuilder(Name)
	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	b.AddOption(WithLookupAddr([]string{}))
	b.AddOption(WithNsqdAddr([]string{mock.NsqdAddr}))
	b.AddOption(WithNsqdHTTPAddr([]string{mock.NsqdHttpAddr}))
	mb, _ := b.Build("TestClusterBroadcast", log)

	topic := "test.clusterBroadcast"

	mb.RegistTopic(topic, mailbox.ScopeCluster)

	channel1 := mb.GetTopic(topic).Sub("Normal_1")
	channel2 := mb.GetTopic(topic).Sub("Normal_2")

	var wg sync.WaitGroup
	done := make(chan struct{})
	wg.Add(2)

	channel1.Arrived(func(msg *mailbox.Message) {
		wg.Done()
	})
	channel2.Arrived(func(msg *mailbox.Message) {
		wg.Done()
	})

	go func() {
		wg.Wait()
		close(done)
	}()

	mb.GetTopic(topic).Pub(&mailbox.Message{Body: []byte("test msg")})

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

	b := mailbox.GetBuilder(Name)
	log, _ := logger.GetBuilder(zaplogger.Name).Build()
	b.AddOption(WithLookupAddr([]string{}))
	b.AddOption(WithNsqdAddr([]string{mock.NsqdAddr}))
	b.AddOption(WithNsqdHTTPAddr([]string{mock.NsqdHttpAddr}))
	mb, _ := b.Build("TestClusterNotify", log)

	var tick uint64

	topic := "test.clusterNotify"

	mb.RegistTopic(topic, mailbox.ScopeCluster)

	channel1 := mb.GetTopic(topic).Sub("Normal")
	channel2 := mb.GetTopic(topic).Sub("Normal")

	channel1.Arrived(func(msg *mailbox.Message) {
		atomic.AddUint64(&tick, 1)
	})
	channel2.Arrived(func(msg *mailbox.Message) {
		atomic.AddUint64(&tick, 1)
	})

	mb.GetTopic(topic).Pub(&mailbox.Message{Body: []byte("msg")})

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

		mb.GetTopic(topic).Sub("consumer_1").Arrived(func(msg *mailbox.Message) {
			fmt.Println("consumer a receive", string(msg.Body))
		})
		mb.GetTopic(topic).Sub("consumer_1").Arrived(func(msg *mailbox.Message) {
			fmt.Println("consumer b receive", string(msg.Body))
		})

		for i := 0; i < 10; i++ {
			mb.GetTopic(topic).Pub(&mailbox.Message{Body: []byte(strconv.Itoa(i))})
		}

		for {
			<-time.After(time.Second * 2)
			t.FailNow()
		}
	*/
}

func BenchmarkClusterBroadcast(b *testing.B) {
	log, _ := logger.GetBuilder(zaplogger.Name).Build()

	mbb := mailbox.GetBuilder(Name)
	mbb.AddOption(WithLookupAddr([]string{mock.NSQLookupdAddr}))
	mbb.AddOption(WithNsqdAddr([]string{mock.NsqdAddr}))
	mbb.AddOption(WithNsqdHTTPAddr([]string{mock.NsqdHttpAddr}))

	mb, _ := mbb.Build("BenchmarkClusterBroadcast", log)
	topic := "benchmark.ClusterBroadcast"
	body := []byte("msg")

	mb.RegistTopic(topic, mailbox.ScopeCluster)

	c1 := mb.GetTopic(topic).Sub("Normal_1")
	c2 := mb.GetTopic(topic).Sub("Normal_2")

	c1.Arrived(func(msg *mailbox.Message) {
	})
	c2.Arrived(func(msg *mailbox.Message) {
	})

	b.SetParallelism(8)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mb.GetTopic(topic).Pub(&mailbox.Message{Body: body})
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

	c1.OnArrived(func(msg mailbox.Message) error {
		return nil
	})

	c2.OnArrived(func(msg mailbox.Message) error {
		return nil
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mb.Pub(mailbox.Proc, topic, &mailbox.Message{Body: body})
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

	c1.OnArrived(func(msg mailbox.Message) error {
		return nil
	})

	c2.OnArrived(func(msg mailbox.Message) error {
		return nil
	})

	b.SetParallelism(8)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mb.PubAsync(mailbox.Proc, topic, &mailbox.Message{Body: body})
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
