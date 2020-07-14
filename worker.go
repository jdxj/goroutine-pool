package goroutine_pool

const (
	unknown = iota
	running
	shutdown
	stop
)

func newWorker(gp *GoPool, work func()) *Worker {
	w := &Worker{
		gp:      gp,
		curWork: work,
	}
	return w
}

type Worker struct {
	gp      *GoPool
	curWork func()
}

// Run 需要显式调用
func (w *Worker) Run() {
	var err error

	go func() {
		// 如果要退出, 说明要销毁该协程,
		// 同时将该 worker 从 pool 中删除.
		defer w.gp.delWorker(w)

		for w.curWork != nil {
			// 执行任务
			w.curWork()
			// 获取下一个任务
			w.curWork, err = w.gp.getWork()
			if err == ErrNoWork {
				w.gp = nil
				return
			}
		}
	}()
}
