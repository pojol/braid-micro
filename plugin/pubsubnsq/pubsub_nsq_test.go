package pubsubnsq

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pojol/braid/mock"
	"github.com/pojol/braid/module/pubsub"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {

	mock.Init()

	m.Run()
}

func TestNsq1V1(t *testing.T) {

	topic := "test_nsq"
	msgByte := []byte("test msg 1")
	onarrived := false

	psb := pubsub.GetBuilder(PubsubName)
	psb.SetCfg(NsqConfig{
		addres:       []string{mock.NsqdAddr},
		lookupAddres: []string{mock.NSQLookupdAddr},
	})
	pb, err := psb.Build()
	assert.Equal(t, err, nil)

	consumer := pb.Sub(topic).AddShared()
	defer consumer.Exit()

	consumer.OnArrived(func(msg *pubsub.Message) error {
		assert.Equal(t, string(msg.Body), string(msgByte))
		fmt.Println("test nsq 1v1 , msg arrived", string(msgByte))
		onarrived = true
		return nil
	})

	go func() {
		pb.Pub(topic, &pubsub.Message{Body: msgByte})
	}()

	time.Sleep(time.Millisecond * 1000)
	assert.Equal(t, onarrived, true)
}

func TestNsqShared(t *testing.T) {
	topic := "test_nsq"
	msgByte := []byte("test msg 1")
	var onarrived uint64

	psb := pubsub.GetBuilder(PubsubName)
	psb.SetCfg(NsqConfig{
		addres:       []string{mock.NsqdAddr},
		lookupAddres: []string{mock.NSQLookupdAddr},
	})
	pb, err := psb.Build()
	assert.Equal(t, err, nil)

	consumer1 := pb.Sub(topic).AddShared()
	defer consumer1.Exit()

	consumer2 := pb.Sub(topic).AddShared()
	defer consumer2.Exit()

	consumer1.OnArrived(func(msg *pubsub.Message) error {
		fmt.Println("consumer 1 msg arrived", string(msgByte))

		atomic.AddUint64(&onarrived, 1)
		return nil
	})

	consumer2.OnArrived(func(msg *pubsub.Message) error {
		fmt.Println("consumer 2 msg arrived", string(msgByte))

		atomic.AddUint64(&onarrived, 1)
		return nil
	})

	go func() {
		pb.Pub(topic, &pubsub.Message{Body: msgByte})
	}()

	time.Sleep(time.Millisecond * 1000)
	assert.Equal(t, onarrived, uint64(2))
}

func TestNsqCompetition(t *testing.T) {
	topic := "test_nsq"
	msgByte := []byte("test msg 1")
	var onarrived uint64

	psb := pubsub.GetBuilder(PubsubName)
	psb.SetCfg(NsqConfig{
		Channel:      "competition1",
		addres:       []string{mock.NsqdAddr},
		lookupAddres: []string{mock.NSQLookupdAddr},
	})
	pb, err := psb.Build()
	assert.Equal(t, err, nil)

	consumer1 := pb.Sub(topic).AddCompetition()
	defer consumer1.Exit()

	consumer2 := pb.Sub(topic).AddCompetition()
	defer consumer2.Exit()

	consumer1.OnArrived(func(msg *pubsub.Message) error {
		fmt.Println("consumer 1 msg arrived", string(msgByte))

		atomic.AddUint64(&onarrived, 1)
		return nil
	})

	consumer2.OnArrived(func(msg *pubsub.Message) error {
		fmt.Println("consumer 2 msg arrived", string(msgByte))

		atomic.AddUint64(&onarrived, 1)
		return nil
	})

	go func() {
		pb.Pub(topic, &pubsub.Message{Body: msgByte})
	}()

	time.Sleep(time.Millisecond * 1000)
	assert.Equal(t, onarrived, uint64(1))
}
