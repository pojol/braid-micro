package mailboxnsq

import (
	"math/rand"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/pojol/braid/internal/buffer"
	"github.com/pojol/braid/module/mailbox"
)

const (
	// Name mailbox plug-in name
	Name = "MailboxNsq"
)

type nsqMailboxBuilder struct {
	opts []interface{}
}

func newNsqMailbox() mailbox.Builder {
	return &nsqMailboxBuilder{}
}

func (nb *nsqMailboxBuilder) AddOption(opt interface{}) {
	nb.opts = append(nb.opts, opt)
}

func (nb *nsqMailboxBuilder) Name() string {
	return Name
}

func (nb *nsqMailboxBuilder) Build(serviceName string) (mailbox.IMailbox, error) {
	p := Parm{
		ServiceName: serviceName,
	}
	for _, opt := range nb.opts {
		opt.(Option)(&p)
	}

	rand.Seed(time.Now().UnixNano())

	cps := make([]*nsq.Producer, 0, len(p.Address))
	for _, addr := range p.Address {
		cp, err := nsq.NewProducer(addr, nsq.NewConfig())
		if err != nil {
			return nil, err
		}

		if err = cp.Ping(); err != nil {
			return nil, err
		}

		cps = append(cps, cp)
	}

	nsqm := &nsqMailbox{
		parm: p,
		proc: &procMailbox{
			subscribers: make(map[string]*procSubscriber),
			recvBuff:    buffer.NewUnbounded(),
			exitChan:    make(chan int),
		},
		cproducers: cps,
	}
	go nsqm.proc.router()
	return nsqm, nil
}

type nsqMailbox struct {
	parm Parm

	proc *procMailbox

	cproducers   []*nsq.Producer
	csubsrcibers []*nsqSubscriber
}

func (nmb *nsqMailbox) Pub(scope string, topic string, msg *mailbox.Message) {

	if scope == mailbox.Proc {
		nmb.proc.pub(topic, msg)
	} else if scope == mailbox.Cluster {
		nmb.pub(topic, msg)
	}

}

func (nmb *nsqMailbox) Sub(scope string, topic string) mailbox.ISubscriber {
	if scope == mailbox.Proc {
		return nmb.proc.sub(topic)
	} else if scope == mailbox.Cluster {
		return nmb.sub(topic)
	}

	return nil
}

func init() {
	mailbox.Register(newNsqMailbox())
}
