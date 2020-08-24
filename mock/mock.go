package mock

import "os"

var (
	// RedisAddr redis server address
	RedisAddr string

	// ConsulAddr consul server address
	ConsulAddr string

	// JaegerAddr jaeger server addr
	JaegerAddr string

	// NsqdAddr NSQd address
	NsqdAddr string

	// NSQLookupdAddr NSQ lookupd address
	NSQLookupdAddr string

	// Owner 用于在测试的时候隔离环境
	Owner string
)

const (
	mockRedisEnv      = "MOCK_REDIS_ADDR"
	mockConsulEnv     = "MOCK_CONSUL_ADDR"
	mockJaegerEnv     = "MOCK_JAEGER_ADDR"
	mockNsqdEnv       = "MOCK_NSQD_ADDR"
	mockNsqLookupdEnv = "MOCK_NSQ_LOOKUPD_ADDR"
	ownerEnv          = "DRONE_REPO_OWNER"
)

// Init 初始化测试环境
func Init() {
	// 构造测试环境
	RedisAddr = os.Getenv(mockRedisEnv)
	if RedisAddr == "" {
		RedisAddr = "redis://127.0.0.1:6379/0"
	}

	// 构造测试环境
	ConsulAddr = os.Getenv(mockConsulEnv)
	if ConsulAddr == "" {
		ConsulAddr = "http://127.0.0.1:8900"
	}

	// 构造测试环境
	JaegerAddr = os.Getenv(mockJaegerEnv)
	if JaegerAddr == "" {
		JaegerAddr = "http://127.0.0.1:9411/api/v2/spans"
	}

	NsqdAddr = os.Getenv(mockNsqdEnv)
	if NsqdAddr == "" {
		NsqdAddr = "127.0.0.1:4150"
	}

	NSQLookupdAddr = os.Getenv(mockNsqLookupdEnv)
	if NSQLookupdAddr == "" {
		NSQLookupdAddr = "127.0.0.1:4161"
	}

	Owner = os.Getenv(ownerEnv)
	if Owner == "" {
		Owner = "normal"
	}
}
