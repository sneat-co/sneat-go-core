package monitoring

import (
	"context"
	"errors"
	"testing"

	"github.com/sneat-co/sneat-go-core/capturer"
)

func TestSetExceptionCapturer(t *testing.T) {
	origError := captureError
	origPanic := capturePanic
	defer func() {
		captureError = origError
		capturePanic = origPanic
	}()

	t.Run("nil_error_capturer", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("SetErrorCapturer() should panic")
			}
		}()
		SetErrorCapturer(nil)
	})

	t.Run("valid_error_capturer", func(t *testing.T) {
		called := false
		SetErrorCapturer(func(ctx context.Context, err error) Event {
			called = true
			return Event{ID: "err1"}
		})
		ev := CaptureError(context.Background(), errors.New("test error"))
		if !called || ev.ID != "err1" {
			t.Fatalf("unexpected result: called=%v, ev=%+v", called, ev)
		}
	})

	t.Run("nil_panic_capturer", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("SetPanicCapturer() should panic")
			}
		}()
		SetPanicCapturer(nil)
	})

	t.Run("valid_panic_capturer", func(t *testing.T) {
		called := false
		SetPanicCapturer(func(ctx context.Context, v any) Event {
			called = true
			return Event{ID: "panic1"}
		})
		ev := CapturePanic(context.Background(), "panic value")
		if !called || ev.ID != "panic1" {
			t.Fatalf("unexpected result: called=%v, ev=%+v", called, ev)
		}
	})

	t.Run("default_error_capturer_and_captured_error", func(t *testing.T) {
		captureError = nil
		ctx := context.Background()
		baseErr := errors.New("underlying error")
		wrappedErr := capturer.CaptureError(ctx, baseErr)
		ev := CaptureError(ctx, wrappedErr)
		if ev.ID != "" {
			t.Fatalf("expected empty event ID from default capturer, got %v", ev.ID)
		}
	})

	t.Run("default_panic_capturer", func(t *testing.T) {
		capturePanic = nil
		ctx := context.Background()
		ev := CapturePanic(ctx, "something went wrong")
		if ev.ID != "" {
			t.Fatalf("expected empty event ID from default panic capturer, got %v", ev.ID)
		}
	})
}
