package discover

import "github.com/pojol/braid-go/module"

const (
	// EventAddService 有一个新的服务加入到集群
	EventAddService = "event.service.nodeAdd"

	// EventRemoveService 有一个旧的服务从集群中退出
	EventRemoveService = "event.service.nodRmv"

	// EventUpdateService 有一个旧的服务产生了信息的变更（通常是指权重
	EventUpdateService = "event.service.nodUpdate"
)

// IDiscover discover interface
type IDiscover interface {
	module.IModule
}
