// Package eventstream implements a single-producer, multiple-consumer event stream.
package eventstream

import (
	"context"
	"errors"
)

// EventStream allows a single producer to publish events to multiple asynchronous consumers, who each
// receive all events. Note that event values are opaque to the event stream and no copies are made:
// consumers should not mutate consumed event values!
type EventStream interface {
	// Publish adds the next value to the stream. This method can be called concurrently with subscriber reads,
	// but not with other publisher operations.  External synchronization is required if there are multiple concurrent
	// publishers.
	Publish(interface{})

	// Close ends the stream; the same concurrency rules apply to Close() and Publish().
	Close()

	// Subscribe returns a Promise to the next unpublished event. The returned Promise gives the caller
	// the events in the stream from the current position forward.
	Subscribe() Promise
}

// Promise is a handle to the next event in the stream, plus all events following.
// A Promise is effectively immutable and can be shared. To be used concurrently with Publish operations.
type Promise interface {
	// Ready returns the ready channel for this node; the channel closes when this Promise is ready.
	Ready() <-chan struct{}

	// Next returns the next event in the stream, and the next Promise.  Multiple calls return consistent results.
	// Returns (nil, nil) when the stream is Closed.
	//
	// Note that this method internally blocks until the Ready() channel is closed!
	// Typical callers will not call Next() until this Promise is ready.
	Next() (interface{}, Promise)

	// Iterator creates an Iterator based on this Promise.  The Promise is unchanged.
	Iterator() Iterator
}

// ErrDone is returned by Iterator.Next() when the underlying EventStream is closed.
var ErrDone = errors.New("no more items in iterator")

// Iterator iterates an event stream.  To be used concurrently with Publish operations.
//
// Unlike Promises, Iterators are stateful and should not be shared across go routines.
type Iterator interface {
	// Next returns the next event in the stream.
	// - Returns (<event>, nil) when the next event is published.
	// - Returns (nil, ErrDone) when the stream is exhausted.
	// - Returns (nil, ctx.Err()) if the context is cancelled.
	// Blocks until one of these three outcomes occurs.
	Next(ctx context.Context) (interface{}, error)
}
