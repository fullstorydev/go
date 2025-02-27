package errgroup

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
)

const (
	pcFrames  = 1 << 8
	stackSize = 1 << 16
)

var (
	newline = []byte{'\n'}
)

// NewPanicError creates a new PanicError with the given recovered panic value.
// The caller stack frames are captured and attached, starting with the caller of NewPanicError.
func NewPanicError(recovered any) *PanicError {
	return NewPanicErrorCallers(recovered, 2)
}

// NewPanicErrorCallers creates a new PanicError with the given recovered panic value.
//
// The argument skip is the number of stack frames to skip before recording in pc,
// with 0 identifying the frame for NewPanicErrorCallers itself and 1 identifying the
// caller of NewPanicErrorCallers.
func NewPanicErrorCallers(recovered any, skip int) *PanicError {
	pcs := make([]uintptr, pcFrames)
	pcs = pcs[:runtime.Callers(skip+1, pcs)]

	// Manually prune the string stack trace based on skip (this is awkward).
	skipLines := 1 + 2*skip
	stack := make([]byte, stackSize)
	stack = stack[:runtime.Stack(stack, false)]
	var sb strings.Builder
	for i, line := range bytes.Split(stack, newline) {
		if i == 0 || i >= skipLines {
			sb.Write(line)
			sb.WriteByte('\n')
		}
	}

	return &PanicError{recovered: recovered, pcs: pcs, stack: sb.String()}
}

// PanicError represents a wrapped recovered panic value.
type PanicError struct {
	recovered any
	pcs       []uintptr
	stack     string
}

var _ error = (*PanicError)(nil)
var _ fmt.Formatter = (*PanicError)(nil)

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
			_, _ = fmt.Fprintf(s, "panic: %v [recovered]\n\n%s", e.recovered, e.StackTrace())
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

// StackTrace returns the originally captured stack trace as a multiline string.
func (e *PanicError) StackTrace() string {
	return e.stack
}
