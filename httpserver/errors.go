package httpserver

import (
	"errors"
	"fmt"
	"github.com/sneat-co/sneat-go/src/core/facade"
)

// ErrNotABearerToken is returned when provided auth header is not a valid bearer token
var ErrNotABearerToken = fmt.Errorf("%w: authorization header is not a bearer token", facade.ErrUnauthorized)

// IsUnauthorizedError returns true if an error is because of unauthorized request
func IsUnauthorizedError(err error) bool {
	return errors.Is(err, facade.ErrUnauthorized)
}
