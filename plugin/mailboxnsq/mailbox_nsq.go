package mailboxnsq

import (
	"math/rand"
	"sync"

	"github.com/google/uuid"
	"github.com/nsqio/go-nsq"
	"github.com/pojol/braid/internal/braidsync"
	"github.com/pojol/braid/module/mailbox"
)

// Consumer 消费者
type nsqConsumer struct {
	consumer *nsq.Consumer
	exitCh   *braidsync.Switch

	handle mailbox.HandlerFunc
	uuid   string

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

		err := v.PutMsg(&mailbox.Message{
			Body: msg.Body,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

type nsqSubscriber struct {
	Channel string
	Topic   string

	lookupAddres []string
	addres       []string

	ephemeral   bool
	serviceName string

	group map[string]mailbox.IConsumer
	sync.Mutex
}

func (nmb *nsqMailbox) ClusterPub(topic string, msg *mailbox.Message) {
	p := nmb.cproducers[rand.Intn(len(nmb.cproducers))]
	p.Publish(topic, msg.Body)
}

func (ns *nsqSubscriber) GetConsumer(cid string) []mailbox.IConsumer {
	c := []mailbox.IConsumer{}

	if _, ok := ns.group[cid]; ok {
		c = append(c, ns.group[cid])
	}

	return c
}

func (c *nsqConsumer) OnArrived(handler mailbox.HandlerFunc) {
	c.handle = handler
}

func (c *nsqConsumer) Exit() {
	c.exitCh.Done()
}

func (c *nsqConsumer) IsExited() bool {
	return false
}

func (c *nsqConsumer) PutMsg(msg *mailbox.Message) error {
	c.Lock()
	defer c.Unlock()

	return c.handle(msg)
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
		ns:   ns,
		uuid: nc.uuid,
	})

	err = consumer.ConnectToNSQLookupds(ns.lookupAddres)
	if err != nil {
		return nil, err
	}

	nc.consumer = consumer

	return nc, nil
}

// AddCompetition 从固定的管道中竞争消息
func (ns *nsqSubscriber) AddCompetition() (mailbox.IConsumer, error) {
	ns.Lock()
	defer ns.Unlock()

	if ns.Channel == "" {
		ns.Channel = ns.serviceName + "-" + "competition"
	}

	nc, err := ns.subImpl(ns.Channel)
	if err != nil {
		return nil, err
	}
	ns.group[nc.uuid] = nc

	return nc, nil
}

// AddShared 从管道副本中一起消费消息，因为共享需要不同的管道，所以这里默认设置为ephemeral
func (ns *nsqSubscriber) AddShared() (mailbox.IConsumer, error) {
	ns.Lock()
	defer ns.Unlock()

	uid := ns.serviceName + "-" + uuid.New().String() + "#ephemeral"

	nc, err := ns.subImpl(uid)
	if err != nil {
		return nil, err
	}
	ns.group[nc.uuid] = nc

	return nc, nil
}

func (nmb *nsqMailbox) ClusterSub(topic string) mailbox.ISubscriber {

	s := &nsqSubscriber{
		group:        make(map[string]mailbox.IConsumer),
		Channel:      nmb.parm.Channel,
		Topic:        topic,
		lookupAddres: nmb.parm.LookupAddres,
		addres:       nmb.parm.Addres,
		serviceName:  nmb.parm.ServiceName,
	}

	nmb.csubsrcibers = append(nmb.csubsrcibers, s)

	return s
}
