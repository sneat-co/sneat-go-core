package monitoring

import (
	"context"
	"github.com/sneat-co/sneat-go-core/capturer"
	"log"
)

var captureException func(ctx context.Context, err error) Event

// SetExceptionCapturer sets a function that will be called to capture exception.
func SetExceptionCapturer(capture func(ctx context.Context, err error) Event) {
	if capture == nil {
		panic("func SetExceptionCapturer() should not be called with nil `capture` argument")
	}
	captureException = capture
}

// Event represents an event captured my monitoring subsystem.
type Event struct {
	ID string
}

// CaptureException captures exception and returns event ID.
func CaptureException(ctx context.Context, err error) Event {
	if captureException == nil {
		log.Println("WARNING:", "Exception capturer is not set. Call monitoring.SetExceptionCapturer(capture func(err error)) in you app initialization code")

		captureException = func(ctx context.Context, err error) Event {
			log.Println("ERROR:", err)
			return Event{}
		}
	}
	isCapturedErr, capturedErr := capturer.IsCapturedError(err)
	if isCapturedErr {
		err = capturedErr
	}
	return captureException(ctx, err)
}
