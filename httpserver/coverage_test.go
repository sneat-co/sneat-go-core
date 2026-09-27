package httpserver

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sneat-co/sneat-go-core/capturer"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/strongo/validation"
)

type failWriter struct {
	http.ResponseWriter
}

func (f failWriter) Write([]byte) (int, error) {
	return 0, errors.New("write error")
}

func (f failWriter) Header() http.Header {
	return f.ResponseWriter.Header()
}

func (f failWriter) WriteHeader(statusCode int) {
	f.ResponseWriter.WriteHeader(statusCode)
}

type customWrapError struct {
	inner error
}

func (c customWrapError) Error() string { return "custom: " + c.inner.Error() }
func (c customWrapError) Unwrap() error { return c.inner }

func TestAccessControlAllowOrigin(t *testing.T) {
	t.Run("nil_request", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic on nil request")
			}
		}()
		AccessControlAllowOrigin(httptest.NewRecorder(), nil)
	})

	t.Run("nil_writer", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic on nil writer")
			}
		}()
		AccessControlAllowOrigin(nil, httptest.NewRequest(http.MethodGet, "/", nil))
	})

	t.Run("invalid_origin", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Origin", "ftp://evil.com")
		if AccessControlAllowOrigin(w, r) {
			t.Fatal("expected false for invalid origin")
		}
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})

	t.Run("valid_origin", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Origin", "http://localhost:8100")
		if !AccessControlAllowOrigin(w, r) {
			t.Fatal("expected true for valid origin")
		}
		if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:8100" {
			t.Fatalf("unexpected allow origin header: %s", w.Header().Get("Access-Control-Allow-Origin"))
		}

		// When already present
		if !AccessControlAllowOrigin(w, r) {
			t.Fatal("expected true when header already set")
		}
	})
}

func TestErrors(t *testing.T) {
	if !IsUnauthorizedError(ErrNotABearerToken) {
		t.Fatal("expected ErrNotABearerToken to be unauthorized")
	}
	if IsUnauthorizedError(errors.New("other")) {
		t.Fatal("expected other error not to be unauthorized")
	}
	if !IsForbiddenError(facade.ErrForbidden) {
		t.Fatal("expected ErrForbidden to be forbidden")
	}
	if IsForbiddenError(errors.New("other")) {
		t.Fatal("expected other error not to be forbidden")
	}
}

func TestErrorDetails_String(t *testing.T) {
	e := errorDetails{From: "auth", Type: "errorString", Message: "failed"}
	if e.String() == "" {
		t.Fatal("expected non-empty string")
	}
}

func TestHandleError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	t.Run("nil_context", func(t *testing.T) {
		w := httptest.NewRecorder()
		HandleError(nil, errors.New("test error"), "test", w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("captured_error", func(t *testing.T) {
		w := httptest.NewRecorder()
		captured := capturer.CaptureError(context.Background(), errors.New("captured"))
		HandleError(context.Background(), captured, "test", w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("bad_request", func(t *testing.T) {
		w := httptest.NewRecorder()
		err := validation.NewErrBadRequestFieldValue("field", "bad")
		HandleError(context.Background(), err, "test", w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		w := httptest.NewRecorder()
		HandleError(context.Background(), ErrNotABearerToken, "test", w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		w := httptest.NewRecorder()
		HandleError(context.Background(), facade.ErrForbidden, "test", w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})

	t.Run("writer_error", func(t *testing.T) {
		rec := httptest.NewRecorder()
		fw := failWriter{ResponseWriter: rec}
		HandleError(context.Background(), errors.New("err"), "test", fw, req)
	})

	t.Run("nil_test_v_flag", func(t *testing.T) {
		orig := flag.CommandLine
		defer func() { flag.CommandLine = orig }()
		flag.CommandLine = flag.NewFlagSet("custom", flag.ContinueOnError)

		w := httptest.NewRecorder()
		HandleError(context.Background(), errors.New("logged error"), "test", w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})
}

func TestGetErrorTypes_EdgeCases(t *testing.T) {
	// fmt.wrapError wrapping nil
	errWrapNil := fmt.Errorf("wrap nil: %w", nil)
	t1, r1 := getErrorTypes(errWrapNil)
	if t1 == "" {
		t.Fatalf("unexpected t1: %s, r1: %s", t1, r1)
	}

	// custom wrapped error where root is non-errorString
	wrappedCustom := customWrapError{inner: validation.NewErrRecordIsMissingRequiredField("field")}
	doubleWrapped := fmt.Errorf("outer: %w", wrappedCustom)
	_, root := getErrorTypes(doubleWrapped)
	if root == "" {
		t.Fatal("expected non-empty root error type for custom wrapped error")
	}
}
