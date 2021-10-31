package eventstream

type node struct {
	value interface{}   // the value of the event (nil until ready)
	next  *node         // the next node in the stream (nil until ready)
	ready chan struct{} // a channel whose closure marks readiness
}

func (n *node) Ready() <-chan struct{} {
	return n.ready
}

func (n *node) Next() (interface{}, Promise) {
	<-n.ready
	if n.next == nil {
		// force nil interface
		return n.value, nil
	}
	return n.value, n.next
}

func (n *node) Iterator() Iterator {
	return &iterator{p: n}
}

func (n *node) makeReady(v interface{}, next *node) {
	n.value = v
	n.next = next
	close(n.ready)
}

var _ Promise = (*node)(nil)
