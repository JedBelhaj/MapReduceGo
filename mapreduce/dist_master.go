package mapreduce

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

// Task represents a map or reduce task
type Task struct {
	Type      string    `json:"Type"`      // "map" or "reduce"
	File      string    `json:"File"`      // Input file for map tasks or identifier for reduce
	Status    string    `json:"Status"`    // "pending", "in-progress", "completed"
	Worker    string    `json:"Worker"`    // Worker assigned to the task
	StartTime time.Time `json:"-"`         // Task start time (ignore in JSON)
	TaskID    int       `json:"TaskID"`    // Unique task ID
	MapNum    int       `json:"MapNum"`    // Map task index
	ReduceNum int       `json:"ReduceNum"` // Reduce task index
	NMap      int       `json:"NMap"`      // Total map tasks (for reduce tasks)
	NReduce   int       `json:"NReduce"`   // Total reduce tasks (for map tasks)
	JobName   string    `json:"JobName"`   // Job name for context
}

// Master holds the MapReduce job state
type Master struct {
	mu          sync.Mutex
	tasks       []Task
	workers     map[string]string // workerID -> status ("Idle", "Working")
	nMap        int
	nReduce     int
	completed   int
	totalTasks  int
	jobName     string
	inputFiles  []string
	taskTimeout time.Duration
	mapF        func(string) []KeyValue
	reduceF     func(string, []string) string
}

// RPC argument/reply types
type TaskArgs struct {
	WorkerID string
}

type TaskReply struct {
	Task      Task
	Available bool
}

type ReportArgs struct {
	TaskID   int
	WorkerID string
}

// GetTask RPC handler for workers to get a task
func (m *Master) GetTask(args *TaskArgs, reply *TaskReply) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	// Reassign timed-out tasks first
	for i, task := range m.tasks {
		if task.Status == "in-progress" && now.Sub(task.StartTime) > m.taskTimeout {
			log.Printf("Task %d timed out, reassigning\n", task.TaskID)
			task.Status = "pending"
			task.Worker = ""
			m.tasks[i] = task
		}
	}

	// Assign a pending task to the worker
	for i, task := range m.tasks {
		if task.Status == "pending" {
			// For reduce tasks, ensure all map tasks are completed
			if task.Type == "reduce" {
				allMapsDone := true
				for j := 0; j < m.nMap; j++ {
					if m.tasks[j].Status != "completed" {
						allMapsDone = false
						break
					}
				}
				if !allMapsDone {
					continue // skip reduce tasks until map tasks done
				}
			}
			task.Status = "in-progress"
			task.Worker = args.WorkerID
			task.StartTime = now
			m.tasks[i] = task
			reply.Task = task
			reply.Available = true
			m.workers[args.WorkerID] = "Working"
			return nil
		}
	}

	// No tasks available
	reply.Available = false
	return nil
}

// ReportTaskDone marks a task as completed by a worker
func (m *Master) ReportTaskDone(args *ReportArgs, reply *struct{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, task := range m.tasks {
		if task.TaskID == args.TaskID && task.Worker == args.WorkerID && task.Status == "in-progress" {
			task.Status = "completed"
			m.tasks[i] = task
			m.completed++
			m.workers[args.WorkerID] = "Idle"
			log.Printf("Task %d completed by worker %s\n", task.TaskID, args.WorkerID)
			break
		}
	}
	return nil
}

// WorkerInfo is the JSON structure for workers in the dashboard data
type WorkerInfo struct {
	Name   string `json:"Name"`
	Status string `json:"Status"`
}

// DashboardData is the JSON response sent to the dashboard frontend
type DashboardData struct {
	Workers  []WorkerInfo `json:"Workers"`
	Tasks    []Task       `json:"Tasks"`
	Progress float64      `json:"Progress"`
}

// StartDashboard starts the HTTP server serving the dashboard UI and data
func (m *Master) StartDashboard() {
	http.HandleFunc("/", m.dashboardHandler)
	http.HandleFunc("/data", m.dataHandler)

	log.Println("Dashboard server started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Dashboard failed:", err)
	}
}

// dashboardHandler serves the dashboard HTML page
func (m *Master) dashboardHandler(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	jobDone := m.completed >= m.totalTasks
	m.mu.Unlock()

	var result string
	if jobDone {
		data, err := os.ReadFile(AnsName(m.jobName)) // usually mrtmp.wordcount
		if err != nil {
			result = "Error reading result file."
		} else {
			result = string(data)
		}
	}

	tmpl, err := template.ParseFiles("mapreduce/dashboard.html")
	if err != nil {
		http.Error(w, "Internal server error - template not found", http.StatusInternalServerError)
		return
	}

	data := struct {
		FinalResult string
	}{
		FinalResult: result,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Internal server error - template execution failed", http.StatusInternalServerError)
	}
}

// dataHandler serves the live JSON data for the dashboard
func (m *Master) dataHandler(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	data := DashboardData{
		Workers:  make([]WorkerInfo, 0, len(m.workers)),
		Tasks:    m.tasks,
		Progress: 0,
	}

	if m.totalTasks > 0 {
		data.Progress = float64(m.completed) / float64(m.totalTasks) * 100
	}

	for w, s := range m.workers {
		data.Workers = append(data.Workers, WorkerInfo{Name: w, Status: s})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// StartDistributed runs the distributed MapReduce master server
func StartDistributed(jobName string, files []string, nReduce int,
	mapF func(string) []KeyValue, reduceF func(string, []string) string) {

	m := &Master{
		tasks:       make([]Task, 0),
		workers:     make(map[string]string),
		nMap:        len(files),
		nReduce:     nReduce,
		completed:   0,
		totalTasks:  len(files) + nReduce,
		jobName:     jobName,
		inputFiles:  files,
		taskTimeout: 10 * time.Second,
		mapF:        mapF,
		reduceF:     reduceF,
	}

	// Create map tasks
	for i, file := range files {
		m.tasks = append(m.tasks, Task{
			Type:    "map",
			File:    file,
			Status:  "pending",
			TaskID:  i,
			MapNum:  i,
			NReduce: nReduce,
			JobName: jobName,
		})
	}

	// Create reduce tasks
	for i := 0; i < nReduce; i++ {
		m.tasks = append(m.tasks, Task{
			Type:      "reduce",
			File:      fmt.Sprintf("reduce-%d", i),
			Status:    "pending",
			TaskID:    len(files) + i,
			ReduceNum: i,
			NMap:      len(files),
			JobName:   jobName,
		})
	}

	// Start RPC server
	rpcServer := rpc.NewServer()
	err := rpcServer.Register(m)
	CheckError(err, "RPC registration failed: %v\n", err)
	listener, err := net.Listen("tcp", ":1234")
	CheckError(err, "RPC listen failed: %v\n", err)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("RPC accept error: %v\n", err)
				continue
			}
			go rpcServer.ServeConn(conn)
		}
	}()

	// Start dashboard HTTP server in goroutine
	go m.StartDashboard()

	// Wait for all tasks to complete
	for m.completed < m.totalTasks {
		time.Sleep(time.Second)
	}

	// Merge reduce output files
	resFiles := make([]string, nReduce)
	for i := 0; i < nReduce; i++ {
		resFiles[i] = MergeName(jobName, i)
	}
	err = concatFiles(AnsName(jobName), resFiles)
	CheckError(err, "Failed to merge results: %v\n", err)
	select {}
}
