package core

// Validatable defines an interface for a struct that can be validated
type Validatable interface {

	// Validate returns error if not valid
	Validate() error
}
