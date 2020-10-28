package mailboxnsq

import (
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
	mb, _ := b.Build("testproc")
	var onarrived uint64
	topic := "testproc_topic"

	sub := mb.ProcSub(topic)
	c1, _ := sub.AddShared()
	c1.OnArrived(func(msg *mailbox.Message) error {
		atomic.AddUint64(&onarrived, 1)
		return nil
	})
	c2, _ := sub.AddShared()
	c2.OnArrived(func(msg *mailbox.Message) error {
		atomic.AddUint64(&onarrived, 1)
		return nil
	})

	mb.ProcPub(topic, &mailbox.Message{Body: []byte("msg")})
	time.Sleep(time.Millisecond * 500)

	assert.Equal(t, onarrived, uint64(2))
}

func TestCompetition(t *testing.T) {
	b := mailbox.GetBuilder(Name)
	mb, _ := b.Build("TestCompetition")
	var onarrived uint64
	topic := "testcompetition_topic"

	sub := mb.ProcSub(topic)
	c1, _ := sub.AddCompetition()
	c1.OnArrived(func(msg *mailbox.Message) error {
		atomic.AddUint64(&onarrived, 1)
		return nil
	})
	c2, _ := sub.AddCompetition()
	c2.OnArrived(func(msg *mailbox.Message) error {
		atomic.AddUint64(&onarrived, 1)
		return nil
	})

	mb.ProcPub(topic, &mailbox.Message{Body: []byte("msg")})
	time.Sleep(time.Millisecond * 500)

	assert.Equal(t, onarrived, uint64(1))
}

func BenchmarkShared(b *testing.B) {
	mbb := mailbox.GetBuilder(Name)
	mb, _ := mbb.Build("BenchmarkShared")
	var onarrived uint64
	topic := "benchmarkshared_topic"

	sub := mb.ProcSub(topic)
	c1, _ := sub.AddShared()
	c1.OnArrived(func(msg *mailbox.Message) error {
		atomic.AddUint64(&onarrived, 1)
		return nil
	})
	c2, _ := sub.AddShared()
	c2.OnArrived(func(msg *mailbox.Message) error {
		atomic.AddUint64(&onarrived, 1)
		return nil
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mb.ProcPub(topic, &mailbox.Message{Body: []byte("msg")})
	}
}

func BenchmarkCompetition(b *testing.B) {
	mbb := mailbox.GetBuilder(Name)
	mb, _ := mbb.Build("BenchmarkComptition")
	var onarrived uint64
	topic := "benchmarkcompetition_topic"

	sub := mb.ProcSub(topic)
	c1, _ := sub.AddCompetition()
	c1.OnArrived(func(msg *mailbox.Message) error {
		atomic.AddUint64(&onarrived, 1)
		return nil
	})
	c2, _ := sub.AddCompetition()
	c2.OnArrived(func(msg *mailbox.Message) error {
		atomic.AddUint64(&onarrived, 1)
		return nil
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mb.ProcPub(topic, &mailbox.Message{Body: []byte("msg")})
	}
}
