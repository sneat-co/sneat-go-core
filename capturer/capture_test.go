package capturer

import (
	"context"
	"errors"
	"testing"
)

func TestCaptureError(t *testing.T) {
	ctx := context.Background()

	err := errors.New("test error")

	capturedErr := CaptureError(ctx, err)
	if capturedErr == nil {
		t.Errorf("CaptureError() should return an error")
	}
	if isCaptured, _ := IsCapturedError(capturedErr); !isCaptured {
		t.Errorf("CaptureError() should return a captured error")
	}
}
