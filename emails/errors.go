package emails

import "fmt"

// SendEmailError error and message
type SendEmailError struct {
	msg string
	err error
}

// Error returns error
func (v *SendEmailError) Error() string {
	return fmt.Sprintf("%v: %v", v.msg, v.err)
}

// Unwrap returns original error
func (v *SendEmailError) Unwrap() error {
	return v.err
}

// NewSendEmailError wraps a send email error
func NewSendEmailError(msg string, err error) error {
	return &SendEmailError{msg: msg, err: err}
}
