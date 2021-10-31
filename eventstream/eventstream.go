package eventstream

import (
	"sync/atomic"
)

const defaultBufferSize = 1024

type eventStream struct {
	tail          *node        // the current tail, as used only by the publisher
	subscribeTail atomic.Value // the current tail, as published to subscribers

	buffer    []node // a buffer of nodes to use
	bufferPos int    // the next available node within buffer
}

var _ EventStream = (*eventStream)(nil)

func (e *eventStream) Publish(v interface{}) {
	if e.buffer == nil {
		panic("closed")
	}
	pub := e.tail
	nextTail := e.initNextTail()
	pub.makeReady(v, nextTail)
}

func (e *eventStream) Close() {
	e.buffer = nil
	e.bufferPos = 0
	e.tail.makeReady(nil, nil)
}

func (e *eventStream) Subscribe() Promise {
	return e.subscribeTail.Load().(*node)
}

func (e *eventStream) initNextTail() *node {
	if e.bufferPos >= len(e.buffer) {
		e.buffer = make([]node, len(e.buffer))
		e.bufferPos = 0
	}

	newTail := &e.buffer[e.bufferPos]
	e.bufferPos++
	newTail.ready = make(chan struct{})
	e.tail = newTail
	e.subscribeTail.Store(newTail)
	return newTail
}

// NewEventStream creates an EventStream with the default buffer size.
func NewEventStream() EventStream {
	return NewEventStreamWithBuffer(defaultBufferSize)
}

// NewEventStreamWithBuffer creates an EventStream with the given buffer size.
func NewEventStreamWithBuffer(bufferSize int) EventStream {
	if bufferSize < 1 {
		panic("invalid buffer size")
	}
	ret := &eventStream{
		buffer:    make([]node, bufferSize),
		bufferPos: 0,
	}
	ret.initNextTail()
	return ret
}
