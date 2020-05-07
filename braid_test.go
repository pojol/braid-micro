package braid

import (
	"context"
	"testing"
	"time"

	"github.com/pojol/braid/log"
	"github.com/pojol/braid/mock"
	"gopkg.in/yaml.v2"
)

var TestComposeFile = `
name : coordinate
mode : debug
tracing : false

install : 
    log :
        open : true
        path : test
        suffex : .sys

    redis :
        open : true
        read_timeout : 5000
        write_timeout : 5000
        conn_timeout : 2000
        idle_timeout : 0
        max_idle : 16
        max_active : 128

    service :
        open : true
    linker :
        open : true
    balancer :
        open : true
    election :
        open : true
        lock_tick : 2000
        refush_tick : 5000
    discover :
        open : true
        interval : 3000
    caller : 
        open : true
        pool_init_num : 32
        pool_cap : 128
        pool_idle : 120
    tracer : 
        open : true
        probabilistic : 0.01
        slow_req : 100
        slow_span : 20
`

func TestCompose(t *testing.T) {

	mock.Init()

	l := log.New()
	l.Init(log.Config{
		Path:   "test",
		Suffex: ".log",
		Mode:   "debug",
	})

	conf := &ComposeConf{}
	err := yaml.Unmarshal([]byte(TestComposeFile), conf)
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
