package monitoring

var captureException func(err error) Event

// SetExceptionCapturer sets a function that will be called to capture exception.
func SetExceptionCapturer(capture func(err error) Event) {
	if capture == nil {
		panic("SetExceptionCapturer() should not be called with nil")
	}
	captureException = capture
}

type Event struct {
	ID string
}

func CaptureException(err error) Event {
	if captureException == nil {
		panic("exception capturer is not set, call monitoring.SetExceptionCapturer(capture func(err error)) in you app initialization code")
	}
	return captureException(err)
}
