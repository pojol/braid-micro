package mock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMock(t *testing.T) {

	Init()

	assert.NotEqual(t, mockRedisEnv, "")
	assert.NotEqual(t, mockConsulEnv, "")
	assert.NotEqual(t, mockJaegerEnv, "")

}
