package eventstream_test

import (
	"context"
	"testing"

	"github.com/fullstorydev/go/eventstream"
	"golang.org/x/sync/errgroup"
	"gotest.tools/v3/assert"
)

func TestEventStream_Serial(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	es := eventstream.NewEventStreamWithBuffer(16) // small buffer so we exercise rolling over

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
	assertDone(t, ctx, it1)

	for i := 50; i < 100; i++ {
		v, err := it2.Next(ctx)
		assert.NilError(t, err, "should not err")
		assert.Equal(t, i, v.(int), "wrong")
	}
	assertDone(t, ctx, it2)

	assertDone(t, ctx, it3)
	assertDone(t, ctx, it4)
}

func TestEventStream_Concurrent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	es := eventstream.NewEventStreamWithBuffer(16) // small buffer so we exercise rolling over

	it1 := es.Subscribe().Iterator() // sees everything
	g.Go(func() error {
		for i := 0; i < 100; i++ {
			v, err := it1.Next(ctx)
			assert.NilError(t, err, "should not err")
			assert.Equal(t, i, v.(int), "wrong")
		}
		assertDone(t, ctx, it1)
		return nil
	})

	for i := 0; i < 100; i++ {
		if i == 50 {
			it2 := es.Subscribe().Iterator() // sees 50 - 99
			g.Go(func() error {
				for i := 50; i < 100; i++ {
					v, err := it2.Next(ctx)
					assert.NilError(t, err, "should not err")
					assert.Equal(t, i, v.(int), "wrong")
				}
				assertDone(t, ctx, it2)
				return nil
			})
		}
		es.Publish(i)
	}
	it3 := es.Subscribe().Iterator() // sees nothing
	g.Go(func() error {
		assertDone(t, ctx, it3)
		return nil
	})
	es.Close()
	it4 := es.Subscribe().Iterator() // subscribe after close sees nothing
	g.Go(func() error {
		assertDone(t, ctx, it4)
		return nil
	})

	assert.NilError(t, g.Wait(), "errgroup failed")
}

func assertDone(t *testing.T, ctx context.Context, p eventstream.Iterator) {
	t.Helper()
	v, err := p.Next(ctx)
	assert.Assert(t, v == nil)
	assert.Equal(t, eventstream.ErrDone, err, "should be io.EOF")
}
