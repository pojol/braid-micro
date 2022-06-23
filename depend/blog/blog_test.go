package blog

import (
	"testing"
)

func TestLog(t *testing.T) {

	BuildWithNormal()
	defer Close()

	Infof("msg %v", 1)
}
