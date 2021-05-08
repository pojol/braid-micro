package mailboxnsq

import (
	"math/rand"
	"sync"

	"github.com/nsqio/go-nsq"
	"github.com/pojol/braid-go/module/mailbox"
)

type IConsumer interface {
	Arrived() <-chan *mailbox.Message
	Put(msg *mailbox.Message)
	Exit()
}

type mailboxChannel struct {
	sync.RWMutex
	clusterConsumer IConsumer

	producer []*nsq.Producer
	msgCh    chan *mailbox.Message

	mailbox *nsqMailbox

	Name      string
	TopicName string
	scope     mailbox.ScopeTy
}

func newChannel(topicName, channelName string, scope mailbox.ScopeTy, n *nsqMailbox) *mailboxChannel {

	c := &mailboxChannel{
		Name:      channelName,
		TopicName: topicName,
		scope:     scope,
		mailbox:   n,
		msgCh:     make(chan *mailbox.Message, 10000),
	}

	if scope == mailbox.ScopeCluster {
		cps := make([]*nsq.Producer, 0, len(n.parm.Address))
		var err error
		var cp *nsq.Producer

		for _, addr := range n.parm.Address {
			cp, err = nsq.NewProducer(addr, nsq.NewConfig())
			if err != nil {
				n.log.Errorf("Channel new nsq producer err %v", err.Error())
				continue
			}

			if err = cp.Ping(); err != nil {
				n.log.Errorf("Channel nsq producer ping err %v", err.Error())
				continue
			}

			cps = append(cps, cp)
			c.producer = cps
		}

		c.clusterConsumer, err = newClusterConsumer(topicName, channelName, n)
		if err != nil {
			n.log.Errorf("Channel nsq create consumer err %v", err.Error())
		}
	}

	return c
}

func (c *mailboxChannel) Put(msg *mailbox.Message) {
	if c.scope == mailbox.ScopeProc {
		select {
		case c.msgCh <- msg:
			//default: 就阻塞在这里吧。。
			//	c.mailbox.log.Errorf("put msg err channel %v is full", c.Name)
		}
	} else if c.scope == mailbox.ScopeCluster {
		c.producer[rand.Intn(len(c.producer))].Publish(c.TopicName, msg.Body)
	}
}

func (c *mailboxChannel) Arrived() <-chan *mailbox.Message {
	if c.scope == mailbox.ScopeProc {
		return c.msgCh
	} else if c.scope == mailbox.ScopeCluster {
		return c.clusterConsumer.Arrived()
	}

	return nil
}

func (c *mailboxChannel) Exit() error {

	return nil
}
