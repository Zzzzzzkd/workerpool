package workerpool

import (
	"fmt"
	"sync"
	"time"
)

type WorkerPoolReg struct {
	reg map[string]workerPool
	m   sync.RWMutex
}

var workerPoolReg WorkerPoolReg

type workerPool struct {
	stopCh      chan bool
	workerQueue chan chan interface{}
	taskQueue   chan interface{}
	workers     []worker
	Name        string
}

type worker struct {
	taskQueue   chan interface{}
	stopCh      chan bool
	handler     func(interface{})
	workerQueue chan chan interface{}
}

func init() {
	workerPoolReg = WorkerPoolReg{
		reg: map[string]workerPool{},
	}
}

func delWorkerPool(name string) bool {
	workerPoolReg.m.Lock()
	defer workerPoolReg.m.Unlock()
	if reg, ok := workerPoolReg.reg[name]; ok {
		delete(workerPoolReg.reg, name)
		close(reg.stopCh)
	}
	return true
}

// create a new workerPool by specified worker number and task handler and name
func NewPool(name string, workersCount int, f func(interface{})) workerPool {
	workerPoolReg.m.Lock()
	defer workerPoolReg.m.Unlock()
	if wp, ok := workerPoolReg.reg[name]; ok {
		return wp
	}
	wp := workerPool{
		stopCh:      make(chan bool),
		workerQueue: make(chan chan interface{}, workersCount),
		workers:     make([]worker, workersCount),
		taskQueue:   make(chan interface{}),
		Name:        name,
	}
	workerPoolReg.reg[name] = wp
	for i := 0; i < workersCount; i++ {
		wp.workers[i] = worker{handler: f, taskQueue: make(chan interface{}), workerQueue: wp.workerQueue, stopCh: wp.stopCh}
	}
	return wp
}

// start all workers and schduler
func (wp workerPool) Start() {
	go wp.schedule()
	for i := range wp.workers {
		go wp.workers[i].run()
	}
}

// add a task to be handle
// time out avoid goroutine leak
func (wp workerPool) NewTask(task interface{}, t time.Duration) error {
	select {
	case wp.taskQueue <- task:
		return nil
	case <-time.After(t):
		return fmt.Errorf("NewTask time out")
	}
}

// stop a workerPool and delete from reg
func (wp workerPool) Stop() {
	delWorkerPool(wp.Name)
}

func (wp workerPool) schedule() {
	for {
		select {
		case task := <-wp.taskQueue:
			wq := <-wp.workerQueue
			wq <- task
		case <-wp.stopCh:
			return
		}
	}
}

func (w worker) run() {
	defer func() {
		if perr := recover(); perr != nil {
			go w.run()
		}
	}()
	for {
		w.workerQueue <- w.taskQueue
		select {
		case <-w.stopCh:
			return
		case task := <-w.taskQueue:
			w.handler(task)
		}
	}
}
