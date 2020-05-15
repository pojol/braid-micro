package balancer

import (
	"errors"
	"sync"

	"github.com/pojol/braid/log"
)

// Cfg 负载均衡选择器配置
type Cfg struct {
}

// Balancer 负载均衡选择器
type Balancer struct {
	m sync.Map
}

var (
	// ErrBalanceEmpty 没有权重节点
	ErrBalanceEmpty = errors.New("weighted node list is empty")
	// ErrUninitialized 未初始化
	ErrUninitialized = errors.New("balancer uninitalized")
)

// Node 权重节点
type Node struct {
	ID      string
	Name    string
	Address string

	// 节点的权重值
	Weight int
}

// IGroup 组均衡器
type IGroup interface {
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
	b *Balancer

	defaultSelectorCfg = Cfg{}
)

// New 初始化负载均衡选择器
func New() *Balancer {
	b = &Balancer{}
	return b
}

// GetGroup 获取负载均衡选择器
func GetGroup(nodName string) (IGroup, error) {

	if b == nil {
		return nil, ErrUninitialized
	}

	return b.group(nodName), nil
}

// Group 获取组
func (b *Balancer) group(nodName string) IGroup {

	wb, ok := b.m.Load(nodName)
	if !ok {
		wb = &WeightedRoundrobin{
			Name: nodName,
		}

		b.m.Store(nodName, wb)
		log.Debugf("add balance group %v\n", nodName)
	}

	return wb.(IGroup)
}
