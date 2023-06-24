// 实现文件 balancerswrr 平滑加权负载均衡算法实现
package balancer

import (
	"errors"
	"sync"

	"github.com/pojol/braid-go/module/meta"
)

type weightedNod struct {
	orgNod    meta.Node
	curWeight int
}

// swrrBalancer 平滑加权轮询
type swrrBalancer struct {
	totalWeight int
	nods        []weightedNod
	sync.Mutex
}

func (wr *swrrBalancer) calcTotalWeight() {
	wr.totalWeight = 0

	for _, v := range wr.nods {
		wr.totalWeight += v.orgNod.GetWidget()
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
func (wr *swrrBalancer) Get() (meta.Node, error) {
	var tmpWeight int
	var idx int
	wr.Lock()
	defer wr.Unlock()

	if len(wr.nods) <= 0 {
		return meta.Node{}, errors.New("empty")
	}

	for k, v := range wr.nods {
		if tmpWeight < v.curWeight+wr.totalWeight {
			tmpWeight = v.curWeight + wr.totalWeight
			idx = k
		}
	}

	for k := range wr.nods {
		if k == idx {
			wr.nods[idx].curWeight = wr.nods[idx].curWeight - wr.totalWeight + wr.nods[idx].orgNod.GetWidget()
		} else {
			wr.nods[k].curWeight += wr.nods[k].orgNod.GetWidget()
		}
	}

	return wr.nods[idx].orgNod, nil
}

func (wr *swrrBalancer) Add(nod meta.Node) {

	if _, ok := wr.isExist(nod.ID); ok {
		return
	}

	wr.nods = append(wr.nods, weightedNod{
		orgNod:    nod,
		curWeight: int(nod.GetWidget()),
	})

	wr.calcTotalWeight()

	//fmt.Printf("add weighted nod id : %s name : %s weight : %d\n", nod.ID, nod.Name, nod.Weight)
}

func (wr *swrrBalancer) Rmv(nod meta.Node) {

	var ok bool
	var idx int

	idx, ok = wr.isExist(nod.ID)
	if !ok {
		// log
		return
	}

	wr.nods = append(wr.nods[:idx], wr.nods[idx+1:]...)

	wr.calcTotalWeight()
	//fmt.Println("rmv weighted nod id : %s name : %s\n", nod.ID, nod.Name)
}

func (wr *swrrBalancer) Update(nod meta.Node) {

	var ok bool
	var idx int

	idx, ok = wr.isExist(nod.ID)
	if ok {
		wr.nods[idx].orgNod.SetWidget(nod.GetWidget())
		wr.calcTotalWeight()
	}

	//fmt.Println("update weighted nod id : %s name : %s weight : %d\n", nod.ID, nod.Name, nod.Weight)
}
