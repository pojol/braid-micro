package balancerrandom

import (
	"errors"
	"math/rand"
	"time"

	"github.com/pojol/braid/module"
	"github.com/pojol/braid/module/balancer"
	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/module/logger"
)

const (
	// Name 随机负载均衡器
	Name = "RandomBalance"
)

type randomBalanceBuilder struct {
}

func newRandomBalance() balancer.Builder {
	return &randomBalanceBuilder{}
}

func (*randomBalanceBuilder) Name() string {
	return Name
}

func (*randomBalanceBuilder) Type() string {
	return module.TyBalancer
}

func (*randomBalanceBuilder) Build(logger logger.ILogger) (balancer.IBalancer, error) {

	rb := &randomBalance{
		logger: logger,
	}

	return rb, nil

}

type randomBalance struct {
	logger logger.ILogger
	nods   []discover.Node
}

func (rb *randomBalance) exist(id string) (int, bool) {
	for k, v := range rb.nods {
		if v.ID == id {
			return k, true
		}
	}

	return -1, false
}

func (rb *randomBalance) Add(nod discover.Node) {

	if _, ok := rb.exist(nod.ID); ok {
		return
	}

	rb.nods = append(rb.nods, nod)
}

func (rb *randomBalance) Rmv(nod discover.Node) {

	idx, ok := rb.exist(nod.ID)
	if !ok {
		return
	}

	rb.nods = append(rb.nods[:idx], rb.nods[idx+1:]...)
}

func (rb *randomBalance) Update(nod discover.Node) {
	//
}

func (rb *randomBalance) Pick() (discover.Node, error) {

	if len(rb.nods) <= 0 {
		return discover.Node{}, errors.New("empty")
	}

	rand.Seed(time.Now().UnixNano())
	return rb.nods[rand.Intn(len(rb.nods))], nil
}

func init() {
	balancer.Register(newRandomBalance())
}
