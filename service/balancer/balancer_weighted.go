package balancer

import (
	"sync"
)

type weightedNod struct {
	orgNod    Node
	curWeight int
}

// WeightedRoundrobin 平滑加权轮询
type WeightedRoundrobin struct {
	Name string

	totalWeight int
	Nods        []weightedNod

	sync.Mutex
}

func (wr *WeightedRoundrobin) calcTotalWeight() {
	wr.totalWeight = 0

	for _, v := range wr.Nods {
		wr.totalWeight += v.orgNod.Weight
	}
}

func (wr *WeightedRoundrobin) isExist(id string) (int, bool) {
	for k, v := range wr.Nods {
		if v.orgNod.ID == id {
			return k, true
		}
	}

	return -1, false
}

// Add 新增一个节点
func (wr *WeightedRoundrobin) Add(nod Node) {

	wr.Lock()
	defer wr.Unlock()

	if _, ok := wr.isExist(nod.ID); ok {
		// log
		return
	}

	wr.Nods = append(wr.Nods, weightedNod{
		orgNod:    nod,
		curWeight: int(nod.Weight),
	})

	wr.calcTotalWeight()
}

// Rmv 移除一个节点
func (wr *WeightedRoundrobin) Rmv(id string) {

	wr.Lock()
	defer wr.Unlock()

	var ok bool
	var idx int

	idx, ok = wr.isExist(id)
	if !ok {
		// log
		return
	}

	wr.Nods = append(wr.Nods[:idx], wr.Nods[idx+1:]...)
	wr.calcTotalWeight()
}

// SyncWeight 同步节点权重值
func (wr *WeightedRoundrobin) SyncWeight(id string, weight int) {
	wr.Lock()
	defer wr.Unlock()

	var ok bool
	var idx int

	idx, ok = wr.isExist(id)
	if ok {
		wr.Nods[idx].orgNod.Weight = weight
		wr.calcTotalWeight()
	}
}

// Next 获取下一个节点
func (wr *WeightedRoundrobin) Next() (*Node, error) {

	var tmpWeight int
	var idx int

	if len(wr.Nods) <= 0 {
		return nil, ErrBalanceEmpty
	}

	for k, v := range wr.Nods {
		if tmpWeight < v.curWeight+wr.totalWeight {
			tmpWeight = v.curWeight + wr.totalWeight
			idx = k
		}
	}

	for k := range wr.Nods {
		if k == idx {
			wr.Nods[idx].curWeight = wr.Nods[idx].curWeight - wr.totalWeight + wr.Nods[idx].orgNod.Weight
		} else {
			wr.Nods[k].curWeight += wr.Nods[k].orgNod.Weight
		}
	}

	return &wr.Nods[idx].orgNod, nil
}
