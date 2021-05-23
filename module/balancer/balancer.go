package balancer

import (
	"strings"

	"github.com/pojol/braid-go/module/discover"
	"github.com/pojol/braid-go/module/logger"
)

// Builder balancer builder
type Builder interface {
	Build(logger logger.ILogger) (IBalancer, error)
	Name() string
}

// IBalancer 负载均衡
type IBalancer interface {
	// Pick 从当前的负载均衡算法中，选取一个匹配的节点
	Pick() (nod discover.Node, err error)

	// Add 为当前的服务添加一个新的节点 service gate : [ gate1, gate2 ]
	Add(discover.Node)

	// Rmv 从当前的服务中移除一个旧的节点
	Rmv(discover.Node)

	// Update 更新一个当前服务中的节点（通常是权重信息
	Update(discover.Node)
}

var (
	m = make(map[string]Builder)
)

// Register 注册balancer
func Register(b Builder) {
	m[strings.ToLower(b.Name())] = b
}

// GetBuilder 获取构建器
func GetBuilder(name string) Builder {
	if b, ok := m[strings.ToLower(name)]; ok {
		return b
	}
	return nil
}
