package monitoring

import "log"

var captureException func(err error) Event

// SetExceptionCapturer sets a function that will be called to capture exception.
func SetExceptionCapturer(capture func(err error) Event) {
	if capture == nil {
		panic("SetExceptionCapturer() should not be called with nil")
	}
	captureException = capture
}

// Event represents an event that was captured my monitoring subsystem.
type Event struct {
	ID string
}

// CaptureException captures exception and returns event ID.
func CaptureException(err error) Event {
	if captureException == nil {
		log.Println("WARNING:", "Exception capturer is not set. Call monitoring.SetExceptionCapturer(capture func(err error)) in you app initialization code")

		captureException = func(err error) Event {
			log.Println("ERROR:", err)
			return Event{}
		}
	}

	return captureException(err)
}
