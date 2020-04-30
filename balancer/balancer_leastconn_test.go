package balancer

import (
	"strconv"
	"testing"

	"github.com/pojol/braid/log"
	"github.com/stretchr/testify/assert"
)

func TestLeastConnBalancer(t *testing.T) {

	b := New()
	b.Init(SelectorCfg{})

	l := log.New()
	l.Init(log.Config{
		Path:   "test",
		Suffex: ".log",
		Mode:   "debug",
	})

	_, err := GetSelector("test").Next()
	assert.Equal(t, err, ErrBalanceEmpty)

	for i := 0; i < 4; i++ {
		GetSelector("test").Add(Node{
			ID:     strconv.Itoa(i),
			Tick:   0,
			Weight: 1,
		})
	}

	GetSelector("test").Add(Node{
		ID: "0",
	})
	GetSelector("test").Rmv("3")

	for i := 0; i < 3; i++ {
		n, err := GetSelector("test").Next()
		assert.Equal(t, err, nil)
		assert.Equal(t, n.ID, strconv.Itoa(i))
	}
}

func BenchmarkLeastConnBalancer(b *testing.B) {

	wrr := LeastConnBalancer{}

	for i := 0; i < 100; i++ {
		wrr.Add(Node{
			ID:     strconv.Itoa(i),
			Tick:   0,
			Weight: 1,
		})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wrr.Next()
	}

}
