package linker

import (
	"strings"
)

// Builder 构建器接口
type Builder interface {
	Build(cfg interface{}) ILinker
	Name() string
}

// ILinker 链接器 (保存链路信息
// nodid 节点id
// token 用户链接在节点上的身份id
// target 节点的真实地址
//
// 链接器是一个维护多个进程之间，用户链路关系的服务，
// 通常相关的操作指令都需要投送到相关的父master节点上，通过消费消息进行信息的添加删除操作。
type ILinker interface {
	// 提供正向查找功能，通过token检索到token原本指向的nod address
	Target(token string) (target string, err error)

	// 将token绑定到nod
	Link(token string, nodid string, target string) error
	// 解除token在nod的绑定
	Unlink(token string) error

	// 提供nod中token的数量
	Num(nodid string) (int, error)

	// 清空nod中的token
	Offline(nodid string) error
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
