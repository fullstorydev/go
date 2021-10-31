package eventstream

import (
	"context"
)

type iterator struct {
	p Promise
}

func (it *iterator) Next(ctx context.Context) (interface{}, error) {
	if it.p == nil {
		return nil, ErrDone
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	select {
	case <-it.p.Ready():
		v, p := it.p.Next()
		it.p = p
		if p == nil {
			return nil, ErrDone
		}
		return v, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
