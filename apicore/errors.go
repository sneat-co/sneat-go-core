package apicore

import (
	"net/http"
)

type errWithResponseStatusCode struct {
	status int
	err    error
}

func (v errWithResponseStatusCode) Error() string {
	return v.err.Error()
}

func (v errWithResponseStatusCode) StatusCode() int {
	return v.status
}

func ErrWithStatusCode(status int, err error) error {
	return errWithResponseStatusCode{status: status, err: err}
}

var ErrUnauthorized = errWithResponseStatusCode{status: http.StatusUnauthorized}
