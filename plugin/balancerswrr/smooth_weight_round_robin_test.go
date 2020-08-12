package balancerswrr

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/module/balancer"
	"github.com/pojol/braid/module/discover"
	"github.com/pojol/braid/module/pubsub"
	"github.com/pojol/braid/plugin/pubsubproc"
	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.M) {
	l := log.New(log.Config{
		Mode:   log.DebugMode,
		Path:   "testNormal",
		Suffex: ".log",
	}, log.WithSys(log.Config{
		Mode:   log.DebugMode,
		Path:   "testSys",
		Suffex: ".sys",
	}))
	defer l.Close()

	t.Run()
}

func TestWRR(t *testing.T) {
	ps := pubsub.GetBuilder(pubsubproc.PubsubName).Build()
	balancer.NewGroup(balancer.GetBuilder(BalancerName), ps)
	bw := balancer.Get("test")

	ps.Pub(discover.EventAdd, discover.Node{
		ID:      "A",
		Address: "A",
		Weight:  4,
		Name:    "test",
	})

	ps.Pub(discover.EventAdd, discover.Node{
		ID:      "B",
		Address: "B",
		Weight:  2,
		Name:    "test",
	})

	ps.Pub(discover.EventAdd, discover.Node{
		ID:      "C",
		Address: "C",
		Weight:  1,
		Name:    "test",
	})

	var tests = []struct {
		ID string
	}{
		{"A"}, {"B"}, {"A"}, {"C"}, {"A"}, {"B"}, {"A"},
	}

	time.Sleep(time.Millisecond * 100)
	for _, v := range tests {
		nod, _ := bw.Pick()
		assert.Equal(t, nod.Address, v.ID)
	}

}

func TestWRRDymc(t *testing.T) {
	ps := pubsub.GetBuilder(pubsubproc.PubsubName).Build()
	balancer.NewGroup(balancer.GetBuilder(BalancerName), ps)
	bw := balancer.Get("test")
	pmap := make(map[string]int)

	ps.Pub(discover.EventAdd, discover.Node{
		ID:      "A",
		Address: "A",
		Weight:  1000,
		Name:    "test",
	})

	ps.Pub(discover.EventAdd, discover.Node{
		ID:      "B",
		Address: "B",
		Weight:  1000,
		Name:    "test",
	})

	ps.Pub(discover.EventAdd, discover.Node{
		ID:      "C",
		Address: "C",
		Weight:  1000,
		Name:    "test",
	})

	time.Sleep(time.Millisecond * 100)

	for i := 0; i < 100; i++ {
		nod, _ := bw.Pick()
		pmap[nod.ID]++
	}

	fmt.Println("step 1", pmap)

	ps.Pub(discover.EventUpdate, discover.Node{
		ID:     "A",
		Weight: 500,
	})
	time.Sleep(time.Millisecond * 100)

	for i := 0; i < 100; i++ {
		nod, _ := bw.Pick()
		pmap[nod.ID]++
	}

	fmt.Println("step 2", pmap)
}

func TestWRROp(t *testing.T) {

	ps := pubsub.GetBuilder(pubsubproc.PubsubName).Build()
	balancer.NewGroup(balancer.GetBuilder(BalancerName), ps)
	bw := balancer.Get("test")

	ps.Pub(discover.EventAdd, discover.Node{
		ID:     "A",
		Name:   "test",
		Weight: 4,
	})

	ps.Pub(discover.EventRmv, discover.Node{
		ID:   "A",
		Name: "test",
	})

	ps.Pub(discover.EventAdd, discover.Node{
		ID:     "B",
		Name:   "test",
		Weight: 2,
	})

	bw.Pick()
}

//  2637153	       442 ns/op	       0 B/op	       0 allocs/op
func BenchmarkWRR(b *testing.B) {
	ps := pubsub.GetBuilder(pubsubproc.PubsubName).Build()
	balancer.NewGroup(balancer.GetBuilder(BalancerName), ps)
	bw := balancer.Get("test")

	for i := 0; i < 100; i++ {
		ps.Pub(discover.EventAdd, discover.Node{
			ID:     strconv.Itoa(i),
			Name:   "test",
			Weight: i,
		})
	}

	time.Sleep(time.Millisecond * 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bw.Pick()
	}
}
