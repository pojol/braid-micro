package mailboxnsq

import (
	"github.com/nsqio/go-nsq"
	"github.com/pojol/braid-go/module/mailbox"
)

type consumerHandler struct {
	channel string
	c       *clusterConsumer
}

func (ch *consumerHandler) HandleMessage(msg *nsq.Message) error {
	ch.c.mailbox.log.Debug("recv msg")
	ch.c.Put(&mailbox.Message{
		Body: msg.Body,
	})
	return nil
}

type clusterConsumer struct {
	ID          string
	topicName   string
	channelName string
	nsqConsumer *nsq.Consumer
	mailbox     *nsqMailbox
	msgCh       chan *mailbox.Message
}

func newClusterConsumer(topicName, channelName string, n *nsqMailbox) (IConsumer, error) {

	consumer := &clusterConsumer{
		topicName:   topicName,
		channelName: channelName,
		mailbox:     n,
		msgCh:       make(chan *mailbox.Message, 10000),
	}

	nsqConsumer, err := nsq.NewConsumer(topicName, channelName, nsq.NewConfig())
	if err != nil {
		return nil, err
	}
	nsqConsumer.SetLoggerLevel(n.parm.nsqLogLv)

	nsqConsumer.AddHandler(&consumerHandler{
		c:       consumer,
		channel: channelName,
	})

	err = nsqConsumer.ConnectToNSQLookupds(n.parm.LookupAddress)
	if err != nil {
		return nil, err
	}

	n.log.Infof("Cluster consumer %v created", channelName)
	consumer.nsqConsumer = nsqConsumer

	return consumer, nil
}

func (cc *clusterConsumer) Arrived() <-chan *mailbox.Message {
	return cc.msgCh
}

func (cc *clusterConsumer) Put(msg *mailbox.Message) {
	select {
	case cc.msgCh <- msg:
		//default:
		//	cc.mailbox.log.Errorf("put msg err channel %v is full", cc.channelName)
	}
}

func (cc *clusterConsumer) Exit() {

}
