package goroutine_pool

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewGoPool(t *testing.T) {
	gp := NewGoPool(4, 8, minKeepAlive)

	for i := 0; i < 100; i++ {
		i2 := i
		gp.Execute(func() {
			// 模拟耗时操作
			time.Sleep(time.Second)
			fmt.Printf("hello world: %d\n", i2)
		})
	}

	time.Sleep(time.Minute * 5)
}

var runTimes = 1000000

func demoFunc() {
	time.Sleep(10 * time.Millisecond)
}

func BenchmarkGoPool_Execute(b *testing.B) {
	wg := sync.WaitGroup{}
	gp := NewGoPool(200000, 400000, minKeepAlive)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(runTimes)
		for j := 0; j < runTimes; j++ {
			gp.Execute(func() {
				demoFunc()
				wg.Done()
			})
		}
		wg.Wait()
	}
	b.StopTimer()
}
