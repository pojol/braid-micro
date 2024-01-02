// 实现文件 brandom 随机负载均衡算法实现
package balancer

import (
	"errors"
	"math/rand"
	"time"

	"github.com/pojol/braid-go/module/meta"
)

type randomBalancer struct {
	nods []meta.Node
}

func (rb *randomBalancer) exist(id string) (int, bool) {
	for k, v := range rb.nods {
		if v.ID == id {
			return k, true
		}
	}

	return -1, false
}

func (rb *randomBalancer) Add(nod meta.Node) {

	if _, ok := rb.exist(nod.ID); ok {
		return
	}

	rb.nods = append(rb.nods, nod)
}

func (rb *randomBalancer) Rmv(nod meta.Node) {

	idx, ok := rb.exist(nod.ID)
	if !ok {
		return
	}

	rb.nods = append(rb.nods[:idx], rb.nods[idx+1:]...)
}

func (rb *randomBalancer) Update(nod meta.Node) {
	//
}

func (rb *randomBalancer) Get() (meta.Node, error) {

	if len(rb.nods) <= 0 {
		return meta.Node{}, errors.New("empty")
	}

	rand.Seed(time.Now().UnixNano())
	return rb.nods[rand.Intn(len(rb.nods))], nil
}
