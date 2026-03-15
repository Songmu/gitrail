package gitrail

import "fmt"

type exitError struct {
	code int
	msg  string
	err  error
}

func (e *exitError) Error() string {
	return e.msg
}

func (e *exitError) ExitCode() int {
	return e.code
}

func (e *exitError) Unwrap() error {
	return e.err
}

func newExitError(code int, msg string) error {
	return &exitError{code: code, msg: msg}
}

func exitErrorf(code int, format string, args ...any) error {
	return &exitError{
		code: code,
		msg:  fmt.Sprintf(format, args...),
		err:  fmt.Errorf(format, args...),
	}
}
