package meta

const (
	ModuleDiscover = "barid.module.discover" // 服务发现
	ModuleElector  = "barid.module.elector"  // 服务选举
	ModuleBalancer = "barid.module.balancer" // 负载均衡
	ModuleLink     = "barid.module.link"     // 链路缓存
	ModuleMonitor  = "barid.module.monitor"  // 服务监控
	ModulePubsub   = "barid.module.pubsub"   // 发布订阅
	ModuleClient   = "barid.module.client"   // rpc客户端
	ModuleServer   = "barid.module.server"   // rpc服务端
	ModuleTracer   = "braid.module.tracer"   // 链路追踪
)
