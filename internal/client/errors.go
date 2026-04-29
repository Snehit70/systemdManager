package client

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type ServiceError struct {
	Op   string
	Unit string
	Code int
	Err  error
}

func (e *ServiceError) Error() string {
	parts := []string{e.Op}
	if e.Unit != "" {
		parts = append(parts, e.Unit)
	}
	if e.Code != 0 {
		parts = append(parts, fmt.Sprintf("exit code %d", e.Code))
	}
	if e.Err != nil {
		parts = append(parts, e.Err.Error())
	}
	return strings.Join(parts, ": ")
}

func (e *ServiceError) Unwrap() error {
	return e.Err
}

func (e *ServiceError) Is(target error) bool {
	return errors.Is(e.Err, target)
}

var (
	ErrNotFound         = errors.New("service not found")
	ErrPermissionDenied = errors.New("permission denied")
	ErrNotRunning       = errors.New("service not running")
	ErrNotEnabled       = errors.New("service not enabled")
)

func NewServiceError(op, unit string, cmdErr error) *ServiceError {
	err := &ServiceError{
		Op:   op,
		Unit: unit,
		Err:  cmdErr,
	}
	if cmdErr == nil {
		return err
	}

	var exitErr *exec.ExitError
	if errors.As(cmdErr, &exitErr) {
		err.Code = exitErr.ExitCode()
	}

	msg := strings.ToLower(cmdErr.Error())
	if strings.Contains(msg, "not found") || strings.Contains(msg, "could not be found") {
		err.Err = errors.Join(err.Err, ErrNotFound)
	}
	if strings.Contains(msg, "permission denied") || strings.Contains(msg, "access denied") {
		err.Err = errors.Join(err.Err, ErrPermissionDenied)
	}

	return err
}

func UserErrorMessage(err *ServiceError) string {
	if err == nil {
		return "Unknown service error"
	}

	switch {
	case errors.Is(err.Err, ErrNotFound):
		return fmt.Sprintf("Service '%s' not found", err.Unit)
	case errors.Is(err.Err, ErrPermissionDenied):
		return "Permission denied. Try running with --user flag or check permissions."
	case errors.Is(err.Err, ErrNotRunning):
		return fmt.Sprintf("Service '%s' is not running", err.Unit)
	case errors.Is(err.Err, ErrNotEnabled):
		return fmt.Sprintf("Service '%s' is not enabled", err.Unit)
	default:
		return err.Error()
	}
}
