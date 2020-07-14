package goroutine_pool

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

var (
	ErrInvalidWork = fmt.Errorf("invaild work")
	ErrNoWork      = fmt.Errorf("not found work")
)

var (
	minGoAmount   = runtime.NumCPU()
	minKeepAlive  = 3 * time.Second
	maxWorksCache = 20
)

func NewGoPool(min, max int, keep time.Duration) *GoPool {
	if min < 0 {
		min = minGoAmount
	}
	if max < min {
		max = min
	}
	if keep < minKeepAlive {
		keep = minKeepAlive
	}

	gp := &GoPool{
		minAmount:    min,
		maxAmount:    max,
		mutexWorkers: &sync.Mutex{},
		workers:      make(map[*Worker]struct{}),

		keepAlive: minKeepAlive,
		works:     make(chan func(), maxWorksCache),
	}
	return gp
}

type GoPool struct {
	minAmount    int // 最小数量
	maxAmount    int // 最大数量
	mutexWorkers *sync.Mutex
	workers      map[*Worker]struct{} // 每个 worker 代表一个协程

	keepAlive time.Duration // 超过该时间后, 回收协程
	works     chan func()   // 任务缓冲区
}

func (gp *GoPool) Execute(work func()) {
	if work == nil {
		panic(ErrInvalidWork)
	}

	// 如果没到达最小数量要求, 则直接创建新 goroutine 执行任务
	if gp.getWorkersAmount() < gp.minAmount {
		worker := newWorker(gp, work)
		worker.Run()
		gp.addWorker(worker)
		return
	}

	// 如果到达最小数量要求, 那么尝试向 work 队列中放入任务
	if gp.getWorksAmount() < maxWorksCache {
		gp.works <- work
		return
	}

	// 如果队列也满了, 那么尝试创建新的协程
	if gp.getWorkersAmount() < gp.maxAmount {
		worker := newWorker(gp, work)
		worker.Run()
		gp.addWorker(worker)
		return
	}

	// 如果达到最大协程数了, 那么只能等待队列有空闲空间了
	gp.works <- work
}

func (gp *GoPool) getWorkersAmount() (amount int) {
	gp.mutexWorkers.Lock()
	amount = len(gp.workers)
	gp.mutexWorkers.Unlock()
	return
}

func (gp *GoPool) getWork() (func(), error) {
	timer := time.NewTimer(gp.keepAlive)
	defer timer.Stop()

	for {
		select {
		case work := <-gp.works:
			return work, nil
		case <-timer.C:
		}

		// 如果获取不到任务, 且当前有很多的 worker,
		// 那么可以销毁一个 worker.
		if gp.getWorkersAmount() > gp.minAmount {
			return nil, ErrNoWork
		}
		// 否则一直阻塞, 等待任务
		timer.Reset(gp.keepAlive)
	}
}

// delWorker 删除已缓存的 worker (释放协程, worker.Run() 中的 goroutine 返回)
func (gp *GoPool) delWorker(worker *Worker) {
	gp.mutexWorkers.Lock()
	delete(gp.workers, worker)
	gp.mutexWorkers.Unlock()
}

// addWorker 缓存新的 worker (创建协程, 显式调用 worker.Run())
func (gp *GoPool) addWorker(worker *Worker) {
	gp.mutexWorkers.Lock()
	gp.workers[worker] = struct{}{}
	gp.mutexWorkers.Unlock()
}

// getWorksAmount 获取已缓存的任务数量
func (gp *GoPool) getWorksAmount() int {
	return len(gp.works)
}
