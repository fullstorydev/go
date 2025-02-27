package errgroup

import (
	"context"
	"fmt"
	"sync"
)

// Group is a variant of golang.org/x/sync/errgroup.Group that automatically catches panics.
// Provided for drop-in backwards compatibility. New code should use [New] [ContextGroup] instead.
type Group interface {
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

type group struct {
	cancel func(error)

	wg sync.WaitGroup

	sem chan struct{}

	errOnce sync.Once
	err     error
}

var _ Group = (*group)(nil)

func (g *group) done() {
	if g.sem != nil {
		<-g.sem
	}
	g.wg.Done()
}

// WithContext returns a new Group derived from ctx.
//
// All funcs passed into [Group.Go] are wrapped with panic handlers.
func WithContext(ctx context.Context) (Group, context.Context) {
	ctx, cancel := context.WithCancelCause(ctx)
	return &group{cancel: cancel}, ctx
}

// Wait blocks until all function calls from the Go method have returned, then returns the first non-nil error (if any) from them.
func (g *group) Wait() error {
	g.wg.Wait()
	g.cancel(g.err)
	return g.err
}

// Go calls the given function in a new goroutine.
func (g *group) Go(f func() error) {
	if g.sem != nil {
		select {
		case g.sem <- struct{}{}:
		}
	}

	g.wg.Add(1)
	go func() {
		defer g.done()

		if err := recoverWrapper(f)(); err != nil {
			g.error(err)
		}
	}()
}

func (g *group) TryGo(f func() error) bool {
	if g.sem != nil {
		select {
		case g.sem <- struct{}{}:
			// Note: this allows barging iff channels in general allow barging.
		default:
			return false
		}
	}

	g.wg.Add(1)
	go func() {
		defer g.done()

		if err := recoverWrapper(f)(); err != nil {
			g.error(err)
		}
	}()
	return true
}

func (g *group) SetLimit(n int) {
	if n < 0 {
		g.sem = nil
		return
	}
	if len(g.sem) != 0 {
		panic(fmt.Errorf("errgroup: modify limit while %v goroutines in the group are still active", len(g.sem)))
	}
	g.sem = make(chan struct{}, n)
}

func (g *group) error(err error) {
	g.errOnce.Do(func() {
		g.err = err
		g.cancel(err)
	})
}

func recoverWrapper(f func() error) func() error {
	return func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = NewPanicError(r)
			}
		}()
		return f()
	}
}
