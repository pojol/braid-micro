package balancer

import (
	"strconv"
	"testing"
	"time"

	"github.com/pojol/braid/log"
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

	bw := newBalancerWrapper(GetBuilder(balancerName))

	bw.Update(Node{
		ID:      "A",
		Address: "A",
		Weight:  4,
		Name:    "test",
		OpTag:   OpAdd,
	})
	bw.Update(Node{
		ID:      "B",
		Address: "B",
		Weight:  2,
		Name:    "test",
		OpTag:   OpAdd,
	})
	bw.Update(Node{
		ID:      "C",
		Address: "C",
		Weight:  1,
		Name:    "test",
		OpTag:   OpAdd,
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
	bw := newBalancerWrapper(GetBuilder(balancerName))

	bw.Update(Node{
		ID:     "A",
		Name:   "test",
		Weight: 4,
		OpTag:  OpAdd,
	})

	bw.Update(Node{
		ID:    "A",
		Name:  "test",
		OpTag: OpRmv,
	})

	bw.Update(Node{
		ID:     "B",
		Name:   "test",
		Weight: 2,
		OpTag:  OpAdd,
	})
	bw.Update(Node{
		ID:    "B",
		Name:  "test",
		OpTag: OpUp,
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

	bw := newBalancerWrapper(GetBuilder(balancerName))

	for i := 0; i < 100; i++ {
		bw.Update(Node{
			ID:     strconv.Itoa(i),
			Name:   "test",
			Weight: i,
			OpTag:  OpAdd,
		})
	}

	//time.Sleep(time.Millisecond * 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bw.Pick()
	}
}
