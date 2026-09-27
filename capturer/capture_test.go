package capturer

import (
	"context"
	"errors"
	"testing"
)

func TestCaptureError(t *testing.T) {
	ctx := context.Background()

	mockLogger := NoOpErrorLogger{}
	origLoggers := loggers
	defer func() { loggers = origLoggers }()
	loggers = []ErrorLogger{mockLogger}

	err := errors.New("test error")

	capturedErr := CaptureError(ctx, err)
	if capturedErr == nil {
		t.Errorf("CaptureError() should return an error")
	}
	if isCaptured, _ := IsCapturedError(capturedErr); !isCaptured {
		t.Errorf("CaptureError() should return a captured error")
	}

	plainErr := errors.New("not captured")
	if isCaptured, unwrapped := IsCapturedError(plainErr); isCaptured || unwrapped != plainErr {
		t.Errorf("IsCapturedError(plainErr) = (%v, %v), want (false, plainErr)", isCaptured, unwrapped)
	}
}
