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

func (it *iterator[T]) Consume(ctx context.Context, callback func(context.Context, T) error) error {
	for {
		val, err := it.Next(ctx)
		if err == ErrDone {
			return nil
		} else if err != nil {
			return err
		}
		if err := callback(ctx, val); err != nil {
			return err
		}
	}
}
