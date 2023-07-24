package core

// Validatable defines an interface for any DB record
type Validatable interface {

	// Validate returns error if not valid
	Validate() error
}
