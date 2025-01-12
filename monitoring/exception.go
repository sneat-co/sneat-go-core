package monitoring

import (
	"context"
	"github.com/sneat-co/sneat-go-core/capturer"
	"github.com/strongo/logus"
)

type ExceptionCapturer func(ctx context.Context, err error) Event

var captureException ExceptionCapturer

// SetExceptionCapturer sets a function that will be called to capture exception.
func SetExceptionCapturer(capturer ExceptionCapturer) {
	if capturer == nil {
		panic("func SetExceptionCapturer() should not be called with nil `capturer` argument")
	}
	captureException = capturer
}

// Event represents an event captured my monitoring subsystem.
type Event struct {
	ID string
}

// CaptureException captures exception and returns event ID.
func CaptureException(ctx context.Context, err error) Event {
	if captureException == nil {
		logus.Warningf(ctx, "Exception capturer is not set. Call monitoring.SetExceptionCapturer(capture func(err error)) in you app initialization code")

		captureException = func(ctx context.Context, err error) Event {
			logus.Errorf(ctx, err.Error())
			return Event{}
		}
	}
	isCapturedErr, capturedErr := capturer.IsCapturedError(err)
	if isCapturedErr {
		err = capturedErr
	}
	return captureException(ctx, err)
}
