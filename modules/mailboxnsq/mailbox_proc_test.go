package mailboxnsq

import (
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
	c1.OnArrived(func(msg *mailbox.Message) error {
		wg.Done()
		return nil
	})

	c2, _ := sub.Shared()
	defer c2.Exit()
	c2.OnArrived(func(msg *mailbox.Message) error {
		wg.Done()
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
	c1.OnArrived(func(msg *mailbox.Message) error {
		race.Lock()
		atomic.AddUint64(&carrived, 1)
		race.Unlock()
		return nil
	})
	c2, _ := sub.Competition()
	c2.OnArrived(func(msg *mailbox.Message) error {
		race.Lock()
		atomic.AddUint64(&carrived, 1)
		race.Unlock()
		return nil
	})

	mb.Pub(mailbox.Proc, topic, &mailbox.Message{Body: []byte("msg")})
	time.Sleep(time.Millisecond * 500)

	race.Lock()
	assert.Equal(t, carrived, uint64(1))
	race.Unlock()
}

func BenchmarkShared(b *testing.B) {
	mbb := mailbox.GetBuilder(Name)
	mb, _ := mbb.Build("BenchmarkShared")
	topic := "BenchmarkShared"

	sub := mb.Sub(mailbox.Proc, topic)
	c1, _ := sub.Shared()
	c1.OnArrived(func(msg *mailbox.Message) error {
		return nil
	})
	c2, _ := sub.Shared()
	c2.OnArrived(func(msg *mailbox.Message) error {
		return nil
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mb.Pub(mailbox.Proc, topic, &mailbox.Message{Body: []byte("msg")})
	}
}

func BenchmarkCompetition(b *testing.B) {
	mbb := mailbox.GetBuilder(Name)
	mb, _ := mbb.Build("BenchmarkCompetition")
	topic := "BenchmarkCompetition"

	sub := mb.Sub(mailbox.Proc, topic)
	c1, _ := sub.Competition()
	c1.OnArrived(func(msg *mailbox.Message) error {
		return nil
	})
	c2, _ := sub.Competition()
	c2.OnArrived(func(msg *mailbox.Message) error {
		return nil
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mb.Pub(mailbox.Proc, topic, &mailbox.Message{Body: []byte("msg")})
	}
}