package monitoring

import (
	"context"
	"fmt"
	"github.com/sneat-co/sneat-go-core/capturer"
	"github.com/strongo/logus"
)

// TODO: Do we really need both ErrorCapturer & PanicCapturer? If so document intended usage and a use case example.

type ErrorCapturer func(ctx context.Context, err error) Event

type PanicCapturer func(ctx context.Context, v any) Event

var captureError ErrorCapturer
var capturePanic PanicCapturer

// SetErrorCapturer sets a function that will be called to capture an error.
func SetErrorCapturer(capturer ErrorCapturer) {
	if capturer == nil {
		panic("func SetErrorCapturer() should not be called with nil `capturer` argument")
	}
	captureError = capturer
}

// SetPanicCapturer sets a function that will be called to capture a panic.
func SetPanicCapturer(capturer PanicCapturer) {
	if capturer == nil {
		panic("func SetPanicCapturer() should not be called with nil `capturer` argument")
	}
	capturePanic = capturer
}

// Event represents an event captured my monitoring subsystem.
type Event struct {
	ID string
}

// CaptureError captures error and returns event ID.
func CaptureError(ctx context.Context, err error) Event {
	if captureError == nil {
		logus.Warningf(ctx, "Exception capturer is not set. Call monitoring.SetErrorCapturer(capturer ErrorCapturer) in you app initialization code")

		captureError = func(ctx context.Context, err error) Event {
			logus.Errorf(ctx, err.Error())
			return Event{}
		}
	}
	isCapturedErr, capturedErr := capturer.IsCapturedError(err)
	if isCapturedErr {
		err = capturedErr
	}
	return captureError(ctx, err)
}

// CapturePanic captures panic and returns event ID.
func CapturePanic(ctx context.Context, err any) Event {
	if capturePanic == nil {
		logus.Warningf(ctx, "Panic capturer is not set. Call monitoring.SetPanicCapturer(capturer PanicCapturer) in you app initialization code")

		capturePanic = func(ctx context.Context, v any) Event {
			logus.Errorf(ctx, fmt.Sprintf("PANIC: %v", v))
			return Event{}
		}
	}
	return capturePanic(ctx, err)
}
