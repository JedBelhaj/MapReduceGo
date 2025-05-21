package mapreduce

import (
	"fmt"
	"log"
	"math/rand"
	"net/rpc"
	"time"
)

// Worker represents a MapReduce worker
type Worker struct {
	id         string
	masterAddr string
	client     *rpc.Client
	mapF       func(string) []KeyValue
	reduceF    func(string, []string) string
}

func NewWorker(id string, masterAddr string,
	mapF func(string) []KeyValue,
	reduceF func(string, []string) string) *Worker {
	return &Worker{
		id:         id,
		masterAddr: masterAddr,
		mapF:       mapF,
		reduceF:    reduceF,
	}
}

// Start begins the worker's task execution loop
func (w *Worker) Start() {
	// Connect to the Master
	client, err := rpc.Dial("tcp", w.masterAddr)
	CheckError(err, "Failed to connect to Master at %s: %v\n", w.masterAddr, err)
	w.client = client

	for {
		// Request a task
		args := &TaskArgs{WorkerID: w.id}
		var reply TaskReply
		err := w.client.Call("Master.GetTask", args, &reply)
		CheckError(err, "Failed to call GetTask: %v\n", err)

		if !reply.Available {
			time.Sleep(1 * time.Second)
			continue
		}

		if w.simulateFailure() {
			log.Printf("Worker %s simulating crash for task %d\n", w.id, reply.Task.TaskID)
			return
		}
		if w.simulateDelay() {
			log.Printf("Worker %s simulating delay for task %d\n", w.id, reply.Task.TaskID)
			time.Sleep(5 * time.Second)
		}

		if reply.Task.Type == "map" {
			DoMap(reply.Task.JobName, reply.Task.MapNum, reply.Task.File, reply.Task.NReduce, w.mapF)
		} else if reply.Task.Type == "reduce" {
			DoReduce(reply.Task.JobName, reply.Task.ReduceNum, reply.Task.NMap, w.reduceF)
		}

		reportArgs := &ReportArgs{TaskID: reply.Task.TaskID, WorkerID: w.id}
		var reportReply struct{}
		err = w.client.Call("Master.ReportTaskDone", reportArgs, &reportReply)
		CheckError(err, "Failed to call ReportTaskDone: %v\n", err)
	}
}

func (w *Worker) simulateFailure() bool {
	return rand.Float64() < 0.1 // 10% chance of failure
}

func (w *Worker) simulateDelay() bool {
	return rand.Float64() < 0.2 // 20% chance of delay
}

// RunWorkers starts multiple workers concurrently
func RunWorkers(masterAddr string, numWorkers int,
	mapF func(string) []KeyValue,
	reduceF func(string, []string) string) {
	for i := 0; i < numWorkers; i++ {
		workerID := fmt.Sprintf("worker-%d", i)
		worker := NewWorker(workerID, masterAddr, mapF, reduceF)
		go worker.Start()
	}
}
