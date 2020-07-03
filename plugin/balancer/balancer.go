package balancer

import (
	"errors"
	"strings"
	"sync"

	"github.com/pojol/braid/internal/braidsync"
)

var (
	m = make(map[string]Builder)
)

// Register 注册balancer
func Register(b Builder) {
	m[strings.ToLower(b.Name())] = b
}

// GetBuilder 获取balancer构建器
func GetBuilder(name string) Builder {
	if b, ok := m[strings.ToLower(name)]; ok {
		return b
	}
	return nil
}

// Builder 构建器接口
type Builder interface {
	Build() Balancer
	Name() string
}

// Balancer 均衡器接口
type Balancer interface {
	// 更新均衡器内容
	Update(nod Node)

	// 选取
	Pick() (nod Node, err error)
}

// Wrapper 负载均衡包装器
type Wrapper struct {
	b   Balancer
	bmu sync.Mutex
	q   *braidsync.Unbounded
	s   *braidsync.Switch
}

var (
	// ErrBalanceEmpty 没有权重节点
	ErrBalanceEmpty = errors.New("weighted node list is empty")
	// ErrUninitialized 未初始化
	ErrUninitialized = errors.New("balancer uninitialized")
)

const (
	// OpAdd 添加
	OpAdd = "add"

	// OpRmv 移除
	OpRmv = "rmv"

	// OpUp 更新
	OpUp = "update"
)

// Node 权重节点
type Node struct {
	ID string
	// 负载均衡节点的名称，这个名称主要用于均衡节点分组。
	Name    string
	Address string

	// 节点的权重值
	Weight int

	// 更新操作符
	OpTag string
}

// newBalancerWrapper 构建一个新的负载均衡包装器
func newBalancerWrapper(builder Builder) *Wrapper {
	w := &Wrapper{
		b: builder.Build(),
		q: braidsync.NewUnbounded(),
		s: braidsync.NewSwitch(),
	}

	go w.watcher()

	return w
}

func (w *Wrapper) watcher() {
	for {

		select {
		case nod := <-w.q.Get():
			w.q.Load()

			w.bmu.Lock()
			w.b.Update(nod.(Node))
			w.bmu.Unlock()
		case <-w.s.Done():
			return
		}

	}
}

// Update 将新的节点信息更新到balancer
func (w *Wrapper) Update(nod Node) {

	w.q.Put(nod)

}

// Pick 选取一个节点
func (w *Wrapper) Pick() (Node, error) {
	w.bmu.Lock()
	defer w.bmu.Unlock()

	return w.b.Pick()
}

// Close 关闭负载均衡器
func (w *Wrapper) Close() {
	w.s.Open()
}
