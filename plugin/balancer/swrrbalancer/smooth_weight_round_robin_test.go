package swrrbalancer

import (
	"strconv"
	"testing"
	"time"

	"github.com/pojol/braid/3rd/log"
	"github.com/pojol/braid/plugin/balancer"
	"github.com/stretchr/testify/assert"
)

func TestWRR(t *testing.T) {

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

	g := balancer.NewGroup()
	bw := g.Get("test")

	bw.Update(balancer.Node{
		ID:      "A",
		Address: "A",
		Weight:  4,
		Name:    "test",
		OpTag:   balancer.OpAdd,
	})
	bw.Update(balancer.Node{
		ID:      "B",
		Address: "B",
		Weight:  2,
		Name:    "test",
		OpTag:   balancer.OpAdd,
	})
	bw.Update(balancer.Node{
		ID:      "C",
		Address: "C",
		Weight:  1,
		Name:    "test",
		OpTag:   balancer.OpAdd,
	})

	var tests = []struct {
		ID string
	}{
		{"A"}, {"B"}, {"A"}, {"C"}, {"A"}, {"B"}, {"A"},
	}

	time.Sleep(time.Millisecond * 100)
	for _, v := range tests {
		addr, _ := bw.Pick()
		assert.Equal(t, addr, v.ID)
	}

}

func TestWRROp(t *testing.T) {

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
	g := balancer.NewGroup()
	bw := g.Get("test")

	bw.Update(balancer.Node{
		ID:     "A",
		Name:   "test",
		Weight: 4,
		OpTag:  balancer.OpAdd,
	})

	bw.Update(balancer.Node{
		ID:    "A",
		Name:  "test",
		OpTag: balancer.OpRmv,
	})

	bw.Update(balancer.Node{
		ID:     "B",
		Name:   "test",
		Weight: 2,
		OpTag:  balancer.OpAdd,
	})
	bw.Update(balancer.Node{
		ID:    "B",
		Name:  "test",
		OpTag: balancer.OpUp,
	})

	bw.Pick()
}

//  2637153	       442 ns/op	       0 B/op	       0 allocs/op
func BenchmarkWRR(b *testing.B) {
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

	g := balancer.NewGroup()
	bw := g.Get("test")

	for i := 0; i < 100; i++ {
		bw.Update(balancer.Node{
			ID:     strconv.Itoa(i),
			Name:   "test",
			Weight: i,
			OpTag:  balancer.OpAdd,
		})
	}

	//time.Sleep(time.Millisecond * 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bw.Pick()
	}
}
