package threadpool

import "sync"

type QueueNode[T any] struct {
	emptyItem T
	items     []T
	l         int
	r         int
	next      *QueueNode[T]
}

const NODE_SIZE = 10

func NewQueueNode[T any]() *QueueNode[T] {
	return &QueueNode[T]{
		*new(T),
		make([]T, NODE_SIZE),
		-1,
		0,
		nil,
	}
}

func (n *QueueNode[T]) Push(t T) bool {
	if n.OutOfCapacity() {
		return false
	}

	n.items[n.r] = t
	n.r += 1
	return true
}

func (n *QueueNode[T]) OutOfCapacity() bool { return n.r >= len(n.items) }
func (n *QueueNode[T]) HasMore() bool       { return n.l+1 < n.r }

func (n *QueueNode[T]) Pop() (T, bool) {
	if !n.HasMore() {
		return n.emptyItem, false
	}

	n.l += 1
	return n.items[n.l], true
}

type Queue[T any] struct {
	emptyItem T
	head      *QueueNode[T]
	tail      *QueueNode[T]
	mu        sync.Mutex
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		*new(T),
		nil,
		nil,
		sync.Mutex{},
	}
}

func (q *Queue[T]) Push(t T) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.head == nil || q.tail == nil {
		// Queue is empty
		q.head = NewQueueNode[T]()
		q.tail = q.head
	}

	if q.tail.OutOfCapacity() {
		node := NewQueueNode[T]()
		q.tail.next = node
		q.tail = node
	}

	ok := q.tail.Push(t)
	if !ok {
		panic("Error pushing new item to queue")
	}
}

func (q *Queue[T]) Pop() (T, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for q.head != nil {
		if q.head.HasMore() {
			return q.head.Pop()
		}

		if q.head.OutOfCapacity() {
			// Head won't have any more items added since it's full already
			q.head = q.head.next
			if q.head == nil {
				q.tail = nil
			}
		} else {
			break
		}
	}

	return q.emptyItem, false
}
