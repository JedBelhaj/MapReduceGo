package main

import (
	"mapreduce/common"
	"strconv"
	"strings"
)

func mapF(filename string, contents string) []common.KeyValue {
	words := strings.Fields(contents)
	kva := []common.KeyValue{}
	for _, word := range words {
		kva = append(kva, common.KeyValue{Key: word, Value: "1"})
	}
	return kva
}

func reduceF(key string, values []string) string {
	return strconv.Itoa(len(values)) // sum of occurrences
}

func main() {
	// MAP PHASE
	common.DoMap("job1", 0, "input.txt", 3, mapF)

	// REDUCE PHASE
	for i := 0; i < 3; i++ {
		common.DoReduce("job1", i, 1, reduceF)
	}
}
