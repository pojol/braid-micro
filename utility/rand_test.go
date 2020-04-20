package utility

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRandSpace(t *testing.T) {

	rand.Seed(time.Now().UnixNano())

	var tests = []struct {
		Arg1 int
		Arg2 int
	}{
		{0, 3},
		{3, 7},
		{7, 10},
		{-3, 3},
	}

	for _, v := range tests {
		r := int(RandSpace(int64(v.Arg1), int64(v.Arg2)))
		fmt.Println(r)
		assert.Equal(t, (r >= v.Arg1 && r <= v.Arg2), true)
	}

}
