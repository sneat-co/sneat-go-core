package httpserver

import (
	"errors"
	"fmt"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/strongo/validation"
)

// ErrNotABearerToken is returned when provided auth header is not a valid bearer token
var ErrNotABearerToken = fmt.Errorf("%w: authorization header is not a bearer token", facade.ErrUnauthorized)

// IsUnauthorizedError returns true if an error is because of unauthorized request
func IsUnauthorizedError(err error) bool {
	return errors.Is(err, facade.ErrUnauthorized)
}

// IsForbiddenError returns true when the caller is authenticated but does not
// have authority for the requested operation.
func IsForbiddenError(err error) bool {
	return errors.Is(err, facade.ErrForbidden)
}

// IsBadRequestError returns true when the error means the caller sent an invalid request: either a
// validation "bad request" error (e.g. validation.NewErrBadRequestFieldValue) or facade.ErrBadRequest,
// matched through any depth of wrapping.
func IsBadRequestError(err error) bool {
	return validation.IsBadRequestError(err) || errors.Is(err, facade.ErrBadRequest)
}
