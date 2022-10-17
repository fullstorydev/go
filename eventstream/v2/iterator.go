package eventstream

import (
	"context"
)

type iterator[T any] struct {
	p Promise[T]
}

func (it *iterator[T]) Next(ctx context.Context) (T, error) {
	var none T
	if it.p == nil {
		return none, ErrDone
	}
	if err := ctx.Err(); err != nil {
		return none, err
	}

	select {
	case <-it.p.Ready():
		v, p := it.p.Next()
		it.p = p
		if p == nil {
			return none, ErrDone
		}
		return v, nil
	case <-ctx.Done():
		return none, ctx.Err()
	}
}
