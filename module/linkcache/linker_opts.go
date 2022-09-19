package linkcache

// mode
const (
	LinkerRedisModeLocal = "mode_local"
	LinkerRedisModeRedis = "mode_redis"
)

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
