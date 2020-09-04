package balancerswrr

import (
	"encoding/json"
	"sync"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/module/balancer"
	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/module/pubsub"
)

const (
	// BalancerName 平滑加权负载均衡
	BalancerName = "SmoothWeightedRoundrobin"
)

type smoothWeightRoundrobinBuilder struct{}

func newSmoothWightRoundrobinBalancer() balancer.Builder {
	return &smoothWeightRoundrobinBuilder{}
}

func (*smoothWeightRoundrobinBuilder) Build(pubsub pubsub.IPubsub, serviceName string) balancer.Balancer {

	swrr := &swrrBalancer{
		addSub: pubsub.Sub(discover.EventAdd + "_" + serviceName).AddCompetition(),
		rmvSub: pubsub.Sub(discover.EventRmv + "_" + serviceName).AddCompetition(),
		upSub:  pubsub.Sub(discover.EventUpdate + "_" + serviceName).AddCompetition(),
	}

	go swrr.watcher()

	return swrr
}

func (wr *swrrBalancer) watcher() {

	wr.addSub.OnArrived(func(msg *pubsub.Message) error {
		nod := discover.Node{}
		json.Unmarshal(msg.Body, &nod)

		wr.add(nod)
		return nil
	})

	wr.rmvSub.OnArrived(func(msg *pubsub.Message) error {
		nod := discover.Node{}
		json.Unmarshal(msg.Body, &nod)
		wr.rmv(nod)
		return nil
	})

	wr.upSub.OnArrived(func(msg *pubsub.Message) error {
		nod := discover.Node{}
		json.Unmarshal(msg.Body, &nod)
		wr.syncWeight(nod)
		return nil
	})
}

func (*smoothWeightRoundrobinBuilder) Name() string {
	return BalancerName
}

type weightedNod struct {
	orgNod    discover.Node
	curWeight int
}

// swrrBalancer 平滑加权轮询
type swrrBalancer struct {
	addSub pubsub.IConsumer
	rmvSub pubsub.IConsumer
	upSub  pubsub.IConsumer

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
		return discover.Node{}, balancer.ErrBalanceEmpty
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
	balancer.Register(newSmoothWightRoundrobinBalancer())
}
