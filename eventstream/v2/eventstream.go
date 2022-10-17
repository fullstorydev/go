package eventstream

import (
	"sync/atomic"
)

const defaultBufferSize = 1024

type eventStream[T any] struct {
	tail          *node[T]                // the current tail, as used only by the publisher
	subscribeTail atomic.Pointer[node[T]] // the current tail, as published to subscribers

	buffer    []node[T] // a buffer of nodes to use
	bufferPos int       // the next available node within buffer
}

var _ EventStream[any] = (*eventStream[any])(nil)

func (e *eventStream[T]) Publish(v T) {
	if e.buffer == nil {
		panic("closed")
	}
	pub := e.tail
	nextTail := e.initNextTail()
	pub.makeReady(v, nextTail)
}

func (e *eventStream[T]) Close() {
	e.buffer = nil
	e.bufferPos = 0
	var none T
	e.tail.makeReady(none, nil)
}

func (e *eventStream[T]) Subscribe() Promise[T] {
	return e.subscribeTail.Load()
}

func (e *eventStream[T]) initNextTail() *node[T] {
	if e.bufferPos >= len(e.buffer) {
		e.buffer = make([]node[T], len(e.buffer))
		e.bufferPos = 0
	}

	newTail := &e.buffer[e.bufferPos]
	e.bufferPos++
	newTail.ready = make(chan struct{})
	e.tail = newTail
	e.subscribeTail.Store(newTail)
	return newTail
}

// New creates an EventStream with the default buffer size.
func New[T any]() EventStream[T] {
	return NewWithBuffer[T](defaultBufferSize)
}

// NewWithBuffer creates an EventStream with the given buffer size.
func NewWithBuffer[T any](bufferSize int) EventStream[T] {
	if bufferSize < 1 {
		panic("invalid buffer size")
	}
	ret := &eventStream[T]{
		buffer:    make([]node[T], bufferSize),
		bufferPos: 0,
	}
	ret.initNextTail()
	return ret
}
