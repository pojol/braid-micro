package linkerredis

// mode
const (
	LinkerRedisModeLocal = "mode_local"
	LinkerRedisModeRedis = "mode_redis"
)

// Parm Service 配置
type Parm struct {
	Mode           string
	SyncTick       int // ms
	RedisAddr      string
	RedisMaxIdle   int
	RedisMaxActive int

	//
	syncRelationTick int // second

	//
	syncOfflineTick int // second
}

// Option config wraps
type Option func(*Parm)

// WithRedisAddr with redis addr
func WithRedisAddr(addr string) Option {
	return func(c *Parm) {
		c.RedisAddr = addr
	}
}

// WithRedisMaxIdle 修改redis最大空闲连接
func WithRedisMaxIdle(maxIdle int) Option {
	return func(c *Parm) {
		c.RedisMaxIdle = maxIdle
	}
}

// WithRedisMaxActive 修改redis最大活跃连接
func WithRedisMaxActive(maxActive int) Option {
	return func(c *Parm) {
		c.RedisMaxActive = maxActive
	}
}

// WithSyncTick 同步周期
func WithSyncTick(mstime int) Option {
	return func(c *Parm) {
		c.SyncTick = mstime
	}
}

// WithMode 设置redis link-cache的执行模式
func WithMode(mode string) Option {
	return func(c *Parm) {
		c.Mode = mode
	}
}
