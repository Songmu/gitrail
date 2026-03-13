package gitrail

import "fmt"

type exitError struct {
	code int
	msg  string
}

func (e *exitError) Error() string {
	return e.msg
}

func (e *exitError) ExitCode() int {
	return e.code
}

func newExitError(code int, msg string) error {
	return &exitError{code: code, msg: msg}
}

func exitErrorf(code int, format string, args ...any) error {
	return &exitError{code: code, msg: fmt.Sprintf(format, args...)}
}
