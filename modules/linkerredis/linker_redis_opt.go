package linkerredis

// Parm Service 配置
type Parm struct {
	RedisAddr      string
	RedisMaxIdle   int
	RedisMaxActive int
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
