package linkcacheredis

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/pojol/braid-go/depend/blog"
	"github.com/pojol/braid-go/depend/bredis"
	"github.com/pojol/braid-go/mock"
	"github.com/pojol/braid-go/module/discover"
	"github.com/pojol/braid-go/module/elector"
	"github.com/pojol/braid-go/module/internal/pubsubnsq"
	"github.com/pojol/braid-go/module/linkcache"
	"github.com/pojol/braid-go/module/pubsub"
	"github.com/pojol/braid-go/service"
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

	ps := pubsubnsq.BuildWithOption(
		"",
		blog.BuildWithOption(),
		pubsub.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}),
	)
	redisclient := bredis.BuildWithOption(&redis.Options{
		Addr: mock.RedisAddr,
	})

	lc := BuildWithOption(
		LinkerRedisPrefix,
		blog.BuildWithOption(),
		ps,
		redisclient,
	)

	redisclient.Del(context.TODO(), LinkerRedisPrefix+"*")

	lc.Init()
	lc.Run()
	defer lc.Close()

	ps.LocalTopic(elector.TopicChangeState).Pub(elector.EncodeStateChangeMsg(elector.EMaster))

	nods := []service.Node{
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

	err := lc.Link("token01", nods[0])
	assert.Equal(t, err, nil)

	err = lc.Link("token01", nods[1])
	assert.Equal(t, err, nil)

	err = lc.Link("token02", nods[0])
	assert.Equal(t, err, nil)

	addr, err := lc.Target("token01", "base")
	assert.Equal(t, err, nil)
	assert.Equal(t, addr, "127.0.0.1:12001")

	_, err = lc.Target("unknowtoken", "base")
	assert.NotEqual(t, err, nil)

	ps.ClusterTopic("." + linkcache.TopicUnlink).Pub(&pubsub.Message{Body: []byte("token01")})
	ps.ClusterTopic("." + linkcache.TopicUnlink).Pub(&pubsub.Message{Body: []byte("token02")})

	time.Sleep(time.Millisecond * 500)

	for _, v := range nods {
		ps.LocalTopic(discover.TopicServiceUpdate).Pub(discover.EncodeUpdateMsg(discover.EventRemoveService, v))
	}

	time.Sleep(time.Millisecond * 100)

}

func TestLocalTarget(t *testing.T) {
	var tmu sync.Mutex
	tmu.Lock()
	// 用于生成测试用例使用的key前缀
	LinkerRedisPrefix := "TestLocalTarget-"
	tmu.Unlock()

	log := blog.BuildWithOption()

	ps := pubsubnsq.BuildWithOption(
		"",
		log,
		pubsub.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}),
	)
	redisclient := bredis.BuildWithOption(&redis.Options{Addr: mock.RedisAddr})

	lc := BuildWithOption(
		LinkerRedisPrefix,
		log,
		ps,
		redisclient,
		linkcache.WithMode(linkcache.LinkerRedisModeLocal),
	)

	redisclient.Del(context.TODO(), LinkerRedisPrefix+"*")

	lc.Init()
	lc.Run()
	defer lc.Close()

	ps.LocalTopic(elector.TopicChangeState).Pub(elector.EncodeStateChangeMsg(elector.EMaster))

	nods := []service.Node{
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

	err := lc.Link("localtoken01", nods[0])
	assert.Equal(t, err, nil)

	err = lc.Link("localtoken01", nods[1])
	assert.Equal(t, err, nil)

	err = lc.Link("localtoken02", nods[0])
	assert.Equal(t, err, nil)

	addr, err := lc.Target("localtoken01", "localbase")
	assert.Equal(t, err, nil)
	assert.Equal(t, addr, "127.0.0.1:12001")

	_, err = lc.Target("unknowtoken", "localbase")
	assert.NotEqual(t, err, nil)

	lc.Unlink("localtoken01")

	for _, v := range nods {
		lc.Down(v)
	}

	time.Sleep(time.Millisecond * 500)
}

/*
func BenchmarkLink(b *testing.B) {
	LinkerRedisPrefix := "benchmarklink"

	blog.New(blog.NewWithDefault())

	mbb := module.GetBuilder(pubsubnsq.Name)
	mbb.AddModuleOption(pubsubnsq.WithLookupAddr([]string{}))
	mbb.AddModuleOption(pubsubnsq.WithNsqdAddr([]string{mock.NsqdAddr}, []string{mock.NsqdHttpAddr}))
	mb := mbb.Build("BenchmarkLink").(pubsub.IPubsub)

	eb := module.GetBuilder(electorconsul.Name)
	eb.AddModuleOption(electorconsul.WithConsulAddr(mock.ConsulAddr))
	e := eb.Build("TestLinkerTarget",
		moduleparm.WithPubsub(mb)).(elector.IElector)
	defer e.Close()

	lb := module.GetBuilder(Name)
	lb.AddModuleOption(WithRedisAddr(mock.RedisAddr))

	lc := lb.Build("gate",
		moduleparm.WithPubsub(mb)).(linkcache.ILinkCache)
	rand.Seed(time.Now().UnixNano())

	lc.Init()
	lc.Run()
	defer lc.Close()

	baseTargets := []discover.Node{
		{
			ID:      "a001",
			Name:    "base",
			Address: "127.0.0.1:12001",
		},
		{
			ID:      "a002",
			Name:    "base",
			Address: "127.0.0.1:12002",
		},
		{
			ID:      "a003",
			Name:    "base",
			Address: "127.0.0.1:12003",
		},
	}

	loginTargets := []discover.Node{
		{
			ID:      "b001",
			Name:    "login",
			Address: "127.0.0.1:13001",
		},
		{
			ID:      "b002",
			Name:    "login",
			Address: "127.0.0.1:13001",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lc.Link("token"+strconv.Itoa(i), baseTargets[rand.Intn(len(baseTargets))])
		lc.Link("token"+strconv.Itoa(i), loginTargets[rand.Intn(len(loginTargets))])
	}
}
*/
