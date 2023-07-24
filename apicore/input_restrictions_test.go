package apicore

import (
	"bytes"
	"net/http"
	"testing"
)

func TestIsAllowedContentLength(t *testing.T) {
	req, err := http.NewRequest("POST", "/api2meetings/create-invite", new(bytes.Buffer))
	if err != nil {
		t.Fatal(err)
	}

	if err := validateContentLength(req, 0, 0); err != nil {
		t.Errorf("Content-Length=0, min=0, max=0, expected nil, got: %v", err)
	}
}
