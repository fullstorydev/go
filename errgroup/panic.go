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

// Recovered returns the original value.
func (e *PanicError) Recovered() any {
	return e.recovered
}

// Error returns the full error, including stack trace.
func (e *PanicError) Error() string {
	return fmt.Sprintf("panic: %v [recovered]\n\n%s", e.recovered, e.stack)
}

// Message returns a short description, with the string value of the recovered object.
func (e *PanicError) Message() string {
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

// StackFrames returns a slice of program counters composing this error's stacktrace.
func (e *PanicError) StackFrames() []uintptr {
	return append([]uintptr{}, e.pcs...)
}

// StackTrace returns the originally captured stack trace as a multiline string.
func (e *PanicError) StackTrace() string {
	return e.stack
}
