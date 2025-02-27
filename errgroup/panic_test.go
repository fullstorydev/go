package errgroup

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestPanicError(t *testing.T) {
	err := NewPanicError("test panic")
	for _, tc := range []struct {
		fmt  string
		want string
	}{
		{"%v", `panic: test panic`},
		{"%s", `panic: test panic`},
		{"%q", `"panic: test panic"`},
		{"%#v", `&errgroup.PanicError{recovered:"test panic"}`},
	} {
		got := fmt.Sprintf(tc.fmt, err)
		if got != tc.want {
			t.Errorf("got:  %q, want: %q", got, tc.want)
		}
	}

	// test %+v manually
	got := fmt.Sprintf("%+v", err)
	lines := strings.Split(got, "\n")
	if lines[0] != "panic: test panic [recovered]" {
		t.Error(lines[0])
	}
	if lines[1] != "" {
		t.Error(lines[1])
	}
	if !strings.HasPrefix(lines[2], "goroutine") {
		t.Error("expected goroutine stack trace")
	}
	// lines 1 and 2 contain the explicit call to panic
	if !strings.Contains(lines[3], "errgroup.TestPanicError") {
		t.Error("expected errgroup.TestPanicError")
	}
	if !strings.Contains(lines[4], "errgroup/panic_test.go") {
		t.Error("expected errgroup/panic_test.go")
	}
}

func TestPanicErrorIntegration(t *testing.T) {
	var g Group
	g.Go(func() error {
		panic("test panic")
	})
	var pe *PanicError
	err := g.Wait()
	if !errors.As(err, &pe) {
		t.Fatal("expected panic error")
	}
	trace := pe.StackTrace()
	lines := strings.Split(trace, "\n")
	if !strings.HasPrefix(lines[0], "goroutine") {
		t.Error("expected goroutine stack trace")
	}
	// lines 1 and 2 contain the explicit call to panic
	if !strings.Contains(lines[1], "panic(") {
		t.Error("expected panic(")
	}
	if !strings.Contains(lines[2], "runtime/panic.go") {
		t.Error("expected runtime/panic.go")
	}
	// lines 3 and 4 is our func
	if !strings.Contains(lines[3], "errgroup.TestPanicErrorIntegration.func1()") {
		t.Error("expected errgroup.TestPanicErrorIntegration.func1()")
	}
	if !strings.Contains(lines[4], "errgroup/panic_test.go") {
		t.Error("expected errgroup/panic_test.go")
	}
}

func TestPanicErrorStackTrace(t *testing.T) {
	err := NewPanicError("test panic")
	trace := err.StackTrace()
	lines := strings.Split(trace, "\n")
	if !strings.HasPrefix(lines[0], "goroutine") {
		t.Error("expected goroutine stack trace")
	}
	if !strings.Contains(lines[1], "errgroup.TestPanicErrorStackTrace") {
		t.Error("expected errgroup.TestPanicErrorStackTrace")
	}
	if !strings.Contains(lines[2], "errgroup/panic_test.go") {
		t.Error("expected errgroup/panic_test.go")
	}
}

func TestPanicErrorStackFrames(t *testing.T) {
	err := NewPanicError("test panic")
	first, _ := err.StackFrames().Next()
	const want = `github.com/fullstorydev/go/errgroup.TestPanicErrorStackFrames`
	if got := first.Function; got != want {
		t.Errorf("got:  %q, want: %q", got, want)
	}
}

func TestPanicErrorStackFramesCallers(t *testing.T) {
	err := NewPanicErrorCallers("test panic", 0)
	first, _ := err.StackFrames().Next()
	const want = `github.com/fullstorydev/go/errgroup.NewPanicErrorCallers`
	if got := first.Function; got != want {
		t.Errorf("got:  %q, want: %q", got, want)
	}
}
