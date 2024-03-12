package linkcacheredis

// mode
const (
	LinkerRedisModeLocal = "mode_local"
	LinkerRedisModeRedis = "mode_redis"
)

/*
	braid.linkcache.

	// parent service . child service . token : linkinfo
	// target - 用于查询 token 是否已经有链路的信息
	// link - 建立链路信息
	// unlink - 删除链路信息
	linkcache.gate-1.base-2.xxxyyydd : service_addr:port	// 记录token指向的service地址

	// parent service . child service : { addr, name, id }
	gate-1.base-2 : { addr, name, id }	// 记录父子节点之间的关系

	// parent service . child service : cnt
	linkcache.gate-1.base-2
*/

// Parm Service 配置
type Parm struct {
	Mode     string
	SyncTick int // ms

	//
	SyncRelationTick int // second

	//
	SyncOfflineTick int // second
}

// Option config wraps
type Option func(*Parm)

// WithSyncTick 同步周期
func WithSyncTick(mstime int) Option {
	return func(c *Parm) {
		c.SyncTick = mstime
	}
}

// WithMode 设置redis link-cache的执行模式
// 这边需要更多的注释
func WithMode(mode string) Option {
	return func(c *Parm) {
		c.Mode = mode
	}
}
