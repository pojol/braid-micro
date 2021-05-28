package balancernormal

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pojol/braid-go/module"
	"github.com/pojol/braid-go/module/balancer"
	"github.com/pojol/braid-go/module/discover"
	"github.com/pojol/braid-go/module/logger"
	"github.com/pojol/braid-go/module/pubsub"
	"github.com/pojol/braid-go/modules/moduleparm"
	"github.com/pojol/braid-go/modules/pubsubnsq"
	"github.com/pojol/braid-go/modules/zaplogger"
	"github.com/stretchr/testify/assert"
)

func TestRandomBalancer(t *testing.T) {

	serviceName := "TestRandomBalancer"

	log := module.GetBuilder(zaplogger.Name).Build(serviceName).(logger.ILogger)
	mb := module.GetBuilder(pubsubnsq.Name).Build(serviceName, moduleparm.WithLogger(log)).(pubsub.IPubsub)

	bgb := module.GetBuilder(Name)
	b := bgb.Build(serviceName,
		moduleparm.WithLogger(log),
		moduleparm.WithPubsub(mb))
	bg := b.(balancer.IBalancer)

	bg.Init()
	bg.Run()
	defer bg.Close()

	var atick, btick, ctick uint64
	_, err := bg.Pick(StrategyRandom, serviceName)
	assert.NotEqual(t, err, nil)

	mb.GetTopic(discover.ServiceUpdate).Pub(discover.EncodeUpdateMsg(
		discover.EventAddService,
		discover.Node{
			ID:      "A",
			Address: "A",
			Weight:  4,
			Name:    serviceName,
		},
	))
	mb.GetTopic(discover.ServiceUpdate).Pub(discover.EncodeUpdateMsg(
		discover.EventAddService,
		discover.Node{
			ID:      "B",
			Address: "B",
			Weight:  2,
			Name:    serviceName,
		},
	))
	mb.GetTopic(discover.ServiceUpdate).Pub(discover.EncodeUpdateMsg(
		discover.EventAddService,
		discover.Node{
			ID:      "C",
			Address: "C",
			Weight:  1,
			Name:    serviceName,
		},
	))

	time.Sleep(time.Millisecond * 100)
	fmt.Println("begin pick")
	for i := 0; i < 30000; i++ {
		nod, err := bg.Pick(StrategyRandom, serviceName)
		assert.Equal(t, err, nil)

		if nod.ID == "A" {
			atomic.AddUint64(&atick, 1)
		} else if nod.ID == "B" {
			atomic.AddUint64(&btick, 1)
		} else if nod.ID == "C" {
			atomic.AddUint64(&ctick, 1)
		}
	}

	fmt.Println(atomic.LoadUint64(&atick), atomic.LoadUint64(&btick), atomic.LoadUint64(&ctick))
	assert.Equal(t, true, (atomic.LoadUint64(&atick) >= 9000 && atomic.LoadUint64(&atick) <= 11000))
	assert.Equal(t, true, (atomic.LoadUint64(&btick) >= 9000 && atomic.LoadUint64(&btick) <= 11000))
	assert.Equal(t, true, (atomic.LoadUint64(&ctick) >= 9000 && atomic.LoadUint64(&ctick) <= 11000))

	mb.GetTopic(discover.ServiceUpdate).Pub(discover.EncodeUpdateMsg(
		discover.EventRemoveService,
		discover.Node{
			ID:      "C",
			Address: "C",
			Name:    serviceName,
		},
	))

	time.Sleep(time.Millisecond * 100)
	atomic.SwapUint64(&atick, 0)
	atomic.SwapUint64(&btick, 0)
	atomic.SwapUint64(&ctick, 0)

	for i := 0; i < 20000; i++ {
		nod, _ := bg.Pick(StrategyRandom, serviceName)
		if nod.ID == "A" {
			atomic.AddUint64(&atick, 1)
		} else if nod.ID == "B" {
			atomic.AddUint64(&btick, 1)
		} else if nod.ID == "C" {
			atomic.AddUint64(&ctick, 1)
		}
	}
	assert.Equal(t, true, (atomic.LoadUint64(&atick) >= 9000 && atomic.LoadUint64(&atick) <= 11000))
	assert.Equal(t, true, (atomic.LoadUint64(&btick) >= 9000 && atomic.LoadUint64(&btick) <= 11000))
	assert.Equal(t, true, (atomic.LoadUint64(&ctick) == 0))
}
