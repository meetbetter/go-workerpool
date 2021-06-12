package workerpool

import (
	"fmt"
	"sync"
	"sync/atomic"
)

const (
	STATUS_INIT = iota
	STATUS_RUNNING
	STATUS_STOP
)

// 任务需要实现此接口
type IJob interface {
	Do()
}

type worker struct {
	jobChan chan IJob
	quit    chan struct{}
}

func (wp *workerpool) newWorker() worker {
	return worker{
		jobChan: make(chan IJob), // 可以设置有缓冲
		quit:    wp.quit,
	}
}

func (w worker) run(workerQueue chan chan IJob) {
	go func() {
		for {
			workerQueue <- w.jobChan //将当前的空闲worker放入workerpool的workerqueue

			select {
			case job := <-w.jobChan:
				job.Do()
			case <-w.quit:
				fmt.Println("worker return")
				return
			}
		}
	}()
}

type workerpool struct {
	workerLen   int
	jobChan     chan IJob
	workerQueue chan chan IJob
	isClosed    int64
	quit        chan struct{}
	rwmutex     sync.RWMutex
}

func NewWorkerPool(workerLen int) *workerpool {
	return &workerpool{
		workerLen:   workerLen,
		jobChan:     make(chan IJob),
		workerQueue: make(chan chan IJob, workerLen),
		isClosed:    STATUS_INIT,
		quit:        make(chan struct{}),
	}
}

func (wp *workerpool) Start() {
	wp.rwmutex.Lock()
	defer wp.rwmutex.Unlock()
	if atomic.LoadInt64(&wp.isClosed) != STATUS_RUNNING {
		atomic.StoreInt64(&wp.isClosed, STATUS_RUNNING)
	}

	for i := 0; i < wp.workerLen; i++ {
		worker := wp.newWorker()
		worker.run(wp.workerQueue)
	}

	go func() {
		for {
			if atomic.LoadInt64(&wp.isClosed) == STATUS_STOP {
				return
			}
			job := <-wp.jobChan                // 读取新写入的任务
			workerJobQueue := <-wp.workerQueue // 读取空闲的worker的jobQueue
			workerJobQueue <- job
		}

	}()
}

// 将任务交给workerpool
func (wp *workerpool) Add(job IJob) error {
	wp.rwmutex.RLock()
	defer wp.rwmutex.RUnlock()

	if atomic.LoadInt64(&wp.isClosed) != STATUS_RUNNING {
		return fmt.Errorf("workerpool is not running")
	}

	wp.jobChan <- job

	return nil
}

// 关闭workerpool
func (wp *workerpool) Stop() {
	wp.rwmutex.Lock()
	defer wp.rwmutex.Unlock()
	if atomic.LoadInt64(&wp.isClosed) != STATUS_STOP {
		atomic.StoreInt64(&wp.isClosed, STATUS_STOP)
		close(wp.quit)
	}
}
