package blog

import (
	"testing"
)

func TestLog(t *testing.T) {

	BuildWithOption()
	defer Close()

	Infof("msg %v", 1)
}
