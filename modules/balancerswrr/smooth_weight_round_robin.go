package balancerswrr

import (
	"errors"
	"sync"

	"github.com/pojol/braid/module/balancer"
	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/module/logger"
	"github.com/pojol/braid/module/mailbox"
)

const (
	// Name 平滑加权负载均衡
	Name = "SmoothWeightedRoundrobin"
)

type smoothWeightRoundrobinBuilder struct {
}

func newSmoothWightRoundrobinBalancer() balancer.Builder {
	return &smoothWeightRoundrobinBuilder{}
}

func (b *smoothWeightRoundrobinBuilder) Build(logger logger.ILogger) (balancer.IBalancer, error) {

	swrr := &swrrBalancer{
		logger: logger,
	}

	return swrr, nil
}

func (*smoothWeightRoundrobinBuilder) Name() string {
	return Name
}

type weightedNod struct {
	orgNod    discover.Node
	curWeight int
}

// swrrBalancer 平滑加权轮询
type swrrBalancer struct {
	serviceName string
	mb          mailbox.IMailbox
	logger      logger.ILogger

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

func (wr *swrrBalancer) Add(nod discover.Node) {

	if _, ok := wr.isExist(nod.ID); ok {
		return
	}

	wr.nods = append(wr.nods, weightedNod{
		orgNod:    nod,
		curWeight: int(nod.Weight),
	})

	wr.calcTotalWeight()

	wr.logger.Debugf("add weighted nod id : %s space : %s weight : %d", nod.ID, nod.Name, nod.Weight)
}

func (wr *swrrBalancer) Rmv(nod discover.Node) {

	var ok bool
	var idx int

	idx, ok = wr.isExist(nod.ID)
	if !ok {
		// log
		return
	}

	wr.nods = append(wr.nods[:idx], wr.nods[idx+1:]...)

	wr.calcTotalWeight()
	wr.logger.Debugf("rmv weighted nod id : %s space : %s", nod.ID, nod.Name)
}

func (wr *swrrBalancer) Update(nod discover.Node) {

	var ok bool
	var idx int

	idx, ok = wr.isExist(nod.ID)
	if ok {
		wr.nods[idx].orgNod.Weight = nod.Weight
		wr.calcTotalWeight()
	}

	wr.logger.Debugf("update weighted nod id : %s space : %s weight : %d", nod.ID, nod.Name, nod.Weight)
}

func init() {
	balancer.Register(newSmoothWightRoundrobinBalancer())
}
