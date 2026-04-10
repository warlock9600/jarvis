package common

import "fmt"

const (
	ExitOK      = 0
	ExitError   = 1
	ExitPartial = 2
)

type CLIError struct {
	Code int
	Msg  string
	Err  error
}

func (e *CLIError) Error() string {
	if e.Err == nil {
		return e.Msg
	}
	if e.Msg == "" {
		return e.Err.Error()
	}
	return fmt.Sprintf("%s: %v", e.Msg, e.Err)
}

func NewExitError(code int, msg string, err error) error {
	return &CLIError{Code: code, Msg: msg, Err: err}
}

func ExitCode(err error) int {
	if err == nil {
		return ExitOK
	}
	var ee *CLIError
	if ok := As(err, &ee); ok {
		if ee.Code > 0 {
			return ee.Code
		}
	}
	return ExitError
}
