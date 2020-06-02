package balancer

import (
	"github.com/pojol/braid/log"
)

const (
	balancerName = "WeightedRoundrobin"
)

type weightRoundrobinBuilder struct{}

func newWightRoundrobinBalancer() Builder {
	return &weightRoundrobinBuilder{}
}

func (*weightRoundrobinBuilder) Build() Balancer {
	return &wightedRoundrobin{}
}

func (*weightRoundrobinBuilder) Name() string {
	return balancerName
}

type weightedNod struct {
	orgNod    Node
	curWeight int
}

// WeightedRoundrobin 平滑加权轮询
type wightedRoundrobin struct {
	totalWeight int
	nods        []weightedNod
}

func (wr *wightedRoundrobin) calcTotalWeight() {
	wr.totalWeight = 0

	for _, v := range wr.nods {
		wr.totalWeight += v.orgNod.Weight
	}
}

func (wr *wightedRoundrobin) isExist(id string) (int, bool) {
	for k, v := range wr.nods {
		if v.orgNod.ID == id {
			return k, true
		}
	}

	return -1, false
}

// Update 更新负载均衡节点
func (wr *wightedRoundrobin) Update(nod Node) {

	if nod.OpTag == OpAdd {
		wr.add(nod)
	} else if nod.OpTag == OpRmv {
		wr.rmv(nod)
	} else if nod.OpTag == OpUp {
		wr.syncWeight(nod)
	}

}

// Pick 执行算法，选取节点
func (wr *wightedRoundrobin) Pick() (string, error) {
	var tmpWeight int
	var idx int

	if len(wr.nods) <= 0 {
		return "", ErrBalanceEmpty
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

	return wr.nods[idx].orgNod.Address, nil
}

func (wr *wightedRoundrobin) add(nod Node) {

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

func (wr *wightedRoundrobin) rmv(nod Node) {

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

func (wr *wightedRoundrobin) syncWeight(nod Node) {

	var ok bool
	var idx int

	idx, ok = wr.isExist(nod.ID)
	if ok {
		wr.nods[idx].orgNod.Weight = nod.Weight
		wr.calcTotalWeight()
	}
}

func init() {
	Register(newWightRoundrobinBalancer())
}
