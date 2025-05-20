package common

import (
	"encoding/json"
	"fmt"
	"os"
)

func DoMap(jobName string, mapTaskNumber int, inFile string, nReduce int, mapF func(string, string) []KeyValue) {
	// READ INPUT FILE
	data, err := os.ReadFile(inFile)
	if err != nil {
		panic(err)
	}

	kva := mapF(inFile, string(data))

	// PARTITION RESULTS
	buckets := make([][]KeyValue, nReduce)
	for _, kv := range kva {
		index := IHASH(kv.Key) % nReduce
		buckets[index] = append(buckets[index], kv)
	}

	// WRITE INTERMEDIATE FILES
	for i := 0; i < nReduce; i++ {
		fileName := fmt.Sprintf("mr-%d-%d", mapTaskNumber, i)
		file, err := os.Create(fileName)
		if err != nil {
			panic(err)
		}
		enc := json.NewEncoder(file)
		for _, kv := range buckets[i] {
			if err := enc.Encode(&kv); err != nil {
				panic(err)
			}
		}
		file.Close()
	}
}
