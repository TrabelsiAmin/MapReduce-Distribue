package mapreduce

import (
	"math/rand"
	"net/rpc"
	"os"
	"time"
)

// Worker execute map ou reduce
type Worker struct {
	id         string
	masterAddr string
}

// NewWorker initialise new worker
func NewWorker(id, masterAddr string) *Worker {
	rand.Seed(time.Now().UnixNano())
	return &Worker{
		id:         id,
		masterAddr: masterAddr,
	}
}

// Run starts the worker loop
func (w *Worker) Run() {
	for {
		// Request task
		client, err := rpc.DialHTTP("tcp", w.masterAddr)
		if err != nil {
			Debug("Worker %s: Cannot connect to master: %v\n", w.id, err)
			time.Sleep(time.Second)
			continue
		}
		var reply GetTaskReply
		err = client.Call("Master.GetTask", &GetTaskArgs{WorkerID: w.id}, &reply)
		client.Close()
		if err != nil {
			Debug("Worker %s: GetTask failed: %v\n", w.id, err)
			time.Sleep(time.Second)
			continue
		}

		if reply.Task.Type == IdleTask {
			time.Sleep(time.Second)
			continue
		}

		// Simulate crash (5%) or delay (10%)
		if rand.Float64() < 0.05 {
			Debug("Worker %s: Simulating crash for task %d\n", w.id, reply.Task.ID)
			os.Exit(1)
		}
		if rand.Float64() < 0.1 {
			Debug("Worker %s: Simulating delay for task %d\n", w.id, reply.Task.ID)
			time.Sleep(5 * time.Second)
		}

		// Execute task
		if reply.Task.Type == MapTask {
			Debug("Worker %s: Executing map task %d\n", w.id, reply.Task.ID)
			DoMap(reply.Task.JobName, reply.Task.MapTaskNumber, reply.Task.File, reply.Task.NReduce, MapWordCount)
		} else if reply.Task.Type == ReduceTask {
			Debug("Worker %s: Executing reduce task %d\n", w.id, reply.Task.ID)
			DoReduce(reply.Task.JobName, reply.Task.ReduceTaskNumber, reply.Task.NMap, ReduceWordCount)
		}

		// Wait for 3 seconds after task execution
		Debug("Worker %s: Resting for 3 seconds after task %d\n", w.id, reply.Task.ID)
		time.Sleep(3 * time.Second)

		// Report completion
		client, err = rpc.DialHTTP("tcp", w.masterAddr)
		if err != nil {
			Debug("Worker %s: Cannot report task %d: %v\n", w.id, reply.Task.ID, err)
			continue
		}
		var doneReply ReportTaskDoneReply
		err = client.Call("Master.ReportTaskDone", &ReportTaskDoneArgs{TaskID: reply.Task.ID, WorkerID: w.id}, &doneReply)
		client.Close()
		if err != nil {
			Debug("Worker %s: ReportTaskDone failed for task %d: %v\n", w.id, reply.Task.ID, err)
		}
	}
}
