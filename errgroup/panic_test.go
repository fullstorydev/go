package errgroup

import (
	"fmt"
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
		{"%+v", "panic: test panic\nStack trace:\nTODO"},
		{"%#v", `&errgroup.PanicError{recovered:"test panic"}`},
	} {
		got := fmt.Sprintf(tc.fmt, err)
		if got != tc.want {
			t.Errorf("got:  %q, want: %q", got, tc.want)
		}
	}
	//panic("what")
}
