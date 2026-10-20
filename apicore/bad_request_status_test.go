package apicore

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/models/dbmodels"
	"github.com/strongo/validation"
)

// invalidAuthProviderRequest reproduces the `init_user_record` request: its Validate() returns a
// "bad record field value" error for authProvider, which is NOT a validation "bad request" error.
type invalidAuthProviderRequest struct {
	AuthProvider string `json:"authProvider"`
}

func (v invalidAuthProviderRequest) Validate() error {
	return dbmodels.PersonEmail{Type: "personal", Address: "a@example.com", AuthProvider: v.AuthProvider}.Validate()
}

type errRequest struct{ err error }

func (v *errRequest) Validate() error { return v.err }

func TestDecodeRequestBody_ValidationFailureIs400(t *testing.T) {
	body := []byte(`{"authProvider":"no-such-provider"}`)

	tests := []struct {
		name string
		req  facade.Request
	}{
		{"bad_record_field_value_authProvider", &invalidAuthProviderRequest{}},
		{"bad_request_field_value", &errRequest{validation.NewErrBadRequestFieldValue("f", "bad")}},
		{"plain_error", &errRequest{errors.New("plain")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "http://localhost/", bytes.NewReader(body))
			r.ContentLength = int64(len(body))
			w := httptest.NewRecorder()
			err := DecodeRequestBody(w, r, tt.req)
			if err == nil {
				t.Fatal("expected error")
			}
			if w.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400; body: %s", w.Code, w.Body.String())
			}
			if !errors.Is(err, facade.ErrBadRequest) {
				t.Errorf("error must wrap facade.ErrBadRequest: %v", err)
			}
			if !strings.Contains(err.Error(), "failed initial validation") {
				t.Errorf("message must be unchanged, got: %v", err)
			}
		})
	}
}

func TestDecodeRequestBody_KeepsValidationChain(t *testing.T) {
	body := []byte(`{"authProvider":"no-such-provider"}`)
	r := httptest.NewRequest(http.MethodPost, "http://localhost/", bytes.NewReader(body))
	r.ContentLength = int64(len(body))
	w := httptest.NewRecorder()
	err := DecodeRequestBody(w, r, &invalidAuthProviderRequest{})
	if !validation.IsBadRecordError(err) {
		t.Error("original validation chain must stay reachable")
	}
	var fieldErr validation.ErrBadFieldValue
	if !errors.As(err, &fieldErr) || fieldErr.Field != "authProvider" {
		t.Errorf("expected ErrBadFieldValue for authProvider via errors.As, got: %v", err)
	}
	var resp struct {
		Error struct {
			Message     string `json:"message"`
			From        string `json:"from"`
			Type        string `json:"type"`
			RootErrType string `json:"rootErrType"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not JSON: %v: %s", err, w.Body.String())
	}
	if resp.Error.From != "DecodeRequestBody" {
		t.Errorf("from = %q", resp.Error.From)
	}
	if resp.Error.Type != "validation.ErrBadFieldValue" {
		t.Errorf("type = %q, want validation.ErrBadFieldValue (unchanged from before the 400 fix)", resp.Error.Type)
	}
	if resp.Error.RootErrType != "" {
		t.Errorf("rootErrType = %q, want empty (unchanged)", resp.Error.RootErrType)
	}
	if !strings.HasPrefix(resp.Error.Message, "request struct decoded from HTTP request body failed initial validation *apicore.invalidAuthProviderRequest: bad value for field [authProvider]: ") {
		t.Errorf("message = %q", resp.Error.Message)
	}
}
