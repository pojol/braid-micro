// 接口文件 linkcache 链路缓存，主要用于维护携带 token 相关的链路调用信息
//
// 1. 用于固定链路的调用目标（这样可以辅助用户在本地执行一些优化操作
//
// 2. 广播服务节点的连接信息，用于web展示，以及一些负载均衡算法
package linkcache

import (
	"encoding/json"

	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/pubsub"
	"github.com/pojol/braid-go/service"
)

const (
	// 当前节点连接数事件
	TopicLinkNum = "linkcache.cluster.serviceLinkNum"

	// token 离线事件
	TopicUnlink = "linkcache.cluster.tokenUnlink"
)

// LinkNumMsg msg struct
type LinkNumMsg struct {
	ID  string
	Num int
}

// EncodeLinkNumMsg encode linknum msg
func EncodeNumMsg(id string, num int) *pubsub.Message {
	byt, _ := json.Marshal(&LinkNumMsg{
		ID:  id,
		Num: num,
	})

	return &pubsub.Message{
		Body: byt,
	}
}

// DecodeLinkNumMsg decode linknum msg
func DecodeNumMsg(msg *pubsub.Message) LinkNumMsg {
	lnmsg := LinkNumMsg{}
	json.Unmarshal(msg.Body, &lnmsg)
	return lnmsg
}

// ILinkCache 链路缓存，主要用于维护 token 和多个相关联的服务进程之间的关系
//
//  +---parent----------+
//  |                   |
//  |    +--child----+  |
//  |    |           |  |
//  |    | token ... |  |
//  |    |           |  |
//  |    +-----------+  |
//  +-------------------+
type ILinkCache interface {
	module.IModule

	// Target 通过服务名，获取 token 指向的目标服务器地址信息
	Target(token string, serviceName string) (targetAddr string, err error)

	// Link 将 token 和目标服务器连接信息写入到缓存中
	Link(token string, target service.Node) error

	// Unlink 将 token 和目标服务器连接信息，解除绑定关系
	Unlink(token string) error

	// Down 清理目标节点的连接信息（因为该服务已经退出
	Down(target service.Node) error
}
