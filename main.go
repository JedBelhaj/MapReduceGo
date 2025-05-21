package main

import (
	"flag"
	"fmt"
	"mr/mapreduce"
	"os"
	"strings"
)

func main() {
	mode := flag.String("mode", "", "Mode to run: 'master' or 'worker'")
	files := flag.String("files", "", "Comma-separated input files (used by master)")
	nReduce := flag.Int("nReduce", 3, "Number of reduce tasks")
	masterAddr := flag.String("master", "localhost:1234", "Address of the master (used by workers)")
	numWorkers := flag.Int("nWorkers", 1, "Number of workers to launch (only used in master mode)")
	jobName := flag.String("job", "wordcount", "Job name")

	flag.Parse()

	switch *mode {
	case "master":
		if *files == "" {
			fmt.Println("You must provide input files with -files")
			os.Exit(1)
		}
		inputFiles := strings.Split(*files, ",")

		// Start the master (coordinator) that runs the distributed MapReduce job
		mapreduce.StartDistributed(*jobName, inputFiles, *nReduce, mapreduce.MapWordCount, mapreduce.ReduceWordCount)

	case "worker":
		// This starts a worker that connects to the given master address
		mapreduce.RunWorkers(*masterAddr, *numWorkers, mapreduce.MapWordCount, mapreduce.ReduceWordCount)

		// Keep the worker running indefinitely
		select {}

	default:
		fmt.Println("Invalid mode. Use -mode=master or -mode=worker")
		os.Exit(1)
	}
}
