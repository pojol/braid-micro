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
