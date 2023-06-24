package utils

// containsInSlice 判断字符串是否在 slice 中
func ContainsInSlice(items []string, item string) bool {
	for _, eachItem := range items {
		if eachItem == item {
			return true
		}
	}
	return false
}
