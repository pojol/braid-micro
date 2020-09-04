package pubsubnsq

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"

	"github.com/google/uuid"
	"github.com/nsqio/go-nsq"
	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/internal/braidsync"
	"github.com/pojol/braid/module/pubsub"
)

const (
	// PubsubName 进程内的消息通知
	PubsubName = "NsqPubsub"
)

var (
	// ErrConfigConvert 配置转换失败
	ErrConfigConvert = errors.New("convert config error")
)

type nsqPubsubBuilder struct {
	cfg NsqConfig
}

func newNsqPubsub() pubsub.Builder {
	return &nsqPubsubBuilder{}
}

func (pb *nsqPubsubBuilder) Build() (pubsub.IPubsub, error) {

	producers := make([]*nsq.Producer, 0, len(pb.cfg.Addres))
	for _, addr := range pb.cfg.Addres {
		producer, err := nsq.NewProducer(addr, nsq.NewConfig())
		if err != nil {
			return nil, err
		}

		if err = producer.Ping(); err != nil {
			return nil, err
		}

		log.Debugf("nsq producer build succ %s", addr)
		producers = append(producers, producer)
	}

	fmt.Println("p", producers)
	ps := &nsqPubsub{
		producers: producers,
		cfg:       pb.cfg,
	}

	return ps, nil
}

func (*nsqPubsubBuilder) Name() string {
	return PubsubName
}

func (pb *nsqPubsubBuilder) SetCfg(cfg interface{}) error {

	nsqCfg, ok := cfg.(NsqConfig)
	if !ok {
		return ErrConfigConvert
	}

	pb.cfg = nsqCfg

	return nil
}

type nsqPubsub struct {
	producers   []*nsq.Producer
	subsrcibers []*nsqSubscriber

	cfg NsqConfig
}

// Consumer 消费者
type nsqConsumer struct {
	consumer *nsq.Consumer
	uuid     string

	buff   *braidsync.Unbounded
	exitCh *braidsync.Switch
}

type nsqSubscriber struct {
	Channel string
	Topic   string

	lookupAddres []string
	addres       []string

	group map[string]pubsub.IConsumer
	sync.Mutex
}

type consumerHandler struct {
	uuid string
	ns   *nsqSubscriber
}

func (ch *consumerHandler) HandleMessage(msg *nsq.Message) error {

	consumerLst := ch.ns.GetConsumer(ch.uuid)
	for _, v := range consumerLst {

		// 这里不能异步消费消息（因为不能表达出消费失败，将消息回退的逻辑）待修改。
		// 如果消息执行到这里节点宕机，nsq可以将这个消息重新塞入到队列。
		v.PutMsg(&pubsub.Message{
			Body: msg.Body,
		})
	}

	return nil
}

func (ns *nsqSubscriber) addImpl(channel string) *nsqConsumer {

	config := nsq.NewConfig()
	nc := &nsqConsumer{
		buff:   braidsync.NewUnbounded(),
		exitCh: braidsync.NewSwitch(),
		uuid:   channel,
	}

	consumer, err := nsq.NewConsumer(ns.Topic, nc.uuid, config)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	consumer.AddHandler(&consumerHandler{
		ns:   ns,
		uuid: nc.uuid,
	})

	err = consumer.ConnectToNSQLookupds(ns.lookupAddres)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	nc.consumer = consumer

	return nc
}

func (ns *nsqSubscriber) AddCompetition() pubsub.IConsumer {

	ns.Lock()
	defer ns.Unlock()

	if ns.Channel == "" {
		return nil
	}

	nc := ns.addImpl(ns.Channel)
	ns.group[nc.uuid] = nc

	return nc
}

func (ns *nsqSubscriber) AddShared() pubsub.IConsumer {

	ns.Lock()

	defer ns.Unlock()
	nc := ns.addImpl(uuid.New().String() + "#ephemeral")

	ns.group[nc.uuid] = nc

	return nc

}

func (ns *nsqSubscriber) GetConsumer(cid string) []pubsub.IConsumer {
	c := []pubsub.IConsumer{}

	if _, ok := ns.group[cid]; ok {
		c = append(c, ns.group[cid])
	}

	return c
}

func (c *nsqConsumer) OnArrived(handler pubsub.HandlerFunc) {
	go func() {
		for {
			select {
			case msg := <-c.buff.Get():
				handler(msg.(*pubsub.Message))
				c.buff.Load()
			case <-c.exitCh.Done():
			}

			if c.exitCh.HasOpend() {
				c.consumer.Stop()
				return
			}
		}
	}()
}

func (c *nsqConsumer) Exit() {
	c.exitCh.Done()
}

func (c *nsqConsumer) IsExited() bool {
	return false
}

func (c *nsqConsumer) PutMsg(msg *pubsub.Message) {
	c.buff.Put(msg)
}

func (kps *nsqPubsub) Sub(topic string) pubsub.ISubscriber {
	s := &nsqSubscriber{
		group:        make(map[string]pubsub.IConsumer),
		Channel:      kps.cfg.Channel,
		Topic:        topic,
		lookupAddres: kps.cfg.LookupAddres,
		addres:       kps.cfg.Addres,
	}

	kps.subsrcibers = append(kps.subsrcibers, s)

	return s
}

func (kps *nsqPubsub) Pub(topic string, msg *pubsub.Message) {

	p := kps.producers[rand.Intn(len(kps.producers))]
	p.Publish(topic, msg.Body)

}

func init() {
	pubsub.Register(newNsqPubsub())
}
