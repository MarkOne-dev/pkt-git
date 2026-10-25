package cli

import (
	"errors"
	"fmt"
)

// errUsage marks an invalid invocation so Execute can return exit code 2.
var errUsage = errors.New("usage error")

// usageError carries a user-facing usage message. It unwraps to errUsage so
// callers can classify it without string matching.
type usageError struct {
	message string
}

func (e *usageError) Error() string {
	if e.message == "" {
		return errUsage.Error()
	}
	return e.message
}

func (e *usageError) Unwrap() error {
	return errUsage
}

func usagef(format string, args ...any) error {
	return &usageError{message: fmt.Sprintf(format, args...)}
}
