package tracer

import (
	"testing"
)

func TestTracer(t *testing.T) {
	/*
		tr, err := New("test", WithHTTP("http://127.0.0.1:14268/api/traces"))
		assert.Equal(t, err, nil)
		defer tr.Close()

		span := tr.tracing.StartSpan("test-1")
		span.Finish()
	*/
}

func TestTracerWithUDP(t *testing.T) {
	/*
		tr, err := New("test", WithUDP("127.0.0.1:6831"))
		assert.Equal(t, err, nil)
		defer tr.Close()

		span := tr.tracing.StartSpan("test-udp")
		span.Finish()
	*/
}
