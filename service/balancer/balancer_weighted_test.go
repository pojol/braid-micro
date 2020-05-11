package balancer

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWRR(t *testing.T) {

	wrr := WeightedRoundrobin{}

	wrr.Add(Node{
		ID:     "A",
		Weight: 4,
	})
	wrr.Add(Node{
		ID:     "B",
		Weight: 2,
	})
	wrr.Add(Node{
		ID:     "C",
		Weight: 1,
	})

	var tests = []struct {
		ID string
	}{
		{"A"}, {"B"}, {"A"}, {"C"}, {"A"}, {"B"}, {"A"},
	}

	for _, v := range tests {
		n, _ := wrr.Next()
		assert.Equal(t, n.ID, v.ID)
	}

}

func TestWRROp(t *testing.T) {

	wrr := WeightedRoundrobin{}

	wrr.Add(Node{
		ID:     "A",
		Weight: 4,
	})

	wrr.Rmv("A")

	_, ok := wrr.isExist("A")
	assert.Equal(t, ok, false)

	wrr.Add(Node{
		ID:     "B",
		Weight: 2,
	})
	wrr.SyncWeight("B", 4)
	n, _ := wrr.Next()
	assert.Equal(t, n.Weight, wrr.totalWeight)
}

func TestSelector(t *testing.T) {
	s := New()
	s.Init(Cfg{})
	s.Run()
	defer s.Close()

	ib, _ := GetGroup("test")

	ib.Add(Node{
		ID:     "A",
		Weight: 4,
	})

	ib.Rmv("A")
}

func BenchmarkWRR(b *testing.B) {
	wrr := WeightedRoundrobin{}

	for i := 0; i < 100; i++ {
		wrr.Add(Node{
			ID:     strconv.Itoa(i),
			Weight: i,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wrr.Next()
	}
}
