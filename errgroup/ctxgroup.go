package errgroup

import (
	"context"
	"sync"
)

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

type ctxGroup struct {
	ctx    context.Context
	cancel func(error)

	wg sync.WaitGroup

	sem chan struct{}

	errOnce sync.Once
	err     error
}

var _ ContextGroup = (*ctxGroup)(nil)

func (g *ctxGroup) done() {
	if g.sem != nil {
		<-g.sem
	}
	g.wg.Done()
}

// New returns a new ContextGroup derived from ctx.
//
// All funcs passed into [ContextGroup.Go] are wrapped with panic handlers and receive the
// group context automatically.
func New(ctx context.Context) ContextGroup {
	ctx, cancel := context.WithCancelCause(ctx)
	return &ctxGroup{ctx: ctx, cancel: cancel}
}

// NewWithLimit returns a new ContextGroup derived from ctx.
//
// Any subsequent call to the Go method will block until it can add an active
// goroutine without exceeding the configured limit.
func NewWithLimit(ctx context.Context, limit int) ContextGroup {
	if limit < 0 {
		panic("negative limit")
	}
	ctx, cancel := context.WithCancelCause(ctx)
	return &ctxGroup{ctx: ctx, cancel: cancel, sem: make(chan struct{}, limit)}
}

// Wait blocks until all function calls from the Go method have returned, then returns the first non-nil error (if any) from them.
func (g *ctxGroup) Wait() error {
	g.wg.Wait()
	g.cancel(g.err)
	return g.err
}

// Go calls the given function in a new goroutine.
func (g *ctxGroup) Go(f func(context.Context) error) {
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
		panicked := true
		defer func() {
			if panicked {
				g.error(NewPanicError(recover()))
			}
		}()
		err := f(g.ctx)
		panicked = false
		if err != nil {
			g.error(err)
		}
	}()
}

func (g *ctxGroup) TryGo(f func(context.Context) error) bool {
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
		panicked := true
		defer func() {
			if panicked {
				g.error(NewPanicError(recover()))
			}
		}()
		err := f(g.ctx)
		panicked = false
		if err != nil {
			g.error(err)
		}
	}()
	return true
}

func (g *ctxGroup) error(err error) {
	g.errOnce.Do(func() {
		g.err = err
		g.cancel(err)
	})
}

type ctxGroupBuilder struct {
	limit int
}

func (b ctxGroupBuilder) New(ctx context.Context) ContextGroup {
	ctx, cancel := context.WithCancelCause(ctx)
	return &ctxGroup{ctx: ctx, cancel: cancel, sem: make(chan struct{}, b.limit)}
}

func Limit(limit int) ctxGroupBuilder {
	if limit < 0 {
		panic("negative limit")
	}
	return ctxGroupBuilder{limit: limit}
}
