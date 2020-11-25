package mailboxnsq

import (
	"math/rand"
	"sync"

	"github.com/google/uuid"
	"github.com/nsqio/go-nsq"
	"github.com/pojol/braid/internal/braidsync"
	"github.com/pojol/braid/module/mailbox"
)

type consumerHandler struct {
	uuid string
	c    *nsqConsumer
}

// Consumer 消费者
type nsqConsumer struct {
	consumer *nsq.Consumer
	exitCh   *braidsync.Switch

	handle mailbox.HandlerFunc
	uuid   string

	lock sync.Mutex
}

type nsqSubscriber struct {
	Channel string
	Topic   string

	lookupAddress []string
	address       []string

	ephemeral   bool
	serviceName string

	consumer mailbox.IConsumer
	//group map[string]mailbox.IConsumer
	//sync.Mutex
}

func (ch *consumerHandler) HandleMessage(msg *nsq.Message) error {
	ch.c.lock.Lock()
	defer ch.c.lock.Unlock()

	err := ch.c.PutMsg(&mailbox.Message{
		Body: msg.Body,
	})
	if err != nil {
		return err
	}

	return nil
}

func (nmb *nsqMailbox) pub(topic string, msg *mailbox.Message) {
	p := nmb.cproducers[rand.Intn(len(nmb.cproducers))]
	p.Publish(topic, msg.Body)
}

/*
func (ns *nsqSubscriber) GetConsumer(cid string) []mailbox.IConsumer {
	c := []mailbox.IConsumer{}

	if _, ok := ns.group[cid]; ok {
		c = append(c, ns.group[cid])
	}

	return c
}
*/
func (c *nsqConsumer) OnArrived(handler mailbox.HandlerFunc) {
	c.lock.Lock()
	c.handle = handler
	c.lock.Unlock()
}

func (c *nsqConsumer) Exit() {
	c.exitCh.Done()
}

func (c *nsqConsumer) IsExited() bool {
	return false
}

func (c *nsqConsumer) PutMsg(msg *mailbox.Message) error {
	return c.handle(*msg)
}

func (ns *nsqSubscriber) subImpl(channel string) (*nsqConsumer, error) {
	config := nsq.NewConfig()
	nc := &nsqConsumer{
		exitCh: braidsync.NewSwitch(),
		uuid:   channel,
	}

	consumer, err := nsq.NewConsumer(ns.Topic, nc.uuid, config)
	if err != nil {
		return nil, err
	}

	consumer.AddHandler(&consumerHandler{
		c:    nc,
		uuid: nc.uuid,
	})

	err = consumer.ConnectToNSQLookupds(ns.lookupAddress)
	if err != nil {
		return nil, err
	}

	nc.consumer = consumer

	return nc, nil
}

// AddCompetition 从固定的管道中竞争消息
func (ns *nsqSubscriber) Competition() (mailbox.IConsumer, error) {

	if ns.Channel == "" {
		ns.Channel = ns.serviceName + "-" + "competition"
	}

	nc, err := ns.subImpl(ns.Channel)
	if err != nil {
		return nil, err
	}
	ns.consumer = nc

	return nc, nil
}

// AddShared 从管道副本中一起消费消息，因为共享需要不同的管道，所以这里默认设置为ephemeral
func (ns *nsqSubscriber) Shared() (mailbox.IConsumer, error) {

	uid := ns.serviceName + "-" + uuid.New().String() + "#ephemeral"

	nc, err := ns.subImpl(uid)
	if err != nil {
		return nil, err
	}
	ns.consumer = nc

	return nc, nil
}

func (nmb *nsqMailbox) sub(topic string) mailbox.ISubscriber {

	s := &nsqSubscriber{
		Channel:       nmb.parm.Channel,
		Topic:         topic,
		lookupAddress: nmb.parm.LookupAddress,
		address:       nmb.parm.Address,
		serviceName:   nmb.parm.ServiceName,
	}

	nmb.csubsrcibers = append(nmb.csubsrcibers, s)

	return s
}
