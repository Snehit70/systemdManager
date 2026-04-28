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
	return target == ErrNotFound || target == ErrPermissionDenied
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

	if exitErr, ok := cmdErr.(*exec.ExitError); ok {
		err.Code = exitErr.ExitCode()
	}

	if strings.Contains(cmdErr.Error(), "not found") || strings.Contains(cmdErr.Error(), "could not be found") {
		err.Err = errors.Join(err.Err, ErrNotFound)
	}
	if strings.Contains(cmdErr.Error(), "permission denied") || strings.Contains(cmdErr.Error(), "Access denied") {
		err.Err = errors.Join(err.Err, ErrPermissionDenied)
	}

	return err
}

func UserErrorMessage(err *ServiceError) string {
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
