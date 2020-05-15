package discover

import (
	"time"
)

type config struct {
	Name string

	// 同步节点信息间隔
	Interval time.Duration

	ConsulAddress string
}

// Option config wraps
type Option func(*Discover)

// WithInterval 发现节点频率
func WithInterval(ms int) Option {
	return func(dc *Discover) {
		dc.cfg.Interval = time.Duration(ms) * time.Millisecond
	}
}
