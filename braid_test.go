package braid

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pojol/braid/election"
	"github.com/pojol/braid/service"
	"github.com/pojol/braid/tracer"

	"github.com/pojol/braid/mock"

	"github.com/pojol/braid/log"

	"github.com/pojol/braid/balancer"
	"github.com/pojol/braid/link"

	"github.com/pojol/braid/cache/redis"
)

func TestCompose(t *testing.T) {

	mock.Init()

	exePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		t.Error(err)
	}

	Compose([]NodeCompose{
		NodeCompose{Ty: Logger, Cfg: log.Config{
			Mode:   "debug",
			Path:   exePath,
			Suffex: ".box",
		}},
		NodeCompose{Ty: Redis, Cfg: redis.Config{
			Address: mock.RedisAddr,
		}},
		NodeCompose{Ty: Linker, Cfg: link.Config{}},
		NodeCompose{Ty: Balancer, Cfg: balancer.SelectorCfg{}},
		/*
			NodeCompose{Ty: Caller, Cfg: caller.Config{
				ConsulAddress: mock.ConsulAddr,
				PoolCapacity:  32,
				PoolIdle:      time.Second * 10,
			}},
		*/
		/*
			NodeCompose{Ty: Discover, Cfg: discover.Config{}},
		*/

		NodeCompose{Ty: Election, Cfg: election.Config{
			Address: mock.ConsulAddr,
			Name:    "test",
		}},
		NodeCompose{Ty: Service, Cfg: service.Config{
			Tracing: true,
			Name:    "test",
		}},
		NodeCompose{Ty: Tracer, Cfg: tracer.Config{
			Endpoint:      mock.JaegerAddr,
			Name:          "test",
			Probabilistic: 0.9,
		}}},
	)

	Regist("test", func(ctx context.Context, in []byte) (out []byte, err error) {
		return nil, nil
	})

	Run()
	time.Sleep(time.Second)
	Close()

}
