package test

import (
	"context"
	"github.com/fullstorydev/go/eventstream"
	"testing"

	"golang.org/x/sync/errgroup"
	"gotest.tools/v3/assert"
)

func TestEventStream_Serial(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	es := eventstream.NewWithBuffer(16) // small buffer so we exercise rolling over

	it1 := es.Subscribe().Iterator() // sees everything
	var it2 eventstream.Iterator
	for i := 0; i < 100; i++ {
		if i == 50 {
			it2 = es.Subscribe().Iterator() // sees 50 - 99
		}
		es.Publish(i)
	}
	it3 := es.Subscribe().Iterator() // sees nothing
	es.Close()
	it4 := es.Subscribe().Iterator() // subscribe after close sees nothing

	for i := 0; i < 100; i++ {
		v, err := it1.Next(ctx)
		assert.NilError(t, err, "should not err")
		assert.Equal(t, i, v.(int), "wrong")
	}
	assertDone(ctx, t, it1)

	for i := 50; i < 100; i++ {
		v, err := it2.Next(ctx)
		assert.NilError(t, err, "should not err")
		assert.Equal(t, i, v.(int), "wrong")
	}
	assertDone(ctx, t, it2)

	assertDone(ctx, t, it3)
	assertDone(ctx, t, it4)
}

func TestEventStream_Concurrent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	es := eventstream.NewWithBuffer(16) // small buffer so we exercise rolling over

	itEverything := es.Subscribe().Iterator() // sees everything
	g.Go(func() error {
		for i := 0; i < 100; i++ {
			v, err := itEverything.Next(ctx)
			assert.NilError(t, err)
			assert.Equal(t, i, v.(int))
		}
		assertDone(ctx, t, itEverything)
		return nil
	})

	for i := 0; i < 100; i++ {
		if i == 50 {
			itHalf := es.Subscribe().Iterator() // sees 50 - 99
			g.Go(func() error {
				for i := 50; i < 100; i++ {
					v, err := itHalf.Next(ctx)
					assert.NilError(t, err)
					assert.Equal(t, i, v.(int))
				}
				assertDone(ctx, t, itHalf)
				return nil
			})
		}
		es.Publish(i)
	}
	itNone := es.Subscribe().Iterator() // sees nothing
	g.Go(func() error {
		assertDone(ctx, t, itNone)
		return nil
	})
	es.Close()
	itClosed := es.Subscribe().Iterator() // subscribe after close sees nothing
	g.Go(func() error {
		assertDone(ctx, t, itClosed)
		return nil
	})

	assert.NilError(t, g.Wait())
}

func assertDone(ctx context.Context, t *testing.T, p eventstream.Iterator) {
	t.Helper()
	v, err := p.Next(ctx)
	assert.Assert(t, v == nil)
	assert.Equal(t, eventstream.ErrDone, err)
}
