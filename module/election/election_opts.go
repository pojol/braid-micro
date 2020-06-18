package election

import "time"

// config 选举器配置项
type config struct {
	Address           string
	Name              string
	LockTick          time.Duration
	RefushSessionTick time.Duration
}

// Option config wraps
type Option func(*Election)

// WithLockTick 竞争锁的频率 ms
func WithLockTick(ms int) Option {
	return func(e *Election) {
		e.cfg.LockTick = time.Duration(ms) * time.Millisecond
	}
}

// WithRefushTick 心跳保活的频率
func WithRefushTick(ms int) Option {
	return func(e *Election) {
		e.cfg.RefushSessionTick = time.Duration(ms) * time.Millisecond
	}
}
