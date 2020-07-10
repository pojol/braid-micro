package swrrbalancer

import (
	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/plugin/balancer"
)

const (
	// BalancerName 平滑加权负载均衡
	BalancerName = "SmoothWeightedRoundrobin"
)

type smoothWeightRoundrobinBuilder struct{}

func newSmoothWightRoundrobinBalancer() balancer.Builder {
	return &smoothWeightRoundrobinBuilder{}
}

func (*smoothWeightRoundrobinBuilder) Build() balancer.Balancer {
	return &swrrBalancer{}
}

func (*smoothWeightRoundrobinBuilder) Name() string {
	return BalancerName
}

type weightedNod struct {
	orgNod    balancer.Node
	curWeight int
}

// swrrBalancer 平滑加权轮询
type swrrBalancer struct {
	totalWeight int
	nods        []weightedNod
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

// Update 更新负载均衡节点
func (wr *swrrBalancer) Update(nod balancer.Node) {

	if nod.OpTag == balancer.OpAdd {
		wr.add(nod)
	} else if nod.OpTag == balancer.OpRmv {
		wr.rmv(nod)
	} else if nod.OpTag == balancer.OpUp {
		wr.syncWeight(nod)
	}

}

// Pick 执行算法，选取节点
func (wr *swrrBalancer) Pick() (balancer.Node, error) {
	var tmpWeight int
	var idx int

	if len(wr.nods) <= 0 {
		return balancer.Node{}, balancer.ErrBalanceEmpty
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

func (wr *swrrBalancer) add(nod balancer.Node) {

	if _, ok := wr.isExist(nod.ID); ok {
		// log
		return
	}

	wr.nods = append(wr.nods, weightedNod{
		orgNod:    nod,
		curWeight: int(nod.Weight),
	})

	wr.calcTotalWeight()
	log.Debugf("add weighted nod id : %s space : %s", nod.ID, nod.Name)
}

func (wr *swrrBalancer) rmv(nod balancer.Node) {

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

func (wr *swrrBalancer) syncWeight(nod balancer.Node) {

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
