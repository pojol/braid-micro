package blog

import (
	"testing"
)

func TestLog(t *testing.T) {

	BuildWithOption()
	defer Close()

	logPtr.Infof("msg %v", 1)
}
