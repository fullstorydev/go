# errgroup

`errgroup` is a safer alternative to the official [golang.org/x/sync/errgroup](https://pkg.go.dev/golang.org/x/sync/errgroup) package.

## errgroup.Group

`Group` is an API-compatible, drop-in replacement for [golang.org/x/sync/errgroup.Group](https://pkg.go.dev/golang.org/x/sync/errgroup#Group),
with the key difference that panics thrown from subtasks will not crash the Go process.
Any panics are caught and wrapped in a `PanicError`. If this is the first error returned from
a subtask, this error will be the error returned from `Wait()`.

### Using

You can start using this today simply by changing your import statements.

```diff
-import "golang.org/x/sync/errgroup"
+import "github.com/fullstorydev/go/errgroup"
```

## errgroup.ContextGroup

`ContextGroup` takes safety a step further: it is an alternative to [golang.org/x/sync/errgroup.WithContext()](https://pkg.go.dev/golang.org/x/sync/errgroup#WithContext).
`ContextGroup` owns and manages the group context, and its updated API requires subtask functions to accept a `context.Context` argument to use the group context.
This helps elimate a class of bugs where subtasks accidentally use a parent context rather than the correct group context.

### Features

- Automatically catches panics
- Forces the correct context
- Manages the context lifetime
- Early-exits any calls to Go() or TryGo() once the group context is cancelled
- Enforces that any `Limit` immutable and set at construction time.

### Using

Create a new `ContextGroup` via `errgroup.New(ctx)` or `errgroup.WithLimit(4).New(ctx)`.

Old code:
```go
func fetchUrls(ctx context.Context, urls []string) ([]string, error) {
	ret := make([]string, len(urls))
	g, gCtx := errgroup.WithContext(ctx)

	for i, url := range urls {
		i, url := i, url
		g.Go(func() error {
			// remember to use the correct ctx here
			body, err := httpGet(gCtx, url)
			ret[i] = body
			return err
		})
	}
	return ret, g.Wait()
}
```

New code:
```go
func fetchUrls(ctx context.Context, urls []string) ([]string, error) {
	ret := make([]string, len(urls))
	g := errgroup.New(ctx)

	for i, url := range urls {
		i, url := i, url
		g.Go(func(ctx context.Context) error {
			body, err := httpGet(ctx, url)
			ret[i] = body
			return err
		})
	}
	return ret, g.Wait()
}
```
