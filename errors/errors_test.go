package errors

import (
	stdlib "errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/pkg/errors"
)

func TestCause(t *testing.T) {
	std := stdlib.New("this is a stdlib error")

	cases := map[string]struct {
		err  error
		root error
	}{
		"Errors are self-causing": {
			err:  ErrNotFound,
			root: ErrNotFound,
		},
		"Wrap reveals root cause": {
			err:  Wrap(ErrNotFound, "foo"),
			root: ErrNotFound,
		},
		"Cause works for stderr as root": {
			err:  Wrap(std, "Some helpful text"),
			root: std,
		},
	}

	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			if got := errors.Cause(tc.err); got != tc.root {
				t.Fatal("unexpected result")
			}
		})
	}
}

func TestErrorIs(t *testing.T) {
	cases := map[string]struct {
		a      *Error
		b      error
		wantIs bool
	}{
		"instance of the same error": {
			a:      ErrNotFound,
			b:      ErrNotFound,
			wantIs: true,
		},
		"two different coded errors": {
			a:      ErrNotFound,
			b:      ErrModel,
			wantIs: false,
		},
		"successful comparison to a wrapped error": {
			a:      ErrNotFound,
			b:      errors.Wrap(ErrNotFound, "gone"),
			wantIs: true,
		},
		"unsuccessful comparison to a wrapped error": {
			a:      ErrNotFound,
			b:      errors.Wrap(ErrOverflow, "too big"),
			wantIs: false,
		},
		"not equal to stdlib error": {
			a:      ErrNotFound,
			b:      fmt.Errorf("stdlib error"),
			wantIs: false,
		},
		"not equal to a wrapped stdlib error": {
			a:      ErrNotFound,
			b:      errors.Wrap(fmt.Errorf("stdlib error"), "wrapped"),
			wantIs: false,
		},
		"nil is nil": {
			a:      nil,
			b:      nil,
			wantIs: true,
		},
		"nil is any error nil": {
			a:      nil,
			b:      (*customError)(nil),
			wantIs: true,
		},
		"nil is not not-nil": {
			a:      nil,
			b:      ErrNotFound,
			wantIs: false,
		},
		"not-nil is not nil": {
			a:      ErrNotFound,
			b:      nil,
			wantIs: false,
		},
		"multierr with the same error": {
			a:      ErrNotFound,
			b:      Append(ErrNotFound, ErrState),
			wantIs: true,
		},
		"multierr with random order": {
			a:      ErrNotFound,
			b:      Append(ErrState, ErrNotFound),
			wantIs: true,
		},
		"multierr with wrapped err": {
			a:      ErrNotFound,
			b:      Append(ErrState, Wrap(ErrNotFound, "test")),
			wantIs: true,
		},
		"multierr with nil error": {
			a:      ErrNotFound,
			b:      Append(nil, nil),
			wantIs: false,
		},
		"multierr with different error": {
			a:      ErrNotFound,
			b:      Append(ErrState, nil),
			wantIs: false,
		},
		"multierr from nil": {
			a:      nil,
			b:      Append(ErrState, ErrNotFound),
			wantIs: false,
		},
	}
	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			if got := tc.a.Is(tc.b); got != tc.wantIs {
				t.Fatalf("unexpected result - got:%v want: %v", got, tc.wantIs)
			}
		})
	}
}

func TestUnwrap(t *testing.T) {
	specs := map[string]struct {
		src    error
		expErr error
		expMsg []string
	}{
		"Wrapped":        {src: Wrap(ErrNotFound, "myMsg"), expErr: ErrNotFound, expMsg: []string{"myMsg"}},
		"Double wrapped": {src: Wrap(Wrap(ErrNotFound, "first"), "second"), expErr: ErrNotFound, expMsg: []string{"first", "second"}},
		"Not Wrapped":    {src: ErrNotFound, expErr: ErrNotFound},
		"Nil":            {src: nil, expErr: nil},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			root, msgs := unWrap(spec.src)
			if exp, got := spec.expErr, root; !reflect.DeepEqual(exp, got) {
				t.Errorf("expected %T but got %T", exp, got)
			}
			if exp, got := spec.expMsg, msgs; !reflect.DeepEqual(exp, got) {
				t.Errorf("expected %v but got %v", exp, got)
			}
		})
	}
}

type customError struct {
}

func (customError) Error() string {
	return "custom error"
}

func TestWrapEmpty(t *testing.T) {
	if err := Wrap(nil, "wrapping <nil>"); err != nil {
		t.Fatal(err)
	}
}
