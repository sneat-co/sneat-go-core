package apicore

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/sneat-co/sneat-go-core/apicore/verify"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/sneatauth"
)

type validatableRequest struct {
	Name string `json:"name"`
	fail bool
}

func (v *validatableRequest) Validate() error {
	if v.fail {
		return errors.New("validation failed")
	}
	return nil
}

type validatableResponse struct {
	Value string `json:"value"`
	fail  bool
}

func (v validatableResponse) Validate() error {
	if v.fail {
		return errors.New("response validation failed")
	}
	return nil
}

func TestErrors(t *testing.T) {
	err := ErrWithStatusCode(http.StatusNotFound, errors.New("not found"))
	if err.Error() != "not found" {
		t.Fatalf("unexpected error message: %s", err.Error())
	}
	if sc, ok := err.(interface{ StatusCode() int }); !ok || sc.StatusCode() != http.StatusNotFound {
		t.Fatalf("unexpected status code: %v", err)
	}
	if ErrUnauthorized.StatusCode() != http.StatusUnauthorized {
		t.Fatalf("unexpected ErrUnauthorized status: %d", ErrUnauthorized.StatusCode())
	}
}

func TestResponse_Helpers(t *testing.T) {
	ctx := context.Background()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// IfNoErrorReturnOK
	w := httptest.NewRecorder()
	IfNoErrorReturnOK(ctx, w, req, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	w = httptest.NewRecorder()
	IfNoErrorReturnOK(ctx, w, req, errors.New("err"))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	// IfNoErrorReturnCreatedOK
	w = httptest.NewRecorder()
	IfNoErrorReturnCreatedOK(ctx, w, req, nil)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	w = httptest.NewRecorder()
	IfNoErrorReturnCreatedOK(ctx, w, req, errors.New("err"))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	// ReturnStatus
	w = httptest.NewRecorder()
	ReturnStatus(ctx, w, req, http.StatusOK, nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for 200, got %d", w.Code)
	}
	w = httptest.NewRecorder()
	ReturnStatus(ctx, w, req, http.StatusAccepted, nil)
	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", w.Code)
	}
	w = httptest.NewRecorder()
	ReturnStatus(ctx, w, req, http.StatusOK, errors.New("err"))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 on err, got %d", w.Code)
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic on statusCode >= 400 with nil err")
			}
		}()
		ReturnStatus(ctx, httptest.NewRecorder(), req, http.StatusBadRequest, nil)
	}()
}

func TestReturnJSON_Comprehensive(t *testing.T) {
	ctx := context.Background()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Unauthorized error
	w := httptest.NewRecorder()
	ReturnJSON(ctx, w, req, http.StatusOK, facade.ErrUnauthorized, nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}

	// Validatable response error
	w = httptest.NewRecorder()
	ReturnJSON(ctx, w, req, http.StatusOK, nil, validatableResponse{fail: true})
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 on invalid response, got %d", w.Code)
	}

	// Validatable response success
	w = httptest.NewRecorder()
	ReturnJSON(ctx, w, req, http.StatusOK, nil, validatableResponse{Value: "ok"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// 204 No Content with nil response
	w = httptest.NewRecorder()
	ReturnJSON(ctx, w, req, http.StatusNoContent, nil, nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	// 204 No Content with non-nil response panics
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic on 204 with non-nil response")
			}
		}()
		ReturnJSON(ctx, httptest.NewRecorder(), req, http.StatusNoContent, nil, "non-nil")
	}()

	// 200 OK with nil response panics
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic on 200 with nil response")
			}
		}()
		ReturnJSON(ctx, httptest.NewRecorder(), req, http.StatusOK, nil, nil)
	}()

	// JSON encode error
	w = httptest.NewRecorder()
	ReturnJSON(ctx, w, req, http.StatusOK, nil, make(chan int))
	if !strings.Contains(w.Body.String(), "Failed to encode response") {
		t.Fatalf("expected encode failure message in body, got: %s", w.Body.String())
	}
}

func TestVerifyRequestAndCreateUserContext(t *testing.T) {
	origGetAuthToken := GetAuthTokenFromHttpRequest
	defer func() { GetAuthTokenFromHttpRequest = origGetAuthToken }()

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://localhost/", nil)

	// Panics
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic on nil request")
			}
		}()
		_, _ = VerifyRequestAndCreateUserContext(w, nil, verify.Request())
	}()

	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic on nil writer")
			}
		}()
		_, _ = VerifyRequestAndCreateUserContext(nil, r, verify.Request())
	}()

	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic on nil options")
			}
		}()
		_, _ = VerifyRequestAndCreateUserContext(w, r, nil)
	}()

	// VerifyRequest panics when GetAuthTokenFromHttpRequest is nil
	GetAuthTokenFromHttpRequest = nil
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic on nil GetAuthTokenFromHttpRequest")
			}
		}()
		_, _ = VerifyRequest(w, r, verify.Request())
	}()

	// VerifyRequest fails when token getter returns error
	GetAuthTokenFromHttpRequest = func(r *http.Request, authRequired bool) (*sneatauth.Token, error) {
		return nil, errors.New("auth provider error")
	}
	_, err := VerifyRequest(w, r, verify.Request())
	if err == nil {
		t.Fatal("expected error from auth provider failure")
	}

	// VerifyRequest fails when authRequired and token is nil
	GetAuthTokenFromHttpRequest = func(r *http.Request, authRequired bool) (*sneatauth.Token, error) {
		return nil, nil
	}
	_, err = VerifyRequest(w, r, verify.Request(verify.AuthenticationRequired(true)))
	if err == nil {
		t.Fatal("expected error when auth required and token is nil")
	}

	// VerifyRequest invalid origin
	rInvalidOrigin := httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
	rInvalidOrigin.Header.Set("Origin", "ftp://evil.com")
	_, err = VerifyRequest(w, rInvalidOrigin, verify.Request())
	if err == nil {
		t.Fatal("expected error on invalid origin")
	}

	// VerifyRequest content length error
	rBadLen := httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
	rBadLen.ContentLength = 100
	_, err = VerifyRequest(w, rBadLen, verify.Request(verify.MaximumContentLength(50)))
	if err == nil {
		t.Fatal("expected error on max content length exceeded")
	}

	// VerifyRequestAndCreateUserContext success with token
	GetAuthTokenFromHttpRequest = func(r *http.Request, authRequired bool) (*sneatauth.Token, error) {
		return &sneatauth.Token{UID: "user123"}, nil
	}
	ctxWithUser, err := VerifyRequestAndCreateUserContext(w, r, verify.Request())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctxWithUser == nil || ctxWithUser.User().GetUserID() != "user123" {
		t.Fatal("expected user context with user123")
	}

	// VerifyRequestAndCreateUserContext fails on bad origin
	w = httptest.NewRecorder()
	_, err = VerifyRequestAndCreateUserContext(w, rInvalidOrigin, verify.Request())
	if err == nil {
		t.Fatal("expected error on bad origin in VerifyRequestAndCreateUserContext")
	}
}

func TestDecodeRequestBody(t *testing.T) {
	w := httptest.NewRecorder()

	// Unsupported method
	rGet := httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
	if err := DecodeRequestBody(w, rGet, &validatableRequest{}); err == nil {
		t.Fatal("expected error for GET method in DecodeRequestBody")
	}

	// Valid POST with localhost logging
	bodyJSON := []byte(`{"name":"test"}`)
	rPost := httptest.NewRequest(http.MethodPost, "http://localhost/", bytes.NewReader(bodyJSON))
	rPost.ContentLength = int64(len(bodyJSON))
	reqObj := &validatableRequest{}
	if err := DecodeRequestBody(w, rPost, reqObj); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reqObj.Name != "test" {
		t.Fatalf("expected name=test, got %s", reqObj.Name)
	}

	// Read error
	rReadErr := httptest.NewRequest(http.MethodPost, "http://localhost/", iotest.ErrReader(errors.New("read error")))
	rReadErr.ContentLength = 10
	if err := DecodeRequestBody(w, rReadErr, &validatableRequest{}); err == nil {
		t.Fatal("expected error on read error")
	}

	// Bad JSON
	badJSON := []byte(`{invalid`)
	rBad := httptest.NewRequest(http.MethodPost, "http://localhost/", bytes.NewReader(badJSON))
	rBad.ContentLength = int64(len(badJSON))
	if err := DecodeRequestBody(w, rBad, &validatableRequest{}); err == nil {
		t.Fatal("expected error on invalid JSON")
	}

	// Validation fails
	rFailVal := httptest.NewRequest(http.MethodPost, "http://localhost/", bytes.NewReader(bodyJSON))
	rFailVal.ContentLength = int64(len(bodyJSON))
	if err := DecodeRequestBody(w, rFailVal, &validatableRequest{fail: true}); err == nil {
		t.Fatal("expected error on validation failure")
	}
}

func TestVerifyAuthenticatedRequestAndDecodeBody(t *testing.T) {
	origGetAuthToken := GetAuthTokenFromHttpRequest
	defer func() { GetAuthTokenFromHttpRequest = origGetAuthToken }()

	GetAuthTokenFromHttpRequest = func(r *http.Request, authRequired bool) (*sneatauth.Token, error) {
		return &sneatauth.Token{UID: "user123"}, nil
	}

	w := httptest.NewRecorder()
	bodyJSON := []byte(`{"name":"test"}`)
	rPost := httptest.NewRequest(http.MethodPost, "http://localhost/", bytes.NewReader(bodyJSON))
	rPost.ContentLength = int64(len(bodyJSON))

	// Success
	ctxWithUser, err := VerifyAuthenticatedRequestAndDecodeBody(w, rPost, verify.Request(verify.MaximumContentLength(1024)), &validatableRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctxWithUser == nil {
		t.Fatal("expected non-nil ctxWithUser")
	}

	// Verify fails
	rBadOrigin := httptest.NewRequest(http.MethodPost, "http://localhost/", nil)
	rBadOrigin.Header.Set("Origin", "ftp://evil.com")
	_, err = VerifyAuthenticatedRequestAndDecodeBody(w, rBadOrigin, verify.Request(verify.MaximumContentLength(1024)), &validatableRequest{})
	if err == nil {
		t.Fatal("expected error when verify fails")
	}

	// Decode fails
	rBadJSON := httptest.NewRequest(http.MethodPost, "http://localhost/", bytes.NewReader([]byte(`{`)))
	rBadJSON.ContentLength = 1
	_, err = VerifyAuthenticatedRequestAndDecodeBody(w, rBadJSON, verify.Request(verify.MaximumContentLength(1024)), &validatableRequest{})
	if err == nil {
		t.Fatal("expected error when decode fails")
	}
}

func TestExecute_Comprehensive(t *testing.T) {
	origGetAuthToken := GetAuthTokenFromHttpRequest
	defer func() { GetAuthTokenFromHttpRequest = origGetAuthToken }()

	GetAuthTokenFromHttpRequest = func(r *http.Request, authRequired bool) (*sneatauth.Token, error) {
		return &sneatauth.Token{UID: "user123"}, nil
	}

	// VerifyRequest fails
	w := httptest.NewRecorder()
	rBadOrigin := httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
	rBadOrigin.Header.Set("Origin", "ftp://evil.com")
	Execute(w, rBadOrigin, nil, verify.Request(), http.StatusOK, nil, nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}

	// DecodeRequestBody fails
	w = httptest.NewRecorder()
	rBadJSON := httptest.NewRequest(http.MethodPost, "http://localhost/", bytes.NewReader([]byte(`{`)))
	rBadJSON.ContentLength = 1
	Execute(w, rBadJSON, &validatableRequest{}, verify.Request(verify.MaximumContentLength(100)), http.StatusOK, nil, nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	// getContext fails
	w = httptest.NewRecorder()
	rGet := httptest.NewRequest(http.MethodGet, "http://localhost/", nil)
	getContextErr := func(r *http.Request) (context.Context, error) {
		return nil, errors.New("get context error")
	}
	Execute(w, rGet, nil, verify.Request(), http.StatusOK, getContextErr, nil)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	// Success
	w = httptest.NewRecorder()
	getContextOK := func(r *http.Request) (context.Context, error) {
		return r.Context(), nil
	}
	handlerOK := func(ctx context.Context) (ResponseDTO, error) {
		return validatableResponse{Value: "done"}, nil
	}
	Execute(w, rGet, nil, verify.Request(), http.StatusOK, getContextOK, handlerOK)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
