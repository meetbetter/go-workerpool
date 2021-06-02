package workerpool

import (
	"fmt"
	"sync"
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
		quit:    wp.quit,         //类似context cancel的思想使用channel close来通知goroutine退出
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
	quit        chan struct{}
	// once        sync.Once
	rwmutex sync.RWMutex
}

func NewWorkerPool(workerLen int) *workerpool {
	return &workerpool{
		workerLen:   workerLen,
		jobChan:     make(chan IJob),
		workerQueue: make(chan chan IJob, workerLen),
		quit:        make(chan struct{}),
		// rwmutex: sync.RWMutex{},
	}
}

func (wp *workerpool) Start() {
	for i := 0; i < wp.workerLen; i++ {
		worker := wp.newWorker()
		worker.run(wp.workerQueue)
	}

	go func() {
		for {
			select {
			case job := <-wp.jobChan: // 读取新写入的任务
				workerJobQueue := <-wp.workerQueue // 读取空闲的worker的jobQueue
				workerJobQueue <- job
			case <-wp.quit:
				return
			}
		}

	}()
}

// 将任务交给workerpool
func (wp *workerpool) Add(job IJob) error {
	// _, ok := <-wp.quit
	// if !ok {
	// 	return fmt.Errorf("workerpool has been closed")
	// }

	// wp.rwmutex.RLock()
	// defer wp.rwmutex.RUnlock()

	// if _, ok := <-wp.quit; !ok {
	// 	return fmt.Errorf("workerpool has been closed")
	// }

	wp.jobChan <- job

	return nil
}

// 关闭workerpool
func (wp *workerpool) Stop() {
	// wp.once.Do(func() { // 单例防止重复关闭channel导致panic
	// 	close(wp.quit)
	// })
	// wp.once.Do(func() { // 单例防止重复关闭channel导致panic
	// 	close(wp.jobChan)
	// })
	// wp.rwmutex.Lock()
	// defer wp.rwmutex.Unlock()
	if _, ok := <-wp.quit; ok {
		close(wp.quit)
	}
}
