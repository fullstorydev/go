package errgroup

import (
	"fmt"
	"runtime"
)

var (
	strip   = []byte(`github.com/fullstorydev/go/errgroup/panic.go`)
	newline = []byte("\n")
)

const (
	pcFrames = 256
	stackLen = 64 << 10
)

func NewPanicError(recovered any) *PanicError {
	pcs := make([]uintptr, pcFrames)
	pcs = pcs[:runtime.Callers(1, pcs)]

	// TODO: capture goroutine number?

	return &PanicError{recovered: recovered, pcs: pcs}
}

type PanicError struct {
	recovered any
	pcs       []uintptr
}

func (e *PanicError) Recovered() any {
	return e.recovered
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

func (e *PanicError) Stack() []uintptr {
	return append([]uintptr{}, e.pcs...)
}

func (e *PanicError) StackFrames() *runtime.Frames {
	return runtime.CallersFrames(e.pcs)
}

func (e *PanicError) StackTrace() string {
	return ""
}
