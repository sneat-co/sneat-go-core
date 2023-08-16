package monitoring

var captureException func(err error) Event

func SetExceptionCapturer(capture func(err error) Event) {
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
