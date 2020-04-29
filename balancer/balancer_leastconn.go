package balancer

import (
	"sort"
	"sync"

	"github.com/pojol/braid/log"
)

// LeastConnBalancer 最少连接数均衡器
type LeastConnBalancer struct {
	NodeName string
	lst      Nodes
	sync.Mutex
}

// Add 新增一个节点
func (b *LeastConnBalancer) Add(node Node) {
	b.Lock()
	defer b.Unlock()

	for _, v := range b.lst {
		if v.ID == node.ID {
			return
		}
	}

	log.Debugf("add balance nod", node.Name, node.ID, node.Address)
	b.lst = append(b.lst, node)
}

// Rmv 移除一个节点
func (b *LeastConnBalancer) Rmv(id string) {

	b.Lock()
	defer b.Unlock()

	for k, v := range b.lst {
		if v.ID == id {
			b.lst = append(b.lst[:k], b.lst[k+1:]...)

			log.Debugf("rmv balance nod", v.Name, v.ID, v.Address)
			return
		}
	}

}

// Next 最小连接数轮询
func (b *LeastConnBalancer) Next() (*Node, error) {

	b.Lock()
	defer b.Unlock()

	if len(b.lst) == 0 {
		return nil, ErrBalanceEmpty
	}

	if !sort.IsSorted(b.lst) {
		sort.Sort(b.lst)
	}

	b.lst[0].Tick++

	return &b.lst[0], nil
}
