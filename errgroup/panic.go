package errgroup

import (
	"fmt"
)

func NewPanicError(recovered any) *PanicError {
	// TODO: stack trace
	return &PanicError{recovered: recovered}
}

type PanicError struct {
	recovered any
}

func (e *PanicError) Error() string {
	return fmt.Sprintf("panic: %v", e.recovered)
}

func (e *PanicError) Unwrap() error {
	if wrapped, ok := e.recovered.(error); ok {
		return wrapped
	}
	return nil
}

func (e *PanicError) Recovered() any {
	return e.recovered
}
