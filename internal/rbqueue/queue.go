package rbqueue

import "sync"

type Queue[T any] struct {
	data []T
	sync.Mutex
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		data: make([]T, 0),
	}
}

func (q *Queue[T]) Append(v T) {
	q.data = append(q.data, v)
}

func (q *Queue[T]) Peek() (T, bool) {
	var zero T
	if q.IsEmpty() {
		return zero, false
	}
	return q.data[0], true
}

func (q *Queue[T]) Pop() (T, bool) {
	var zero T
	if q.IsEmpty() {
		return zero, false
	}

	v := q.data[0]
	q.data = q.data[1:]
	return v, true
}

func (q *Queue[T]) Len() int {
	return len(q.data)
}

func (q *Queue[T]) IsEmpty() bool {
	return len(q.data) == 0
}
