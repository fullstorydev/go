package errgroup

import (
	"context"
	"fmt"
	"sync"
)

type token struct{}

// Group is a drop-in replacement for [golang.org/x/sync/errgroup.Group] that automatically catches panics.
// Provided for backwards compatibility; prefer New() [ContextGroup] instead.
type Group struct {
	cancel func(error)

	wg sync.WaitGroup

	sem chan token

	errOnce sync.Once
	err     error
}

func (g *Group) done() {
	if g.sem != nil {
		<-g.sem
	}
	g.wg.Done()
}

// WithContext returns a new Group derived from ctx.
//
// All funcs passed into [Group.Go] are wrapped with panic handlers.
func WithContext(ctx context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancelCause(ctx)
	return &Group{cancel: cancel}, ctx
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the first non-nil error (if any) from them.
func (g *Group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel(g.err)
	}
	return g.err
}

// Go calls the given function in a new goroutine, passing the group context.
//
// The first call to return a non-nil error cancels the group's context.
// The error will be returned by Wait.
func (g *Group) Go(f func() error) {
	if g.sem != nil {
		g.sem <- token{}
	}

	g.wg.Add(1)
	go func() {
		defer g.done()
		panicked := true
		defer func() {
			if panicked {
				g.error(NewPanicError(recover()))
			}
		}()
		err := f()
		panicked = false
		if err != nil {
			g.error(err)
		}
	}()
}

// TryGo calls the given function in a new goroutine only if the number of
// active goroutines in the group is currently below the configured limit.
//
// The return value reports whether the goroutine was started.
func (g *Group) TryGo(f func() error) bool {
	if g.sem != nil {
		select {
		case g.sem <- token{}:
			// Note: this allows barging iff channels in general allow barging.
		default:
			return false
		}
	}

	g.wg.Add(1)
	go func() {
		defer g.done()
		panicked := true
		defer func() {
			if panicked {
				g.error(NewPanicError(recover()))
			}
		}()
		err := f()
		panicked = false
		if err != nil {
			g.error(err)
		}
	}()
	return true
}

// SetLimit limits the number of active goroutines in this group to at most n.
// A negative value indicates no limit.
//
// Any subsequent call to the Go method will block until it can add an active
// goroutine without exceeding the configured limit.
//
// The limit must not be modified while any goroutines in the group are active.
func (g *Group) SetLimit(n int) {
	if n < 0 {
		g.sem = nil
		return
	}
	if len(g.sem) != 0 {
		panic(fmt.Errorf("errgroup: modify limit while %v goroutines in the group are still active", len(g.sem)))
	}
	g.sem = make(chan token, n)
}

func (g *Group) error(err error) {
	g.errOnce.Do(func() {
		g.err = err
		if g.cancel != nil {
			g.cancel(err)
		}
	})
}
