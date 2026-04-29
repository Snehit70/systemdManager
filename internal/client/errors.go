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
	Kind error
}

func (e *ServiceError) Error() string {
	if e == nil {
		return "unknown service error"
	}

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
	if e == nil {
		return nil
	}

	return e.Err
}

func (e *ServiceError) Is(target error) bool {
	if e == nil {
		return false
	}

	return errors.Is(e.Err, target) || errors.Is(e.Kind, target)
}

var (
	ErrNotFound         = errors.New("service not found")
	ErrPermissionDenied = errors.New("permission denied")
	ErrNotRunning       = errors.New("service not running")
	ErrNotEnabled       = errors.New("service not enabled")
	ErrSystemctlMissing = errors.New("systemctl not found")
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
	// Check for missing systemctl binary before generic "not found" patterns.
	if (strings.Contains(msg, "executable file not found") || strings.Contains(msg, "no such file or directory")) &&
		strings.Contains(msg, "systemctl") {
		err.Kind = errors.Join(err.Kind, ErrSystemctlMissing)
	} else if strings.Contains(msg, "service not found") || strings.Contains(msg, "could not be found") ||
		(strings.Contains(msg, "not found") && !strings.Contains(msg, "systemctl")) {
		err.Kind = errors.Join(err.Kind, ErrNotFound)
	}
	if strings.Contains(msg, "permission denied") || strings.Contains(msg, "access denied") {
		err.Kind = errors.Join(err.Kind, ErrPermissionDenied)
	}

	return err
}

func UserErrorMessage(err *ServiceError) string {
	if err == nil {
		return "Unknown service error"
	}

	switch {
	case errors.Is(err.Kind, ErrSystemctlMissing):
		return "systemctl not found in PATH — is systemd installed and available?"
	case errors.Is(err.Kind, ErrNotFound):
		return fmt.Sprintf("%s for unit '%s' failed: service not found", err.Op, err.Unit)
	case errors.Is(err.Kind, ErrPermissionDenied):
		return fmt.Sprintf("%s for unit '%s' failed: permission denied - check your user session and ACLs", err.Op, err.Unit)
	case errors.Is(err.Kind, ErrNotRunning):
		return fmt.Sprintf("%s for unit '%s' failed: service is not running", err.Op, err.Unit)
	case errors.Is(err.Kind, ErrNotEnabled):
		return fmt.Sprintf("%s for unit '%s' failed: service is not enabled", err.Op, err.Unit)
	default:
		return err.Error()
	}
}
