package braid

import (
	"context"
	"testing"
	"time"

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
        path : /var/log/coordinate
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
    discover :
        open : true
        interval : 3000
    caller : 
        open : true
`

func TestCompose(t *testing.T) {

	mock.Init()

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
