package facade

import (
	"github.com/strongo/validation"
)

// Request interface
type Request interface {
	Validate() error
}

var _ Request = (*IDRequest)(nil)

// IDRequest holds a string ID
type IDRequest struct {
	ID string
}

// Validate validates a request
func (v IDRequest) Validate() error { // TODO(StackOverflow): Is it better to have pointer method here?
	if v.ID == "" {
		return validation.NewErrRecordIsMissingRequiredField("ContactID")
	}
	return nil
}
