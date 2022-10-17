package eventstream

type node[T any] struct {
	value T             // the value of the event (nil until ready)
	next  *node[T]      // the next node in the stream (nil until ready)
	ready chan struct{} // a channel whose closure marks readiness
}

func (n *node[T]) Ready() <-chan struct{} {
	return n.ready
}

func (n *node[T]) Next() (T, Promise[T]) {
	<-n.ready
	if n.next == nil {
		// force nil interface
		return n.value, nil
	}
	return n.value, n.next
}

func (n *node[T]) Iterator() Iterator[T] {
	return &iterator[T]{p: n}
}

func (n *node[T]) makeReady(v T, next *node[T]) {
	n.value = v
	n.next = next
	close(n.ready)
}

var _ Promise[any] = (*node[any])(nil)
