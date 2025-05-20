package common

type KeyValue struct {
	Key   string
	Value string
}

// Hash function for partitioning
func IHASH(key string) int {
	hash := 0
	for _, c := range key {
		hash = int(c) + (hash << 6) + (hash << 16) - hash
	}
	return hash & 0x7fffffff
}
