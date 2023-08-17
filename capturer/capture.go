package capturer

import (
	"context"
	"errors"
)

// CaptureError captures & logs an error
func CaptureError(ctx context.Context, err error) error {
	for _, logger := range loggers {
		logger.LogError(ctx, err)
	}
	return &capturedError{error: err}
}

// IsCapturedError checks if an error was already captured
func IsCapturedError(err error) (bool, error) {
	var e *capturedError
	if isCaptured := errors.As(err, &e); isCaptured {
		return isCaptured, e.Unwrap()
	}
	return false, err
}
