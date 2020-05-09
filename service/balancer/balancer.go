package balancer

import (
	"errors"
	"sync"
)

// SelectorCfg 负载均衡选择器配置
type SelectorCfg struct {
}

// Selector 负载均衡选择器
type Selector struct {
	m sync.Map
}

var (
	// ErrBalanceEmpty 没有权重节点
	ErrBalanceEmpty = errors.New("weighted node list is empty")
)

// Node 权重节点
type Node struct {
	ID      string
	Name    string
	Address string

	// 节点的权重值
	Weight int
}

// IBalancer 均衡器
type IBalancer interface {
	// Add 添加一个新节点
	Add(Node)

	// Rmv 移除一个旧节点
	Rmv(string)

	// SyncWeight 调整节点权重值
	SyncWeight(string, int)

	// Next 获取一个节点
	Next() (*Node, error)
}

var (
	selector *Selector

	defaultSelectorCfg = SelectorCfg{}
)

// New 初始化负载均衡选择器
func New() *Selector {
	selector = &Selector{}
	return selector
}

// Init 初始化均衡器
func (s *Selector) Init(cfg interface{}) error {
	return nil
}

// Run r
func (s *Selector) Run() {

}

// Close c
func (s *Selector) Close() {

}

// GetSelector 获取负载均衡选择器
func GetSelector(nodName string) IBalancer {
	return selector.group(nodName)
}

// Group 获取组
func (s *Selector) group(nodName string) IBalancer {

	b, ok := s.m.Load(nodName)
	if !ok {
		b = &WeightedRoundrobin{
			Name: nodName,
		}

		s.m.Store(nodName, b)
	}

	return b.(IBalancer)
}
