package apicore

import (
	"net/http"

	"github.com/sneat-co/sneat-go-core/facade"
)

// badRequestError marks an error as facade.ErrBadRequest without changing its message,
// while keeping the original validation chain reachable via errors.Is / errors.As.
type badRequestError struct {
	err error
}

func (v badRequestError) Error() string {
	return v.err.Error()
}

func (v badRequestError) Unwrap() []error {
	return []error{v.err, facade.ErrBadRequest}
}

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
