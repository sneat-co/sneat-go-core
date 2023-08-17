package apicore

import (
	"bytes"
	"context"
	"github.com/sneat-co/sneat-go-core/apicore/verify"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsAllowedContentLength(t *testing.T) {
	req, err := http.NewRequest("POST", "/api4meetingus/create-invite", new(bytes.Buffer))
	if err != nil {
		t.Fatal(err)
	}

	if err := validateContentLength(req, 0, 0); err != nil {
		t.Errorf("Content-Length=0, min=0, max=0, expected nil, got: %v", err)
	}
}

func TestVerifyRequest(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://localhost/", nil)
	o := verify.Request()
	if o == nil {
		t.Fatal("expected options, got nil")
	}
	NewContextWithToken = func(r *http.Request, authRequired bool) (ctx context.Context, err error) {
		return context.Background(), nil
	}
	ctx, err := VerifyRequest(w, r, o)
	if ctx == nil {
		t.Errorf("expected context, got nil")
	}
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}
