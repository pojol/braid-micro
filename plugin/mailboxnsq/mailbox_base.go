package mailboxnsq

import (
	"math/rand"
	"sync"
	"time"

	"github.com/nsqio/go-nsq"
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
		ServiceName:  serviceName,
		LookupAddres: []string{"127.0.0.1:4161"},
		Addres:       []string{"127.0.0.1:4150"},
	}
	for _, opt := range nb.opts {
		opt.(Option)(&p)
	}

	rand.Seed(time.Now().UnixNano())

	cps := make([]*nsq.Producer, 0, len(p.Addres))
	for _, addr := range p.Addres {
		cp, err := nsq.NewProducer(addr, nsq.NewConfig())
		if err != nil {
			return nil, err
		}
		cps = append(cps, cp)
	}

	nsqm := &nsqMailbox{
		parm:       p,
		cproducers: cps,
	}
	return nsqm, nil
}

type nsqMailbox struct {
	parm Parm

	psubsrcibers sync.Map

	cproducers   []*nsq.Producer
	csubsrcibers []*nsqSubscriber
}

func init() {
	mailbox.Register(newNsqMailbox())
}
