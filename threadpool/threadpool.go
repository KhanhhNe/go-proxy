package threadpool

import (
	"sync"
	"time"
)

type FuncQueueNode struct {
	funcs []func()
	l     int
	r     int
	next  *FuncQueueNode
}

func NewFuncQueueNode() *FuncQueueNode {
	return &FuncQueueNode{
		make([]func(), 2),
		-1,
		0,
		nil,
	}
}

func (n *FuncQueueNode) Push(f func()) bool {
	if n.OutOfCapacity() {
		return false
	}

	n.funcs[n.r] = f
	n.r += 1
	return true
}

func (n *FuncQueueNode) OutOfCapacity() bool { return n.r >= len(n.funcs) }

func (n *FuncQueueNode) Pop() (*func(), bool) {
	if n.l+1 >= n.r {
		return nil, false
	}

	n.l += 1
	return &n.funcs[n.l], true
}

type FuncQueue struct {
	head *FuncQueueNode
	tail *FuncQueueNode
	mu   sync.Mutex
}

func NewFuncQueue() *FuncQueue {
	return &FuncQueue{
		nil,
		nil,
		sync.Mutex{},
	}
}

func (q *FuncQueue) Push(f func()) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.tail == nil {
		// Queue is empty
		q.head = NewFuncQueueNode()
		q.tail = q.head
	}

	// Try to push to existing tail
	ok := q.tail.Push(f)
	if !ok {
		// Tail is out of capacity
		node := NewFuncQueueNode()
		node.Push(f)

		q.tail.next = node
		q.tail = node
	}
}

func (q *FuncQueue) Pop() (*func(), bool) {
	q.mu.Lock()

	if q.head == nil {
		q.mu.Unlock()
		return nil, false
	}

	f, ok := q.head.Pop()
	if !ok {
		if q.head.OutOfCapacity() {
			q.head = q.head.next
			q.mu.Unlock()
			return q.Pop()
		}

		q.mu.Unlock()
		return nil, false
	}

	q.mu.Unlock()
	return f, true
}

type ThreadPool struct {
	Size int

	tasks      chan func()
	tasksQueue *FuncQueue
	newTasks   chan bool
}

func NewThreadPool(size int) *ThreadPool {
	p := &ThreadPool{
		0,
		make(chan func()),
		NewFuncQueue(),
		make(chan bool),
	}

	p.Scale(size)
	go func() {
		var f *func()
		var ok bool

		for {
			select {
			case <-p.newTasks:
			case <-time.After(time.Second):
			}

			for {
				f, ok = p.tasksQueue.Pop()
				if ok {
					p.tasks <- *f
					continue
				}

				break
			}
		}
	}()

	return p
}

func (p *ThreadPool) AddTask(f func()) {
	p.tasksQueue.Push(f)

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
	var f func()
	for {
		f = <-p.tasks
		if id >= p.Size {
			p.AddTask(f)
			return
		}

		f()
	}
}

var ServerPrecheckPool = NewThreadPool(50)
