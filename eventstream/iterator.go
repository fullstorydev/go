package eventstream

import (
	"context"
)

type iterator[T any] struct {
	p Promise[T]
}

func (it *iterator[T]) Next(ctx context.Context) (T, error) {
	var zero T
	if it.p == nil {
		return zero, ErrDone
	}
	if err := ctx.Err(); err != nil {
		return zero, err
	}

	select {
	case <-it.p.Ready():
		v, p := it.p.Next()
		it.p = p
		if p == nil {
			return zero, ErrDone
		}
		return v, nil
	case <-ctx.Done():
		return zero, ctx.Err()
	}
}
