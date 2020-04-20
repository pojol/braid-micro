package service

import (
	"testing"
)

func TestNew(t *testing.T) {

	s := New()
	err := s.Init(Config{
		Tracing: false,
		Name:    "test",
	})
	if err != nil {
		t.Error(err)
	}

}
