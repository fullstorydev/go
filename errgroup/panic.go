package errgroup

import (
	"fmt"
	"runtime"
)

const (
	pcFrames = 256
)

// NewPanicError creates a new PanicError with the given recovered panic value.
// The caller stack frames are captured and attached.
func NewPanicError(recovered any) *PanicError {
	pcs := make([]uintptr, pcFrames)
	pcs = pcs[:runtime.Callers(1, pcs)]

	// TODO: capture goroutine number?

	return &PanicError{recovered: recovered, pcs: pcs}
}

// NewPanicErrorCallers creates a new PanicError with the given recovered panic value.
// The caller stack frames are captured and attached.
func NewPanicErrorCallers(recovered any, skip int) *PanicError {
	pcs := make([]uintptr, pcFrames)
	pcs = pcs[:runtime.Callers(skip+1, pcs)]

	// TODO: capture goroutine number?

	return &PanicError{recovered: recovered, pcs: pcs}
}

// PanicError represents a wrapped recovered panic value.
type PanicError struct {
	recovered any
	pcs       []uintptr
}

// Recovered returns the original value.
func (e *PanicError) Recovered() any {
	return e.recovered
}

func (e *PanicError) Error() string {
	return fmt.Sprintf("panic: %v", e.recovered)
}

// Unwrap returns the recovered value if it is itself an error,
// otherwise returns nil.
func (e *PanicError) Unwrap() error {
	if wrapped, ok := e.recovered.(error); ok {
		return wrapped
	}
	return nil
}

// Format implements the fmt.Formatter interface to support rich formatting for %+v and %#v.
func (e *PanicError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			// Detailed format: include stack trace
			_, _ = fmt.Fprintf(s, "panic: %v\nStack trace:\n%s", e.recovered, e.StackTrace())
			return
		} else if s.Flag('#') {
			_, _ = fmt.Fprintf(s, "&errgroup.PanicError{recovered:%#v}", e.recovered)
			return
		}
	case 'q':
		_, _ = fmt.Fprintf(s, "%q", e.Error())
		return
	}
	_, _ = fmt.Fprintf(s, "panic: %v", e.recovered)
}

// Stack returns a copy of the originally captured stack as PCs.
func (e *PanicError) Stack() []uintptr {
	return append([]uintptr{}, e.pcs...)
}

// StackFrames returns a copy of the original captured stack as runtime.Frames.
func (e *PanicError) StackFrames() *runtime.Frames {
	return runtime.CallersFrames(e.Stack())
}

// StackTrace returns multiline string format of the originally captured stack
// approximating the format of an uncaught panic.
func (e *PanicError) StackTrace() string {
	return "TODO"
}
