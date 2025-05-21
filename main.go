package main

import (
	"flag"
	"fmt"
	"mr/mapreduce"
	"os"
	"strings"
)

func main() {
	// Define flags
	mode := flag.String("mode", "", "Mode to run: 'master' or 'worker'")
	files := "inputs/file1.txt,inputs/file2.txt"
	nReduce := flag.Int("nReduce", 3, "Number of reduce tasks")
	masterAddr := "localhost:1234"
	nWorkers := flag.Int("nWorkers", 1, "Number of workers to launch (only used in master mode)")
	jobName := "wordcount"

	// Parse flags
	flag.Parse()
	flag.VisitAll(func(f *flag.Flag) {
		fmt.Printf("Flag: -%s = %s\n", f.Name, f.Value)
	})

	// Validate mode
	if *mode != "master" && *mode != "worker" {
		fmt.Println("Invalid mode. Use -mode=master or -mode=worker")
		flag.Usage()
		os.Exit(1)
	}

	// Validate common flags
	if *nReduce <= 0 {
		fmt.Println("Error: -nReduce must be a positive integer")
		flag.Usage()
		os.Exit(1)
	}
	if jobName == "" {
		fmt.Println("Error: -job must not be empty")
		flag.Usage()
		os.Exit(1)
	}

	switch *mode {
	case "master":
		// Validate master-specific flags
		if files == "" {
			fmt.Println("Error: You must provide input files with -files")
			flag.Usage()
			os.Exit(1)
		}

		// Clean and validate input files
		inputFiles := strings.Split(files, ",")
		var cleanedFiles []string
		for _, f := range inputFiles {
			f = strings.TrimSpace(f)
			if f != "" {
				cleanedFiles = append(cleanedFiles, f)
			}
		}
		if len(cleanedFiles) == 0 {
			fmt.Println("Error: No valid input files provided")
			flag.Usage()
			os.Exit(1)
		}

		// Handle nWorkers
		if *nWorkers < 0 {
			fmt.Println("Error: -nWorkers cannot be negative")
			flag.Usage()
			os.Exit(1)
		}
		fmt.Printf("Starting %d worker(s) for master mode\n", *nWorkers)

		// If nWorkers > 0, start that many workers locally (in goroutines)
		if *nWorkers > 0 {
			go mapreduce.RunWorkers(masterAddr, *nWorkers, mapreduce.MapWordCount, mapreduce.ReduceWordCount)
		} else {
			fmt.Println("No workers started (nWorkers=0)")
		}

		// Start the master (coordinator) that runs the distributed MapReduce job
		mapreduce.StartDistributed(jobName, cleanedFiles, *nReduce, mapreduce.MapWordCount, mapreduce.ReduceWordCount)

	case "worker":
		// Workers don't use nWorkers; ignore it
		fmt.Println("Starting a single worker")
		mapreduce.RunWorkers(masterAddr, 1, mapreduce.MapWordCount, mapreduce.ReduceWordCount)
	}
}
