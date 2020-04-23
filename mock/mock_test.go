package mock

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMock(t *testing.T) {

	tredis := os.Getenv(mockRedisEnv)
	tconsul := os.Getenv(mockConsulEnv)
	tjaeger := os.Getenv(mockJaegerEnv)

	os.Setenv(mockRedisEnv, "")
	os.Setenv(mockConsulEnv, "")
	os.Setenv(mockJaegerEnv, "")

	Init()

	assert.NotEqual(t, mockRedisEnv, "")
	assert.NotEqual(t, mockConsulEnv, "")
	assert.NotEqual(t, mockJaegerEnv, "")

	os.Setenv(mockRedisEnv, tredis)
	os.Setenv(mockConsulEnv, tconsul)
	os.Setenv(mockJaegerEnv, tjaeger)

	Init()

	assert.NotEqual(t, mockRedisEnv, "")
	assert.NotEqual(t, mockConsulEnv, "")
	assert.NotEqual(t, mockJaegerEnv, "")
}
