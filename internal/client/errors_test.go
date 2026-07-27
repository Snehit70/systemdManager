package client

import (
	"errors"
	"testing"
)

func TestServiceErrorIsOnlyMatchesWrappedSentinel(t *testing.T) {
	err := &ServiceError{
		Op:   "start",
		Unit: "demo.service",
		Err:  errors.New("generic failure"),
	}

	for _, sentinel := range []error{ErrNotFound, ErrPermissionDenied, ErrNotRunning, ErrNotEnabled} {
		if errors.Is(err, sentinel) {
			t.Fatalf("generic service error should not match %v", sentinel)
		}
	}
}

func TestServiceErrorIsMatchesJoinedSentinel(t *testing.T) {
	err := &ServiceError{
		Op:   "start",
		Unit: "demo.service",
		Err:  errors.Join(errors.New("systemctl failed"), ErrNotFound),
	}

	if !errors.Is(err, ErrNotFound) {
		t.Fatal("service error should match joined ErrNotFound sentinel")
	}
	if errors.Is(err, ErrPermissionDenied) {
		t.Fatal("not found service error should not match ErrPermissionDenied")
	}
}

func TestNewServiceErrorClassifiesOperationalState(t *testing.T) {
	tests := []struct {
		name        string
		op          string
		message     string
		wantKind    error
		wantMessage string
	}{
		{
			name:        "not running",
			op:          "stop",
			message:     "Unit demo.service is not running.",
			wantKind:    ErrNotRunning,
			wantMessage: "stop for unit 'demo.service' failed: service is not running",
		},
		{
			name:        "not active",
			op:          "restart",
			message:     "Unit demo.service is not active, cannot reload.",
			wantKind:    ErrNotRunning,
			wantMessage: "restart for unit 'demo.service' failed: service is not running",
		},
		{
			name:        "not enabled",
			op:          "disable",
			message:     "Unit file demo.service is not enabled.",
			wantKind:    ErrNotEnabled,
			wantMessage: "disable for unit 'demo.service' failed: service is not enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewServiceError(tt.op, "demo.service", errors.New(tt.message))
			if !errors.Is(err, tt.wantKind) {
				t.Fatalf("expected %v classification, got %v", tt.wantKind, err.Kind)
			}
			if got := UserErrorMessage(err); got != tt.wantMessage {
				t.Fatalf("unexpected user message: %q", got)
			}
		})
	}
}
