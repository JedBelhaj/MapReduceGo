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

func DoReduce(
	jobName string,
	reduceTaskNumber int,
	nMap int,
	reduceF func(string, []string) string,
) {
	// key -> []string
	keyValues := make(map[string][]string)

	// 1. READ all intermediate files from each map task
	for i := 0; i < nMap; i++ {
		fileName := fmt.Sprintf("mr-%d-%d", i, reduceTaskNumber)
		file, err := os.Open(fileName)
		if err != nil {
			panic(err)
		}
		dec := json.NewDecoder(file)
		for {
			var kv KeyValue
			if err := dec.Decode(&kv); err != nil {
				break // EOF
			}
			keyValues[kv.Key] = append(keyValues[kv.Key], kv.Value)
		}
		file.Close()
	}

	// 2. WRITE final output
	outFileName := fmt.Sprintf("mr-out-%d", reduceTaskNumber)
	outFile, err := os.Create(outFileName)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	for key, values := range keyValues {
		result := reduceF(key, values)
		fmt.Fprintf(outFile, "%v %v\n", key, result)
	}
}
