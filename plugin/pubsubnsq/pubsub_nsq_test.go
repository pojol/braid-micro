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
	fmt.Println("nsqd addr", mock.NsqdAddr)
	fmt.Println("lookupd addr", mock.NSQLookupdAddr)

	m.Run()
}

func TestNsq1V1(t *testing.T) {

	topic := "test_nsq"
	msgByte := []byte("test msg 1")
	onarrived := false

	psb := pubsub.GetBuilder(PubsubName)
	psb.SetCfg(NsqConfig{
		Addres:       []string{mock.NsqdAddr},
		LookupAddres: []string{mock.NSQLookupdAddr},
	})
	pb, err := psb.Build()
	assert.Equal(t, err, nil)

	consumer := pb.Sub(topic).AddShared()
	defer consumer.Exit()

	go func() {
		pb.Pub(topic, &pubsub.Message{Body: msgByte})
	}()

	consumer.OnArrived(func(msg *pubsub.Message) error {
		fmt.Println("test nsq 1v1 , msg arrived", string(msgByte))
		onarrived = true
		return nil
	})

	time.Sleep(time.Millisecond * 1000)
	assert.Equal(t, onarrived, true)
}

func TestNsqShared(t *testing.T) {
	topic := "test_shared_nsq"
	msgByte := []byte("test msg 1")
	var onarrived uint64

	psb := pubsub.GetBuilder(PubsubName)
	psb.SetCfg(NsqConfig{
		Addres:       []string{mock.NsqdAddr},
		LookupAddres: []string{mock.NSQLookupdAddr},
	})
	pb, err := psb.Build()
	assert.Equal(t, err, nil)

	consumer1 := pb.Sub(topic).AddShared()
	defer consumer1.Exit()

	consumer2 := pb.Sub(topic).AddShared()
	defer consumer2.Exit()

	go func() {
		pb.Pub(topic, &pubsub.Message{Body: msgByte})
	}()

	consumer1.OnArrived(func(msg *pubsub.Message) error {
		fmt.Println("TestNsqShared consumer 1 msg arrived", string(msgByte))

		atomic.AddUint64(&onarrived, 1)
		return nil
	})

	consumer2.OnArrived(func(msg *pubsub.Message) error {
		fmt.Println("TestNsqShared consumer 2 msg arrived", string(msgByte))

		atomic.AddUint64(&onarrived, 1)
		return nil
	})

	time.Sleep(time.Millisecond * 1000)
	assert.Equal(t, onarrived, uint64(2))
}

func TestNsqCompetition(t *testing.T) {
	topic := "test_competion_nsq"
	msgByte := []byte("test msg 1")
	var onarrived uint64

	psb := pubsub.GetBuilder(PubsubName)
	psb.SetCfg(NsqConfig{
		Channel:      "competition1",
		Addres:       []string{mock.NsqdAddr},
		LookupAddres: []string{mock.NSQLookupdAddr},
	})
	pb, err := psb.Build()
	assert.Equal(t, err, nil)

	go func() {
		pb.Pub(topic, &pubsub.Message{Body: msgByte})
	}()

	consumer1 := pb.Sub(topic).AddCompetition()
	defer consumer1.Exit()

	consumer2 := pb.Sub(topic).AddCompetition()
	defer consumer2.Exit()

	consumer1.OnArrived(func(msg *pubsub.Message) error {
		fmt.Println("TestNsqCompetition consumer 1 msg arrived", string(msgByte))

		atomic.AddUint64(&onarrived, 1)
		return nil
	})

	consumer2.OnArrived(func(msg *pubsub.Message) error {
		fmt.Println("TestNsqCompetition consumer 2 msg arrived", string(msgByte))

		atomic.AddUint64(&onarrived, 1)
		return nil
	})

	time.Sleep(time.Millisecond * 1000)
	assert.Equal(t, onarrived, uint64(1))
}
