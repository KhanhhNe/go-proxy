package threadpool

type ThreadPool struct {
	Size int

	tasks      chan func()
	tasksQueue []func()
	newTasks   chan bool
}

func NewThreadPool(size int) *ThreadPool {
	p := &ThreadPool{
		0,
		make(chan func()),
		[]func(){},
		make(chan bool),
	}

	p.Scale(size)
	go func() {
		for {
			<-p.newTasks
			for _, t := range p.tasksQueue {
				p.tasks <- t
			}
		}
	}()

	return p
}

func (p *ThreadPool) AddTask(f func()) {
	p.tasksQueue = append(p.tasksQueue, f)

	// Avoid blocking, new tasks will be handled by the current running loop already
	select {
	case p.newTasks <- true:
	default:
	}
}

func (p *ThreadPool) Scale(size int) {
	for i := p.Size; i < size; i++ {
		// Scale workers
		go p.worker(i)
	}
	p.Size = size
}

func (p *ThreadPool) worker(id int) {
	for f := range p.tasks {
		if id >= p.Size {
			// This worker exceeded pool size
			p.tasks <- f
			return
		}

		f()
	}
}

var ServerPrecheckPool = NewThreadPool(50)
