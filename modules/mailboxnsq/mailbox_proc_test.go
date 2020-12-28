package mailboxnsq

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pojol/braid/mock"
	"github.com/pojol/braid/module/mailbox"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {

	mock.Init()

	m.Run()
}

func TestSharedProc(t *testing.T) {

	b := mailbox.GetBuilder(Name)
	mb, _ := b.Build("TestSharedProc")
	topic := "TestSharedProc"

	var wg sync.WaitGroup
	done := make(chan struct{})

	sub := mb.Sub(mailbox.Proc, topic)
	c1, _ := sub.Shared()
	defer c1.Exit()
	c2, _ := sub.Shared()
	defer c2.Exit()

	c1.OnArrived(func(msg mailbox.Message) error {
		wg.Done()
		fmt.Printf("%p %s\n", &msg, msg.Body)
		return nil
	})

	c2.OnArrived(func(msg mailbox.Message) error {
		wg.Done()
		fmt.Printf("%p %s\n", &msg, msg.Body)
		return nil
	})

	wg.Add(2)
	mb.Pub(mailbox.Proc, topic, &mailbox.Message{Body: []byte("msg")})

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		//pass
	case <-time.After(time.Millisecond * 500):
		t.FailNow()
	}

}

func TestCompetition(t *testing.T) {
	b := mailbox.GetBuilder(Name)
	mb, _ := b.Build("TestCompetition")
	var carrived uint64
	var race sync.Mutex
	topic := "TestCompetition"

	sub := mb.Sub(mailbox.Proc, topic)
	c1, _ := sub.Competition()
	c2, _ := sub.Competition()

	c1.OnArrived(func(msg mailbox.Message) error {
		race.Lock()
		atomic.AddUint64(&carrived, 1)
		race.Unlock()
		return nil
	})

	c2.OnArrived(func(msg mailbox.Message) error {
		race.Lock()
		atomic.AddUint64(&carrived, 1)
		race.Unlock()
		return nil
	})

	mb.Pub(mailbox.Proc, topic, &mailbox.Message{Body: []byte("msg")})
	time.Sleep(time.Second)

	race.Lock()
	assert.Equal(t, carrived, uint64(1))
	race.Unlock()
}

// BenchmarkShared-8   	 2102298	       527 ns/op	      94 B/op	       2 allocs/op
func BenchmarkShared(b *testing.B) {
	mbb := mailbox.GetBuilder(Name)
	mb, _ := mbb.Build("BenchmarkShared")
	topic := "BenchmarkShared"
	body := []byte("msg")

	sub := mb.Sub(mailbox.Proc, topic)
	c1, _ := sub.Shared()
	c2, _ := sub.Shared()

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

func BenchmarkSharedAsync(b *testing.B) {
	mbb := mailbox.GetBuilder(Name)
	mb, _ := mbb.Build("BenchmarkShared")
	topic := "BenchmarkShared"
	body := []byte("msg")

	sub := mb.Sub(mailbox.Proc, topic)
	c1, _ := sub.Shared()
	c2, _ := sub.Shared()

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

//BenchmarkCompetition-8   	 3238792	       335 ns/op	      79 B/op	       2 allocs/op
func BenchmarkCompetition(b *testing.B) {
	mbb := mailbox.GetBuilder(Name)
	mb, _ := mbb.Build("BenchmarkCompetition")
	topic := "BenchmarkCompetition"
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

func BenchmarkCompetitionAsync(b *testing.B) {
	mbb := mailbox.GetBuilder(Name)
	mb, _ := mbb.Build("BenchmarkCompetition")
	topic := "BenchmarkCompetition"
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
