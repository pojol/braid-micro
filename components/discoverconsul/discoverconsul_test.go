package discoverconsul

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
)

func TestMain(m *testing.M) {
	mock.Init()

	m.Run()
}

func TestDiscover(t *testing.T) {

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

	dc := BuildWithOption(
		meta.ServiceInfo{ID: "id", Name: "name"},
		log,
		consulcli,
		redisps,
		WithSyncServiceInterval(time.Millisecond*100),
		WithSyncServiceWeightInterval(time.Millisecond*100),
		WithBlacklist([]string{"gate"}),
		WithTag("braid"),
	)

	dc.Init()
	dc.Run()

	time.Sleep(time.Second)
	dc.Close()

}
