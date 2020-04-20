package utility

import "math/rand"

// RandSpace 区间随机
func RandSpace(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}
