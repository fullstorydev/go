package errgroup

import (
	"context"
	"fmt"
	"sync"
)

// SafeGroup is a variant of golang.org/x/sync/errgroup.Group that automatically catches panics and forces the correct context.
type SafeGroup interface {
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

type safeGroup struct {
	ctx    context.Context
	cancel func(error)

	wg sync.WaitGroup

	sem chan struct{}

	errOnce sync.Once
	err     error
}

var _ SafeGroup = (*safeGroup)(nil)

func (g *safeGroup) done() {
	if g.sem != nil {
		<-g.sem
	}
	g.wg.Done()
}

// New returns a new SafeGroup derived from ctx.
//
// All funcs passed into [SafeGroup.Go] are wrapped with panic handlers and receive the
// group context automatically.
func New(ctx context.Context) SafeGroup {
	ctx, cancel := context.WithCancelCause(ctx)
	return &safeGroup{ctx: ctx, cancel: cancel}
}

// Wait blocks until all function calls from the Go method have returned, then returns the first non-nil error (if any) from them.
func (g *safeGroup) Wait() error {
	g.wg.Wait()
	g.cancel(g.err)
	return g.err
}

// Go calls the given function in a new goroutine.
func (g *safeGroup) Go(f func(context.Context) error) {
	if g.sem != nil {
		select {
		case <-g.ctx.Done():
			g.error(g.ctx.Err())
			return
		case g.sem <- struct{}{}:
		}
	}

	if err := g.ctx.Err(); err != nil {
		g.error(err)
		return
	}

	g.wg.Add(1)
	go func() {
		defer g.done()

		if err := recoverWrapperCtx(f)(g.ctx); err != nil {
			g.error(err)
		}
	}()
}

func (g *safeGroup) TryGo(f func(context.Context) error) bool {
	if g.sem != nil {
		select {
		case g.sem <- struct{}{}:
			// Note: this allows barging iff channels in general allow barging.
		case <-g.ctx.Done():
			g.error(g.ctx.Err())
			return true
		default:
			return false
		}
	}

	if err := g.ctx.Err(); err != nil {
		g.error(err)
		return true
	}

	g.wg.Add(1)
	go func() {
		defer g.done()

		if err := recoverWrapperCtx(f)(g.ctx); err != nil {
			g.error(err)
		}
	}()
	return true
}

func (g *safeGroup) SetLimit(n int) {
	if n < 0 {
		g.sem = nil
		return
	}
	if len(g.sem) != 0 {
		panic(fmt.Errorf("errgroup: modify limit while %v goroutines in the safeGroup are still active", len(g.sem)))
	}
	g.sem = make(chan struct{}, n)
}

func (g *safeGroup) error(err error) {
	g.errOnce.Do(func() {
		g.err = err
		g.cancel(err)
	})
}

func recoverWrapperCtx(f func(context.Context) error) func(context.Context) error {
	return func(ctx context.Context) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = PanicError{Recovered: r}
			}
		}()
		return f(ctx)
	}
}
