package apicore

import (
	"bytes"
	"github.com/sneat-co/sneat-go-core/apicore/verify"
	"github.com/sneat-co/sneat-go-core/sneatauth"
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
	defer func() {
		GetAuthTokenFromHttpRequest = nil
	}()
	GetAuthTokenFromHttpRequest = func(r *http.Request) (token *sneatauth.Token, err error) {
		return nil, nil
	}

	ctx, err := VerifyRequest(w, r, o)
	if ctx == nil {
		t.Errorf("expected context, got nil")
	}
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}
