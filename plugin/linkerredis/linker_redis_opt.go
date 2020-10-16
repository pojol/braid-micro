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
