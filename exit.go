package charmcli

import (
	"context"
	"errors"
	"fmt"

	"charm.land/huh/v2"
)

// ExitCoder allows a returned error to control process exit status.
type ExitCoder interface {
	ExitCode() int
}

type codedError struct {
	code int
	err  error
}

func (e *codedError) Error() string { return e.err.Error() }
func (e *codedError) Unwrap() error { return e.err }
func (e *codedError) ExitCode() int { return e.code }

type statusError struct {
	code int
}

func (e *statusError) Error() string { return fmt.Sprintf("exit status %d", e.code) }
func (e *statusError) ExitCode() int { return e.code }
func (*statusError) Silent() bool    { return true }

// Usage marks an error as invalid CLI usage (exit 2).
func Usage(err error) error {
	if err == nil {
		return nil
	}
	return &codedError{code: 2, err: err}
}

// Failure marks a command execution failure such as provider, API, RBAC or
// configuration failure (exit 2). A valid command that rendered a negative
// domain result should use Exit(1) instead.
func Failure(err error) error {
	if err == nil {
		return nil
	}
	return &codedError{code: 2, err: err}
}

// Exit returns a silent status error. It is useful when a command has already
// rendered a valid result but intentionally needs a non-zero process status.
func Exit(code int) error {
	if code == 0 {
		return nil
	}
	return &statusError{code: code}
}

// ExitCode resolves an error into the charmcli process exit convention:
// 0 success, 1 explicit valid-negative/domain result, 2 invocation/execution
// failure, and 130 cancellation/user abort.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, huh.ErrUserAborted) {
		return 130
	}
	var coder ExitCoder
	if errors.As(err, &coder) {
		return coder.ExitCode()
	}
	return 2
}

func isSilentError(err error) bool {
	var silent interface{ Silent() bool }
	return errors.As(err, &silent) && silent.Silent()
}
