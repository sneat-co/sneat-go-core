package facade

import (
	"errors"
	"fmt"
)

// ErrUnauthenticated when AuthUser is not authenticated
var ErrUnauthenticated = errors.New("not authenticated")

// ErrUnauthorized when AuthUser have no access to requested resource/operation
var ErrUnauthorized = errors.New("unauthorized")

// ErrNoAuthHeader when auth header has not been provided
var ErrNoAuthHeader = fmt.Errorf("%w: authorization is not provided", ErrUnauthorized)
