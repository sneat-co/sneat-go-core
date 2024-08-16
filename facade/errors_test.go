package facade

import "testing"

func TestErrForbidden(t *testing.T) {
	err := ErrForbidden
	if err.Error() != "forbidden" {
		t.Error("Expected error message to be 'forbidden'")
	}
}

func TestErrBadRequest(t *testing.T) {
	err := ErrBadRequest
	if err.Error() != "bad request" {
		t.Error("Expected error message to be 'bad request'")
	}
}

func TestErrNotInitialized(t *testing.T) {
	err := ErrNotInitialized
	if err.Error() != "not initialized" {
		t.Error("Expected error message to be 'not initialized'")
	}
}
