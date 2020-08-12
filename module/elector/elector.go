package elector

import (
	"strings"
)

// Builder 构建器接口
type Builder interface {
	Build() (IElection, error)
	Name() string
	SetCfg(cfg interface{}) error
}

// IElection 选举器需要提供的接口
type IElection interface {
	// 当前节点是否为Master节点
	// 适用于一些只能在单个进程内处理的业务
	IsMaster() bool

	// 侦听
	Run()
	// 关闭
	Close()
}

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
