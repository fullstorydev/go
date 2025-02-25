package errgroup

import (
	"context"
	"fmt"
	"sync"
)

// Group is a variant of golang.org/x/sync/errgroup.Group that forces the correct context.
type Group interface {
	// Go calls the given function in a new goroutine, passing the group context.
	//
	// The first call to return a non-nil error cancels the group's context.
	// The error will be returned by Wait.
	Go(func(context.Context) error)
	// Wait blocks until all function calls from the Go method have returned, then
	// returns the first non-nil error (if any) from them.
	Wait() error
	// TryGo calls the given function in a new goroutine only if the number of
	// active goroutines in the group is currently below the configured limit.
	//
	// The return value reports whether the goroutine was started.
	TryGo(func(context.Context) error) bool
	// SetLimit limits the number of active goroutines in this group to at most n.
	// A negative value indicates no limit.
	//
	// Any subsequent call to the Go method will block until it can add an active
	// goroutine without exceeding the configured limit.
	//
	// The limit must not be modified while any goroutines in the group are active.
	SetLimit(n int)
}

type token struct{}

type group struct {
	ctx    context.Context
	cancel func(error)

	wg sync.WaitGroup

	sem chan token

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

// New returns a new Group derived from ctx.
//
// All funcs passed into [Group.Go] are wrapped with panic handlers and receive the
// group context automatically.
func New(ctx context.Context) Group {
	ctx, cancel := withCancelCause(ctx)
	return &group{ctx: ctx, cancel: cancel}
}

// Wait blocks until all function calls from the Go method have returned, then returns the first non-nil error (if any) from them.
func (g *group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel(g.err)
	}
	return g.err
}

// Go calls the given function in a new goroutine.
func (g *group) Go(f func(context.Context) error) {
	if g.sem != nil {
		select {
		case <-g.ctx.Done():
			return
		case g.sem <- token{}:
		}
	}

	g.wg.Add(1)
	go func() {
		defer g.done()

		if err := recoverWrapper(f)(g.ctx); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel(g.err)
				}
			})
		}
	}()
}

func (g *group) TryGo(f func(context.Context) error) bool {
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

		if err := recoverWrapper(f)(g.ctx); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel(g.err)
				}
			})
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
	g.sem = make(chan token, n)
}

// withCancelCause is a helper function to create a context with cancel function.
func withCancelCause(ctx context.Context) (context.Context, func(error)) {
	ctx, cancel := context.WithCancel(ctx)
	return ctx, func(err error) {
		cancel()
	}
}

type PanicError struct {
	Recovered any
}

func (e PanicError) Error() string {
	return fmt.Sprintf("panic: %v", e.Recovered)
}

func recoverWrapper(f func(context.Context) error) func(context.Context) error {
	return func(ctx context.Context) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = PanicError{Recovered: r}
			}
		}()
		return f(ctx)
	}
}
