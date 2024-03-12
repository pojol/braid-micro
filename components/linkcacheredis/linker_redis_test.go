package linkcacheredis

import (
	"context"
	"sync"
	"testing"
	"time"

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

func TestLinkerTarget(t *testing.T) {

	var tmu sync.Mutex
	tmu.Lock()
	// 用于生成测试用例使用的key前缀
	LinkerRedisPrefix := "TestLinkerTarget-"
	tmu.Unlock()

	rediscli := bredis.BuildWithOption(&redis.Options{
		Addr: mock.RedisAddr,
	})
	log := blog.BuildWithOption()

	redisps := pubsubredis.BuildWithOption(
		meta.ServiceInfo{ID: "id", Name: "name"},
		log,
		rediscli,
	)

	lc := BuildWithOption(
		meta.ServiceInfo{ID: "id", Name: "name"},
		log,
		redisps,
		rediscli,
	)

	rediscli.Del(context.TODO(), LinkerRedisPrefix+"*")

	lc.Init()
	lc.Run()
	defer lc.Close()

	redisps.GetTopic(meta.TopicElectionChangeState).Pub(context.TODO(),
		meta.EncodeStateChangeMsg(meta.EMaster, "id"))

	nods := []meta.Node{
		{
			ID:      "a001",
			Name:    "base",
			Address: "127.0.0.1:12001",
		},
		{
			ID:      "a002",
			Name:    "login",
			Address: "127.0.0.1:13001",
		},
	}

	err := lc.Link(context.TODO(), "token01", nods[0])
	assert.Equal(t, err, nil)

	err = lc.Link(context.TODO(), "token01", nods[1])
	assert.Equal(t, err, nil)

	err = lc.Link(context.TODO(), "token02", nods[0])
	assert.Equal(t, err, nil)

	addr, err := lc.Target(context.TODO(), "token01", "base")
	assert.Equal(t, err, nil)
	assert.Equal(t, addr, "127.0.0.1:12001")

	_, err = lc.Target(context.TODO(), "unknowtoken", "base")
	assert.NotEqual(t, err, nil)

	redisps.GetTopic(meta.TopicLinkcacheUnlink).Pub(context.TODO(), &meta.Message{Body: []byte("token01")})
	redisps.GetTopic(meta.TopicLinkcacheUnlink).Pub(context.TODO(), &meta.Message{Body: []byte("token02")})

	time.Sleep(time.Millisecond * 500)

	for _, v := range nods {
		redisps.GetTopic(meta.TopicDiscoverServiceUpdate).Pub(context.TODO(), meta.EncodeUpdateMsg(meta.TopicDiscoverServiceNodeRmv, v))
	}

	time.Sleep(time.Millisecond * 100)
}

func TestLocalTarget(t *testing.T) {
	var tmu sync.Mutex
	tmu.Lock()
	// 用于生成测试用例使用的key前缀
	LinkerRedisPrefix := "TestLocalTarget-"
	tmu.Unlock()

	rediscli := bredis.BuildWithOption(&redis.Options{
		Addr: mock.RedisAddr,
	})
	log := blog.BuildWithOption()

	redisps := pubsubredis.BuildWithOption(
		meta.ServiceInfo{ID: "id", Name: "name"},
		log,
		rediscli,
	)

	lc := BuildWithOption(
		meta.ServiceInfo{ID: "id", Name: "name"},
		log,
		redisps,
		rediscli,
	)

	rediscli.Del(context.TODO(), LinkerRedisPrefix+"*")

	lc.Init()
	lc.Run()
	defer lc.Close()

	redisps.GetTopic(meta.TopicElectionChangeState).Pub(context.TODO(),
		meta.EncodeStateChangeMsg(meta.EMaster, "id"))

	nods := []meta.Node{
		{
			ID:      "local001",
			Name:    "localbase",
			Address: "127.0.0.1:12001",
		},
		{
			ID:      "local002",
			Name:    "locallogin",
			Address: "127.0.0.1:13001",
		},
	}

	err := lc.Link(context.TODO(), "localtoken01", nods[0])
	assert.Equal(t, err, nil)

	err = lc.Link(context.TODO(), "localtoken01", nods[1])
	assert.Equal(t, err, nil)

	err = lc.Link(context.TODO(), "localtoken02", nods[0])
	assert.Equal(t, err, nil)

	addr, err := lc.Target(context.TODO(), "localtoken01", "localbase")
	assert.Equal(t, err, nil)
	assert.Equal(t, addr, "127.0.0.1:12001")

	_, err = lc.Target(context.TODO(), "unknowtoken", "localbase")
	assert.NotEqual(t, err, nil)

	lc.Unlink(context.TODO(), "localtoken01")
	/*
		for _, v := range nods {
			lc.Down(v)
		}
	*/
	time.Sleep(time.Millisecond * 500)
}
