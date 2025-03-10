package facade

import (
	"errors"
	"fmt"
)

// ErrUnauthenticated when userContext is not authenticated
var ErrUnauthenticated = errors.New("not authenticated")

// ErrUnauthorized when userContext have no access to requested resource/operation
var ErrUnauthorized = errors.New("unauthorized")

// ErrNoAuthHeader when auth header has not been provided
var ErrNoAuthHeader = fmt.Errorf("%w: authorization header is not provided", ErrUnauthorized)

// NewErrNoAuthHeader returns ErrNoAuthHeader with provided header name
func NewErrNoAuthHeader(headerName string) error {
	if headerName == "" {
		return ErrNoAuthHeader
	}
	return fmt.Errorf("%w: %s", ErrNoAuthHeader, headerName)
}
