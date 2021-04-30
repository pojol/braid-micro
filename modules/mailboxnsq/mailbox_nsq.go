package mailboxnsq

import (
	"math/rand"
	"sync"

	"github.com/google/uuid"
	"github.com/nsqio/go-nsq"
	"github.com/pojol/braid-go/internal/braidsync"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/mailbox"
)

type consumerHandler struct {
	channel string
	c       *nsqConsumer
}

// Consumer 消费者
type nsqConsumer struct {
	consumer *nsq.Consumer

	exitCh *braidsync.Switch
	handle mailbox.HandlerFunc
	sync.Mutex

	connected bool

	lookupAddress []string
	channel       string
}

type nsqSubscriber struct {
	Channel string
	Topic   string

	log      logger.ILogger
	nsqLovLv nsq.LogLevel

	lookupAddress []string
	address       []string

	ephemeral   bool
	serviceName string

	consumer mailbox.IConsumer
}

func (ch *consumerHandler) HandleMessage(msg *nsq.Message) error {

	ch.c.PutMsg(&mailbox.Message{
		Body: msg.Body,
	})

	return nil
}

func (nmb *nsqMailbox) pubasync(topic string, msg *mailbox.Message) {
	nmb.pub(topic, msg)
}

func (nmb *nsqMailbox) pub(topic string, msg *mailbox.Message) error {
	p := nmb.cproducers[rand.Intn(len(nmb.cproducers))]
	return p.Publish(topic, msg.Body)
}

func (c *nsqConsumer) OnArrived(handle mailbox.HandlerFunc) error {
	c.Lock()
	defer c.Unlock()

	if c.handle == nil {
		c.handle = handle

		err := c.consumer.ConnectToNSQLookupds(c.lookupAddress)
		if err != nil {
			return err
		}

		c.connected = true
	}

	return nil
}

func (c *nsqConsumer) Exit() {
	c.exitCh.Open()
}

func (c *nsqConsumer) IsExited() bool {
	return false
}

func (c *nsqConsumer) PutMsgAsync(msg *mailbox.Message) {

}

func (c *nsqConsumer) PutMsg(msg *mailbox.Message) error {
	c.Lock()
	defer c.Unlock()

	return c.handle(*msg)
}

func (ns *nsqSubscriber) subImpl(channel string) (*nsqConsumer, error) {
	config := nsq.NewConfig()
	nc := &nsqConsumer{
		exitCh:        braidsync.NewSwitch(),
		channel:       channel,
		lookupAddress: ns.lookupAddress,
	}

	consumer, err := nsq.NewConsumer(ns.Topic, nc.channel, config)
	if err != nil {
		return nil, err
	}
	ns.log.Infof("new consumer topic:%v, channel:%v", ns.Topic, nc.channel)
	consumer.SetLoggerLevel(ns.nsqLovLv)
	consumer.AddHandler(&consumerHandler{
		c:       nc,
		channel: nc.channel,
	})

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
		nsqLovLv:      nmb.parm.nsqLogLv,
		log:           nmb.log,
	}

	nmb.csubsrcibers = append(nmb.csubsrcibers, s)

	return s
}
