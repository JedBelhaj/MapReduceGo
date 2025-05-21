# MapReduce Distributed System

This project implements a mini MapReduce system in Go with:

- A master node coordinating tasks and tracking progress
- Multiple worker nodes performing map and reduce jobs
- A web dashboard showing live task progress and the final result

## 🛠️ Features

- Distributed task scheduling
- Fault tolerance via task timeout & reassignment
- Live dashboard at `http://localhost:8080`
- Final result appears in `mrtmp.wordcount` and is shown on the dashboard

## 📁 File Structure

```
.
├── mapreduce/
│   ├── master.go        # Master logic: task assignment, dashboard, result handling
│   ├── worker.go        # Worker logic: performs map/reduce functions
│   ├── common.go        # Shared types and utilities
│   ├── dashboard.html   # Frontend for the dashboard UI
│   └── ...              # Map/Reduce functions (e.g., word count)
├── main.go              # Entry point
└── mrtmp.wordcount      # Final output (auto-generated)
```

## ▶️ How to Run

1. **Start the Master**

   ```bash
   go run main.go -mode=master -nReduce=3 -nWorkers=2
   ```

   Starts a master with:

   - 3 reduce tasks
   - 2 worker goroutines

   Opens the dashboard at: `http://localhost:8080`

   Merges final result into `mrtmp.wordcount`

   ⚠️ **Note**: By default, input files are defined in `main.go` (e.g., `pg-*.txt`). Make sure those exist or edit them.

## 🌐 Web Dashboard

Once running, visit:

📍 `http://localhost:8080`

You'll see:

- Live task completion progress
- All map/reduce tasks and their current status
- Worker activity
- The final result printed below once all tasks complete

## 🧪 Example Output

```
word1  17
word2  5
word3  29
...
```

## 🧰 Requirements

- Go 1.18+
- OS: Linux, macOS, or Windows

## 📌 Notes

- Workers are simulated as goroutines in this setup.
- You can scale by implementing remote workers connecting to the master's RPC server (`:1234`).
- The result file is auto-merged into `mrtmp.wordcount` after the reduce phase.
