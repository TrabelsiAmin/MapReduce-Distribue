package mapreduce

// Sequential runs map and reduce tasks sequentially, waiting for each task to
// complete before scheduling the next.
import (
	"encoding/json"
	"net"
	"net/http"
	"net/rpc"
	"path/filepath"
	"sync"
	"time"
)

// TaskType definie le type de tâche
type TaskType string

const (
	MapTask    TaskType = "map"
	ReduceTask TaskType = "reduce"
	IdleTask   TaskType = "idle"
)

// Task représente une tâche de map ou de reduce
type Task struct {
	ID               int
	Type             TaskType
	JobName          string
	File             string // For map tasks
	MapTaskNumber    int
	ReduceTaskNumber int
	NReduce          int
	NMap             int
	Status           string // "pending", "running", "completed"
	WorkerID         string
	StartTime        time.Time
}

// WorkerInfo tracks worker status
type WorkerInfo struct {
	ID      string
	Status  string // "idle", "working", "crashed"
	Address string
}

// Master gere les tasks et les workers
type Master struct {
	tasks      []Task
	workers    map[string]*WorkerInfo
	nReduce    int
	jobName    string
	files      []string
	mu         sync.Mutex
	done       chan bool
	tasksDone  int
	totalTasks int
}

// NewMaster initializes a new master
func NewMaster(jobName string, files []string, nReduce int) *Master {
	m := &Master{
		tasks:     make([]Task, 0),
		workers:   make(map[string]*WorkerInfo),
		nReduce:   nReduce,
		jobName:   jobName,
		files:     files,
		done:      make(chan bool),
		tasksDone: 0,
	}

	// Initialize map tasks
	for i, file := range files {
		m.tasks = append(m.tasks, Task{
			ID:            i,
			Type:          MapTask,
			JobName:       jobName,
			File:          file,
			MapTaskNumber: i,
			NReduce:       nReduce,
			NMap:          len(files),
			Status:        "pending",
		})
	}

	// Initialize reduce tasks
	for i := 0; i < nReduce; i++ {
		m.tasks = append(m.tasks, Task{
			ID:               len(files) + i,
			Type:             ReduceTask,
			JobName:          jobName,
			ReduceTaskNumber: i,
			NReduce:          nReduce,
			NMap:             len(files),
			Status:           "pending",
		})
	}

	m.totalTasks = len(m.tasks)
	return m
}

// GetTask assigne task à un worker
type GetTaskArgs struct {
	WorkerID string
}

type GetTaskReply struct {
	Task Task
}

func (m *Master) GetTask(args *GetTaskArgs, reply *GetTaskReply) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Register worker if new
	if _, exists := m.workers[args.WorkerID]; !exists {
		m.workers[args.WorkerID] = &WorkerInfo{
			ID:      args.WorkerID,
			Status:  "idle",
			Address: args.WorkerID, // Simplified
		}
	}

	// Find a pending or timed-out task
	now := time.Now()
	for i, task := range m.tasks {
		if task.Status == "pending" || (task.Status == "running" && now.Sub(task.StartTime) > 10*time.Second) {
			m.tasks[i].Status = "running"
			m.tasks[i].WorkerID = args.WorkerID
			m.tasks[i].StartTime = now
			reply.Task = task
			m.workers[args.WorkerID].Status = "working"
			Debug("Master: Assigned task %d (%s) to worker %s\n", task.ID, task.Type, args.WorkerID)
			return nil
		}
	}
	// No tasks available
	reply.Task = Task{Type: IdleTask}
	m.workers[args.WorkerID].Status = "idle"
	Debug("Master: No tasks for worker %s, assigned idle\n", args.WorkerID)
	return nil
}

// ReportTaskDone marks a task as completed
type ReportTaskDoneArgs struct {
	TaskID   int
	WorkerID string
}

type ReportTaskDoneReply struct{}

// ReportTaskDone updates the status of a task to completed
// and notifies the master if all tasks are done
func (m *Master) ReportTaskDone(args *ReportTaskDoneArgs, reply *ReportTaskDoneReply) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, task := range m.tasks {
		if task.ID == args.TaskID && task.Status == "running" && task.WorkerID == args.WorkerID {
			m.tasks[i].Status = "completed"
			m.tasksDone++
			m.workers[args.WorkerID].Status = "idle"
			Debug("Master: Task %d completed by worker %s, %d/%d done\n", task.ID, args.WorkerID, m.tasksDone, m.totalTasks)
			if m.tasksDone == m.totalTasks {
				m.done <- true
			}
			break
		}
	}
	return nil
}

// CheckError checks for errors and panics if any
func (m *Master) startRPC() {
	rpc.Register(m)
	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", ":1234")
	CheckError(err, "cannot start RPC server: %v\n", err)
	go http.Serve(listener, nil)
}

// startHTTP starts the HTTP server for monitoring
func (m *Master) startHTTP() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("web", "index.html"))
	})
	http.HandleFunc("/data", m.serveData)
	go http.ListenAndServe(":8080", nil)
}

// serveData serves the current state of the master
// including tasks, workers, and task completion status
func (m *Master) serveData(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	type Data struct {
		Tasks      []Task       `json:"tasks"`
		Workers    []WorkerInfo `json:"workers"`
		TasksDone  int          `json:"tasksDone"`
		TotalTasks int          `json:"totalTasks"`
	}

	data := Data{
		Tasks:      m.tasks,
		Workers:    make([]WorkerInfo, 0, len(m.workers)),
		TasksDone:  m.tasksDone,
		TotalTasks: m.totalTasks,
	}
	for _, worker := range m.workers {
		data.Workers = append(data.Workers, *worker)
	}
	json.NewEncoder(w).Encode(data)
}

// Run starts the master
func (m *Master) Run() {
	Debug("Master: Starting RPC and HTTP servers\n")
	m.startRPC()
	m.startHTTP()
	<-m.done
	resFiles := make([]string, 0, m.nReduce)
	for i := 0; i < m.nReduce; i++ {
		resFiles = append(resFiles, MergeName(m.jobName, i))
	}
	err := concatFiles(AnsName(m.jobName), resFiles)
	CheckError(err, "cannot merge output files: %v\n")
	CleanIntermediary(m.jobName, len(m.files), m.nReduce)
	Debug("Master: Job %s completed\n", m.jobName)
	Debug("Master: Keeping HTTP server alive for 30 seconds\n")
	time.Sleep(30 * time.Second)
}
