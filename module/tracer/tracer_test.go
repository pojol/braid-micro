package tracer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTracer(t *testing.T) {

	tr, err := New("test", "http://127.0.0.1:14268/api/traces")
	assert.Equal(t, err, nil)
	defer tr.Close()

	span := tr.tracing.StartSpan("test-1")
	span.Finish()

}
