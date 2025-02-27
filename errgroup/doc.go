// Package errgroup [golang.org/x/sync/errgroup.Group], providing both backwards compatibility,
// but also introducing new, safer APIs.
package errgroup

import "context"

// ErrGroup defines a compatibility interface between [Group] and [golang.org/x/sync/errgroup.Group].
type ErrGroup interface {
	// Go calls the given function in a new goroutine, passing the group context.
	//
	// The first call to return a non-nil error cancels the group's context.
	// The error will be returned by Wait.
	Go(func() error)
	// Wait blocks until all function calls from the Go method have returned, then
	// returns the first non-nil error (if any) from them.
	Wait() error
	// TryGo calls the given function in a new goroutine only if the number of
	// active goroutines in the group is currently below the configured limit.
	//
	// The return value reports whether the goroutine was started.
	TryGo(func() error) bool
	// SetLimit limits the number of active goroutines in this group to at most n.
	// A negative value indicates no limit.
	//
	// Any subsequent call to the Go method will block until it can add an active
	// goroutine without exceeding the configured limit.
	//
	// The limit must not be modified while any goroutines in the group are active.
	SetLimit(n int)
}

// ContextGroup is a variant of [golang.org/x/sync/errgroup.Group] that:
// - automatically catches panics
// - forces the correct context
// - manages the context lifetime
// - early-exits any calls to Go() or TryGo() once the context is canceled
//
// Unlike the golang version, ContextGroup cannot be reused after Wait() has
// been called, because the context is dead and no new go funcs will be run.
type ContextGroup interface {
	// Go calls the given function in a new goroutine, passing the group context.
	//
	// The first call to return a non-nil error cancels the group's context.
	// The error will be returned by Wait().
	//
	// Go returns immediately if the group context is already cancelled.
	Go(func(context.Context) error)
	// Wait blocks until all function calls from the Go method have returned, then
	// returns the first non-nil error (if any) from them.
	Wait() error
	// TryGo calls the given function in a new goroutine only if the number of
	// active goroutines in the group is currently below the configured limit.
	//
	// The return value reports whether the goroutine was started.
	// TryGo returns true immediately if the group context is already cancelled.
	TryGo(func(context.Context) error) bool
}
