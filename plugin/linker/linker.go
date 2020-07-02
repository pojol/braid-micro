package linker

import (
	"context"
	"strings"
)

// Builder 构建器接口
type Builder interface {
	Build(cfg interface{}) ILinker
	Name() string
}

// ILinker 链接器 (保存nod中的链接信息
// nodid 节点id
// token 用户链接在节点上的身份id
// target 节点的真实地址
type ILinker interface {
	// 提供正向查找功能，通过token检索到token原本指向的nod address
	Target(ctx context.Context, token string) (target string, err error)

	// 将token绑定到nod
	Link(ctx context.Context, token string, nodid string, target string) error
	// 解除token在nod的绑定
	Unlink(token string) error

	// 提供nod中token的数量
	Num(ctx context.Context, nodid string) (int, error)

	// 清空nod中的token
	Offline(ctx context.Context, nodid string) error
}

var (
	m = make(map[string]Builder)
)

// Register 注册linker
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
