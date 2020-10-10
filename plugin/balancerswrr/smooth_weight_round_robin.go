package balancerswrr

import (
	"encoding/json"
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/module"
	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/module/mailbox"
)

const (
	// Name 平滑加权负载均衡
	Name = "SmoothWeightedRoundrobin"
)

type smoothWeightRoundrobinBuilder struct {
	opts []interface{}
}

func newSmoothWightRoundrobinBalancer() module.Builder {
	return &smoothWeightRoundrobinBuilder{}
}

func (b *smoothWeightRoundrobinBuilder) AddOption(opt interface{}) {
	b.opts = append(b.opts, opt)
}

func (b *smoothWeightRoundrobinBuilder) Build(serviceName string, mb mailbox.IMailbox) (module.IModule, error) {

	swrr := &swrrBalancer{
		serviceName: serviceName,
		mb:          mb,
	}

	swrr.addSub, _ = mb.ProcSub(discover.AddService).AddShared()
	swrr.rmvSub, _ = mb.ProcSub(discover.RmvService).AddShared()
	swrr.upSub, _ = mb.ProcSub(discover.UpdateService).AddShared()

	go swrr.watcher()

	return swrr, nil
}

func (wr *swrrBalancer) Init() {

}

func (wr *swrrBalancer) Run() {

}

func (wr *swrrBalancer) Close() {

}

func (wr *swrrBalancer) watcher() {

	wr.addSub.OnArrived(func(msg *mailbox.Message) error {
		nod := discover.Node{}
		json.Unmarshal(msg.Body, &nod)

		if nod.Name == wr.serviceName {
			wr.add(nod)
		}

		return nil
	})

	wr.rmvSub.OnArrived(func(msg *mailbox.Message) error {
		nod := discover.Node{}
		json.Unmarshal(msg.Body, &nod)

		if nod.Name == wr.serviceName {
			wr.rmv(nod)
		}

		return nil
	})

	wr.upSub.OnArrived(func(msg *mailbox.Message) error {
		nod := discover.Node{}
		json.Unmarshal(msg.Body, &nod)

		if nod.Name == wr.serviceName {
			wr.syncWeight(nod)
		}

		return nil
	})
}

func (*smoothWeightRoundrobinBuilder) Name() string {
	return Name
}

func (*smoothWeightRoundrobinBuilder) Type() string {
	return module.TyBalancer
}

type weightedNod struct {
	orgNod    discover.Node
	curWeight int
}

// swrrBalancer 平滑加权轮询
type swrrBalancer struct {
	addSub mailbox.IConsumer
	rmvSub mailbox.IConsumer
	upSub  mailbox.IConsumer

	serviceName string
	mb          mailbox.IMailbox

	totalWeight int
	nods        []weightedNod
	sync.Mutex
}

func (wr *swrrBalancer) calcTotalWeight() {
	wr.totalWeight = 0

	for _, v := range wr.nods {
		wr.totalWeight += v.orgNod.Weight
	}
}

func (wr *swrrBalancer) isExist(id string) (int, bool) {
	for k, v := range wr.nods {
		if v.orgNod.ID == id {
			return k, true
		}
	}

	return -1, false
}

// Pick 执行算法，选取节点
func (wr *swrrBalancer) Pick() (discover.Node, error) {
	var tmpWeight int
	var idx int
	wr.Lock()
	defer wr.Unlock()

	if len(wr.nods) <= 0 {
		return discover.Node{}, errors.New("empty")
	}

	for k, v := range wr.nods {
		if tmpWeight < v.curWeight+wr.totalWeight {
			tmpWeight = v.curWeight + wr.totalWeight
			idx = k
		}
	}

	for k := range wr.nods {
		if k == idx {
			wr.nods[idx].curWeight = wr.nods[idx].curWeight - wr.totalWeight + wr.nods[idx].orgNod.Weight
		} else {
			wr.nods[k].curWeight += wr.nods[k].orgNod.Weight
		}
	}

	return wr.nods[idx].orgNod, nil
}

// Randowm ..
func (wr *swrrBalancer) Random() (discover.Node, error) {
	wr.Lock()
	defer wr.Unlock()

	if len(wr.nods) <= 0 {
		return discover.Node{}, errors.New("empty")
	}

	rand.Seed(time.Now().UnixNano())
	return wr.nods[rand.Intn(len(wr.nods))].orgNod, nil
}

func (wr *swrrBalancer) add(nod discover.Node) {

	if _, ok := wr.isExist(nod.ID); ok {
		// log
		return
	}

	wr.nods = append(wr.nods, weightedNod{
		orgNod:    nod,
		curWeight: int(nod.Weight),
	})

	wr.calcTotalWeight()

	log.Debugf("add weighted nod id : %s space : %s weight : %d", nod.ID, nod.Name, nod.Weight)
}

func (wr *swrrBalancer) rmv(nod discover.Node) {

	var ok bool
	var idx int

	idx, ok = wr.isExist(nod.ID)
	if !ok {
		// log
		return
	}

	wr.nods = append(wr.nods[:idx], wr.nods[idx+1:]...)

	wr.calcTotalWeight()
	log.Debugf("rmv weighted nod id : %s space : %s", nod.ID, nod.Name)
}

func (wr *swrrBalancer) syncWeight(nod discover.Node) {

	var ok bool
	var idx int

	idx, ok = wr.isExist(nod.ID)
	if ok {
		wr.nods[idx].orgNod.Weight = nod.Weight
		wr.calcTotalWeight()
	}

	log.Debugf("update weighted nod id : %s space : %s weight : %d", nod.ID, nod.Name, nod.Weight)
}

func init() {
	module.Register(newSmoothWightRoundrobinBalancer())
}
