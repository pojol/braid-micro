package braid

import (
	"context"
	"errors"
	"fmt"

	"github.com/pojol/braid-go/components"
	"github.com/pojol/braid-go/components/depends/blog"
	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/meta"
)

const (
	// Version of braid-go
	Version = "v1.4.0"

	banner = `
 _               _     _ 
| |             (_)   | |
| |__  _ __ __ _ _  __| |
| '_ \| '__/ _' | |/ _' |
| |_) | | | (_| | | (_| |
|_.__/|_|  \__,_|_|\__,_| %s

`
)

var (
	ErrTypeConvFailed = errors.New("type conversion failed")
)

// Braid framework instance
type Braid struct {
	info meta.ServiceInfo

	log *blog.Logger

	director components.IDirector
}

var (
	braidGlobal *Braid
)

// NewService - 创建一个新的 braid 服务
//
//	name 服务名称
//	id   服务id （唯一标识
//	director 服务组件构建器
func NewService(name string, id string, director components.IDirector) (*Braid, error) {

	director.SetServiceInfo(meta.ServiceInfo{ID: id, Name: name})
	director.Build()

	braidGlobal = &Braid{
		info:     meta.ServiceInfo{Name: name, ID: id},
		log:      director.Logger(),
		director: director,
	}

	return braidGlobal, nil
}

// Init braid init
func (b *Braid) Init() error {
	return b.director.Init()
}

// Run 运行braid
func (b *Braid) Run() {
	fmt.Printf(banner, Version)
	b.director.Run()
}

// Topic 获取或创建一个pubsub消息主题
func Topic(name string) module.ITopic {
	return braidGlobal.director.Pubsub().GetTopic(name)
}

// Send 发送rpc请求
//
//	target 目标服务名称
//	methon 目标服务方法
//	token  用户的唯一标识id
//	args   请求参数
//	reply  返回参数
//	opts   rpc调用选项
func Send(ctx context.Context, target, methon, token string,
	args, reply interface{},
	opts ...interface{}) error {
	return braidGlobal.director.Client().Invoke(ctx, target, methon, token, args, reply, opts...)
}

// Close 关闭braid
func (b *Braid) Close() {

	b.director.Close()

}
