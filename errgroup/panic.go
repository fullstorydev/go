package errgroup

import (
	"fmt"
)

type PanicError struct {
	Recovered any
}

func (e PanicError) Error() string {
	return fmt.Sprintf("panic: %v", e.Recovered)
}
