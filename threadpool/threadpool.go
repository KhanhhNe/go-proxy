package threadpool

import (
	"sync"
	"time"
)

type Thread interface {
	Id() string
	Run()
}

type Pool[T Thread] struct {
	Size int

	tasks      chan T
	taskIds    map[string]bool
	tasksQueue *Queue[T]
	newTasks   chan bool

	mu sync.Mutex
}

func NewThreadPool[T Thread](size int) *Pool[T] {
	p := &Pool[T]{
		0,
		make(chan T),
		map[string]bool{},
		NewQueue[T](),
		make(chan bool),
		sync.Mutex{},
	}

	p.Scale(size)
	go func() {
		var thread T
		var ok bool

		for {
			select {
			case <-p.newTasks:
			case <-time.After(time.Second):
			}

			for {
				thread, ok = p.tasksQueue.Pop()
				if ok {
					p.tasks <- thread
					continue
				}

				break
			}
		}
	}()

	return p
}

func (p *Pool[T]) AddTask(t T) {
	if _, already := p.taskIds[t.Id()]; already {
		return
	}

	p.mu.Lock()
	p.taskIds[t.Id()] = true
	p.tasksQueue.Push(t)

	// Avoid blocking, new tasks will be handled by the current running loop already
	select {
	case p.newTasks <- true:
	default:
	}

	p.mu.Unlock()
}

func (p *Pool[T]) Scale(size int) {
	for i := p.Size; i < size; i++ {
		// Scale workers
		go p.worker(i)
	}
	p.Size = size
}

func (p *Pool[T]) worker(id int) {
	var t T
	for {
		t = <-p.tasks
		if id >= p.Size {
			p.AddTask(t)
			return
		}

		p.mu.Lock()
		delete(p.taskIds, t.Id())
		p.mu.Unlock()
		t.Run()
	}
}
