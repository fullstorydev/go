package errgroup_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fullstorydev/go/errgroup"
)

// JustErrors illustrates the use of a Group in place of a sync.WaitGroup to
// simplify goroutine counting and error handling. This example is derived from
// the sync.WaitGroup example at https://golang.org/pkg/sync/#example_WaitGroup.
func ExampleNew_justErrors() {
	g := errgroup.New(context.Background())
	var urls = []string{
		"http://www.golang.org/",
		"http://www.google.com/",
		"http://www.somestupidname.com/",
	}
	for _, url := range urls {
		// Launch a goroutine to fetch the URL.
		url := url // https://golang.org/doc/faq#closures_and_goroutines
		g.Go(func(ctx context.Context) error {
			// Fetch the URL.
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				panic(err)
			}
			resp, err := http.DefaultClient.Do(req)
			if err == nil {
				_ = resp.Body.Close()
			}
			return err
		})
	}
	// Wait for all HTTP fetches to complete.
	if err := g.Wait(); err == nil {
		fmt.Println("Successfully fetched all URLs.")
	}

	// Output:
	// Successfully fetched all URLs.
}

// Parallel illustrates the use of a Group for synchronizing a simple parallel
// task: the "Google Search 2.0" function from
// https://talks.golang.org/2012/concurrency.slide#46, augmented with a Context
// and error-handling.
func ExampleNew_parallel() {
	Google := func(ctx context.Context, query string) ([]Result, error) {
		g := errgroup.New(ctx)

		searches := []Search{Web, Image, Video}
		results := make([]Result, len(searches))
		for i, search := range searches {
			i, search := i, search // https://golang.org/doc/faq#closures_and_goroutines
			g.Go(func(ctx context.Context) error {
				result, err := search(ctx, query)
				if err == nil {
					results[i] = result
				}
				return err
			})
		}
		if err := g.Wait(); err != nil {
			return nil, err
		}
		return results, nil
	}

	results, err := Google(context.Background(), "golang")
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return
	}
	for _, result := range results {
		fmt.Println(result)
	}

	// Output:
	// web result for "golang"
	// image result for "golang"
	// video result for "golang"
}

func TestZeroNew(t *testing.T) {
	err1 := errors.New("errgroup_test: 1")
	err2 := errors.New("errgroup_test: 2")

	cases := []struct {
		errs []error
	}{
		{errs: []error{}},
		{errs: []error{nil}},
		{errs: []error{err1}},
		{errs: []error{err1, nil}},
		{errs: []error{err1, nil, err2}},
	}

	for _, tc := range cases {
		g := errgroup.New(context.Background())

		var firstErr error
		for i, err := range tc.errs {
			err := err
			g.Go(func(ctx context.Context) error { return err })

			if firstErr == nil && err != nil {
				firstErr = err
			}

			if gErr := g.Wait(); gErr != firstErr {
				t.Errorf("after %T.Go(func() error { return err }) for err in %v\n"+
					"g.Wait() = %v; want %v",
					g, tc.errs[:i+1], err, firstErr)
			}
		}
	}
}

func TestNew(t *testing.T) {
	errDoom := errors.New("group_test: doomed")

	cases := []struct {
		errs []error
		want error
	}{
		{want: nil},
		{errs: []error{nil}, want: nil},
		{errs: []error{errDoom}, want: errDoom},
		{errs: []error{errDoom, nil}, want: errDoom},
	}

	for _, tc := range cases {
		g := errgroup.New(context.Background())

		for _, err := range tc.errs {
			err := err
			g.Go(func(ctx context.Context) error { return err })
		}

		if err := g.Wait(); err != tc.want {
			t.Errorf("after %T.Go(func() error { return err }) for err in %v\n"+
				"g.Wait() = %v; want %v",
				g, tc.errs, err, tc.want)
		}
	}
}

func TestNewTryGo(t *testing.T) {
	g := errgroup.WithLimit(42).New(context.Background())
	n := 42
	ch := make(chan struct{})
	fn := func(ctx context.Context) error {
		ch <- struct{}{}
		return nil
	}
	for i := 0; i < n; i++ {
		if !g.TryGo(fn) {
			t.Fatalf("TryGo should succeed but got fail at %d-th call.", i)
		}
	}
	if g.TryGo(fn) {
		t.Fatalf("TryGo is expected to fail but succeeded.")
	}
	go func() {
		for i := 0; i < n; i++ {
			<-ch
		}
	}()
	_ = g.Wait()

	// disabled: cannot reuse ContextGroup.
	/*
		if !g.TryGo(fn) {
			t.Fatalf("TryGo should success but got fail after all goroutines.")
		}
		go func() { <-ch }()
		g.Wait()
	*/

	// Switch limit.
	g = errgroup.WithLimit(1).New(context.Background())
	if !g.TryGo(fn) {
		t.Fatalf("TryGo should success but got failed.")
	}
	if g.TryGo(fn) {
		t.Fatalf("TryGo should fail but succeeded.")
	}
	go func() { <-ch }()
	_ = g.Wait()

	// Block all calls.
	g = errgroup.WithLimit(0).New(context.Background())
	for i := 0; i < 1<<10; i++ {
		if g.TryGo(fn) {
			t.Fatalf("TryGo should fail but got succeded.")
		}
	}
	_ = g.Wait()
}

func TestNewGoLimit(t *testing.T) {
	const limit = 10

	g := errgroup.WithLimit(limit).New(context.Background())
	var active int32
	for i := 0; i <= 1<<10; i++ {
		g.Go(func(ctx context.Context) error {
			n := atomic.AddInt32(&active, 1)
			if n > limit {
				return fmt.Errorf("saw %d active goroutines; want ≤ %d", n, limit)
			}
			time.Sleep(1 * time.Microsecond) // Give other goroutines a chance to increment active.
			atomic.AddInt32(&active, -1)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		t.Fatal(err)
	}
}

func TestNewPanic(t *testing.T) {
	t.Run("Go", func(t *testing.T) {
		g := errgroup.New(context.Background())
		g.Go(func(ctx context.Context) error {
			panic("test panic")
		})
		err := g.Wait()
		if err == nil {
			t.Fatalf("Wait should return an error")
		}
		if !strings.HasPrefix(err.Error(), "panic: test panic") {
			t.Fatalf("Error message mismatch: %v", err)
		}
	})

	t.Run("TryGo", func(t *testing.T) {
		g := errgroup.WithLimit(1).New(context.Background())
		g.TryGo(func(ctx context.Context) error {
			panic("test panic")
		})
		err := g.Wait()
		if err == nil {
			t.Fatalf("Wait should return an error")
		}
		if !strings.HasPrefix(err.Error(), "panic: test panic") {
			t.Fatalf("Error message mismatch: %v", err)
		}
	})
}

func TestNewGoExitsEarly(t *testing.T) {
	var counter atomic.Uint32

	fn := func(ctx context.Context) error {
		counter.Add(1)
		<-ctx.Done() // wait until cancelled
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	g := errgroup.WithLimit(1).New(ctx)

	g.Go(fn) // this should succeed
	g.Go(fn) // this should get stuck, then cancelled

	_ = g.Wait()

	if count := counter.Load(); count != 1 {
		t.Fatalf("Counter should be 1, got %d", count)
	}
}

func BenchmarkNewGo(b *testing.B) {
	fn := func() {}
	g := errgroup.New(context.Background())
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		g.Go(func(ctx context.Context) error { fn(); return nil })
	}
	_ = g.Wait()
}
