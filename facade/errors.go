package facade

import "errors"

// ErrBadRequest an error for bad request
var ErrBadRequest = errors.New("bad request")

// ErrForbidden an error for forbidden operations
var ErrForbidden = errors.New("forbidden")

// ErrNotInitialized an error for not initialized situations
var ErrNotInitialized = errors.New("not initialized")
