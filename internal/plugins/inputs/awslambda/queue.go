// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2024-present Guance, Inc.

package awslambda

// CircularQueue represents a circular queue.
type CircularQueue[E any] struct {
	Data []E
	rear int
	len  int
}

// NewCircularQueue initializes a new circular queue.
func NewCircularQueue[S ~[]E, E any](arr S) *CircularQueue[E] {
	return &CircularQueue[E]{
		Data: arr,
		rear: 0,
		len:  0,
	}
}

// NewCircularQueueDefaultCap initializes a new circular queue.
func NewCircularQueueDefaultCap[E any]() *CircularQueue[E] {
	t := [5]E{}
	arr := t[:]
	return &CircularQueue[E]{
		Data: arr,
		rear: 0,
		len:  0,
	}
}

func (q *CircularQueue[T]) Len() int {
	return q.len
}

// Enqueue adds an element to the queue.
func (q *CircularQueue[T]) Enqueue(value T) T {
	t := q.Data[q.rear]
	q.Data[q.rear] = value
	q.rear = (q.rear + 1) % len(q.Data)
	if q.len < len(q.Data) {
		q.len++
	}
	return t
}

// ContainsFunc returns the first index i satisfying f(s[i]),
// or -1 if none do.
func (q *CircularQueue[T]) ContainsFunc(f func(T) bool) (T, bool) {
	for i := q.len; i > 0; i-- {
		idx := (q.rear - i) % len(q.Data)
		if idx < 0 {
			idx += len(q.Data)
		}
		if f(q.Data[idx]) {
			return q.Data[idx], true
		}
	}
	var zero T
	return zero, false
}

func (q *CircularQueue[T]) DeleteFunc(f func(T) bool) (T, bool) {
	var zero T
	if q.len == 0 {
		return zero, false
	}
	for i := q.len; i > 0; i-- {
		idx := (q.rear - i) % len(q.Data)
		if idx < 0 {
			idx += len(q.Data)
		}

		if f(q.Data[idx]) {
			res := q.Data[idx]

			for j := 0; j < q.len; j++ {
				q.Data[(idx+j)%len(q.Data)] = q.Data[(idx+j+1)%len(q.Data)]
			}
			q.rear = (q.rear - 1) % len(q.Data)
			if q.rear < 0 {
				q.rear += len(q.Data)
			}
			q.Data[q.rear] = zero
			q.len--
			return res, true
		}
	}

	return zero, false
}
