package braid

import (
	"context"
	"testing"
	"time"

	"github.com/pojol/braid/mock"
	"gopkg.in/yaml.v2"
)

var ComposeFile = `
    name : serviceName
    mode : debug
    tracing : true

    # 安装模块列表
    install :
        - logger
        - register
        - election
        - rpc
        - tracer

    # 配置安装了的模块参数，不填写使用默认配置。
    config :
        logger_path : test
        logger_suffex : .sys

        # rpc 服务的监听端口
        register_listen_port : 14222

        # 选举器尝试成为主节点的频率 ms
        election_lock_tick : 2000
        # 选举器刷新保活的频率 ms
        election_refush_tick : 5000

        # 发现节点的刷新频率 ms
        rpc_discover_interval : 2000
        # rpc池的初始化大小
        rpc_pool_init_num : 32
        # rpc池的大小
        rpc_pool_cap : 128
        # rpc池的闲置超时时间 second
        rpc_pool_idle : 120

        # 追踪器的采用频率
        tracer_probabilistic : 0.01
        # 追踪器的慢查询超时 ms
        tracer_slow_req : 100
        # 追踪器的慢span超时
        tracer_slow_span : 20
`

func TestCompose(t *testing.T) {

	mock.Init()

	conf := &ComposeConf{}
	err := yaml.Unmarshal([]byte(ComposeFile), conf)
	if err != nil {
		t.Error(err)
	}

	Compose(*conf, DependConf{
		Consul: mock.ConsulAddr,
		Redis:  mock.RedisAddr,
		Jaeger: mock.JaegerAddr,
	})

	Regist("test", func(ctx context.Context, in []byte) (out []byte, err error) {
		return nil, nil
	})

	IsMaster()

	Run()
	time.Sleep(time.Second)
	Close()

}
