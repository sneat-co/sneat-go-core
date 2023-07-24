package capturer

type capturedError struct {
	error
}

func (v *capturedError) Error() string {
	return "captured error"
}

func (v *capturedError) Unwrap() error {
	return v.error
}
