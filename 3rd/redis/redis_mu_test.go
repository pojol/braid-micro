package redis

import (
	"strconv"
	"testing"

	"github.com/pojol/braid-go/mock"
	"github.com/stretchr/testify/assert"
)

func TestRedisMu(t *testing.T) {

	c := New()
	defaultConnPoolConfig.Address = mock.RedisAddr
	err := c.Init(defaultConnPoolConfig)
	if err != nil {
		t.Error(err)
	}

	mu := Mutex{
		Token: "12345678",
	}
	err = mu.Lock("test")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, mu.Lock("test"), ErrRedisMtxAcquire)
	assert.Equal(t, mu.Unlock(), true)

	Get().flushDB()
	Get().Close()
}

func TestNilToken(t *testing.T) {
	c := New()
	defaultConnPoolConfig.Address = mock.RedisAddr
	err := c.Init(defaultConnPoolConfig)
	if err != nil {
		t.Error(err)
	}

	mu := Mutex{
		Token: "",
	}
	err = mu.Lock("test")
	assert.Equal(t, err, ErrRedisMtxInvaildToken)
}

func BenchmarkRedisMu(b *testing.B) {
	c := New()
	defaultConnPoolConfig.Address = mock.RedisAddr
	err := c.Init(defaultConnPoolConfig)
	if err != nil {
		b.Error(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mu := Mutex{
			Token: strconv.Itoa(i),
		}
		err := mu.Lock("benchmark")
		if err == nil {
			mu.Unlock()
		}
	}
}
