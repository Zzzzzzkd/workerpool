package workerpool

import (
	"runtime"
	"testing"
	"time"
)

func TestWorkerPool(t *testing.T) {
	var goroutineCount = runtime.NumGoroutine()
	var x int = 0
	wp := NewPool("test", 20, func(i interface{}) {
		x++
	})
	wp.Start()
	incrNum := runtime.NumGoroutine() - goroutineCount
	if incrNum != 21 {
		t.Fatalf("incr num goroutine is %v not equal to 21", incrNum)
	}
	wp.NewTask(1, time.Second)
	time.Sleep(time.Second)
	if x != 1 {
		t.Fatalf("x should be 1 now is %v", x)
	}
	wp.NewTask(1, time.Second)
	time.Sleep(time.Second)
	if x != 2 {
		t.Fatalf("x should be 2 now is %v", x)
	}
	wp.Stop()
	//runtime.Gosched()
	time.Sleep(time.Second)

	if runtime.NumGoroutine() != goroutineCount {
		t.Fatalf("goroutine count should be %v but now is %v", goroutineCount, runtime.NumGoroutine())
	}
	if len(workerPoolReg.reg) != 0 {
		t.Fatalf("del from reg failed")
	}

}
