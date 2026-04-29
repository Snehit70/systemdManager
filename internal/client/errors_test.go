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

	if errors.Is(err, ErrNotFound) {
		t.Fatal("generic service error should not match ErrNotFound")
	}
	if errors.Is(err, ErrPermissionDenied) {
		t.Fatal("generic service error should not match ErrPermissionDenied")
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
