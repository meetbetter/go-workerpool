package workerpool

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

// 实现Job接口
type something struct {
	num int
}

func (s *something) Do() {
	fmt.Printf("my num is %d, I am doing my job\n", s.num)
	time.Sleep(2 * time.Second)
}

func TestWorkPool(t *testing.T) {

	jobNum := 100
	workerNum := 5

	wp := NewWorkerPool(workerNum)
	wp.Start()

	go func() {
		for {
			fmt.Println("NumGoroutine: ", runtime.NumGoroutine())
			time.Sleep(time.Second)
		}
	}()

	for i := 0; i < jobNum; i++ {
		err := wp.Add(&something{num: i})
		if err != nil {
			fmt.Println("error: ", err)
			return
		}

		// if i > 50 {
		// 	wp.Stop()
		// }
	}

}
