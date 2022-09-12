package pubsub

import "github.com/nsqio/go-nsq"

// Parm nsq config
type Parm struct {
	nsqCfg nsq.Config

	LookupdAddress  []string // lookupd 地址
	NsqdAddress     []string
	NsqdHttpAddress []string

	ServiceName string

	ConcurrentHandler int32 // consumer 接收句柄的并发数（默认1

	HA bool // 是否开启高可用，向每个nsqd发送消息

	ChannelLength int32 // 管道长度，如果设置为0则全部消息都落地到磁盘再进行消费

	NsqLogLv nsq.LogLevel
}

var (
	DefaultConfig = Parm{
		NsqLogLv:          nsq.LogLevelInfo,
		ConcurrentHandler: 1,
	}
)

func NewWithOption(opts ...Option) *Parm {
	logParm := &DefaultConfig

	for _, opt := range opts {
		opt(logParm)
	}

	return logParm
}

func NewWithDefault() *Parm {
	return &DefaultConfig
}

// Option config wraps
type Option func(*Parm)

// WithChannel 通过 channel 构建
func WithChannelSize(channelsize int32) Option {
	return func(c *Parm) {
		c.ChannelLength = channelsize
	}
}

// WithNsqConfig nsq config
func WithNsqConfig(cfg nsq.Config) Option {
	return func(c *Parm) {
		c.nsqCfg = cfg
	}
}

// WithLookupAddr lookup addr
func WithLookupAddr(addr []string) Option {
	return func(c *Parm) {
		c.LookupdAddress = addr
	}
}

// WithNsqdAddr nsqd addr
func WithNsqdAddr(tcpAddr []string, httpAddr []string) Option {
	return func(c *Parm) {
		if len(tcpAddr) != len(httpAddr) {
			panic("The addresses of tcp and http should match")
		}

		c.NsqdAddress = tcpAddr
		c.NsqdHttpAddress = httpAddr
	}
}

// WithNsqLogLv 修改nsq的日志等级
func WithNsqLogLv(lv nsq.LogLevel) Option {
	return func(c *Parm) {
		c.NsqLogLv = lv
	}
}

// WithHandlerConcurrent 消费者接收句柄的并发数量（默认1
func WithHandlerConcurrent(cnt int32) Option {
	return func(c *Parm) {
		c.ConcurrentHandler = cnt
	}
}
