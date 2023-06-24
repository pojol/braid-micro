package electorconsul

import (
	"testing"
	"time"

	"github.com/pojol/braid-go/components/depends/bconsul"
	"github.com/pojol/braid-go/components/depends/blog"
	"github.com/pojol/braid-go/components/depends/bredis"
	"github.com/pojol/braid-go/components/pubsubredis"
	"github.com/pojol/braid-go/mock"
	"github.com/pojol/braid-go/module/meta"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	mock.Init()
	m.Run()
}

func TestElection(t *testing.T) {

	log := blog.BuildWithOption()

	rediscli := bredis.BuildWithOption(&redis.Options{
		Addr: mock.RedisAddr,
	})
	consulcli := bconsul.BuildWithOption(bconsul.WithAddress([]string{mock.ConsulAddr}))

	redisps := pubsubredis.BuildWithOption(
		meta.ServiceInfo{ID: "id", Name: "name"},
		log,
		rediscli,
	)

	e := BuildWithOption(
		meta.ServiceInfo{ID: "id", Name: "name"},
		WithLog(log),
		WithConsulClient(consulcli),
		WithPubsub(redisps),
		WithLockTick(time.Second),
	)

	e.Run()
	time.Sleep(time.Second)
	e.Close()
}

func TestParm(t *testing.T) {

	log := blog.BuildWithOption()

	rediscli := bredis.BuildWithOption(&redis.Options{
		Addr: mock.RedisAddr,
	})
	consulcli := bconsul.BuildWithOption(bconsul.WithAddress([]string{mock.ConsulAddr}))

	redisps := pubsubredis.BuildWithOption(
		meta.ServiceInfo{ID: "id", Name: "name"},
		log,
		rediscli,
	)

	e := BuildWithOption(
		meta.ServiceInfo{ID: "id", Name: "name"},
		WithLog(log),
		WithConsulClient(consulcli),
		WithPubsub(redisps),
		WithLockTick(time.Second),
		WithSessionTick(time.Second),
	)

	ce := e.(*consulElection)

	assert.Equal(t, ce.parm.LockTick, time.Second)
	assert.Equal(t, ce.parm.RefushSessionTick, time.Second)
}
