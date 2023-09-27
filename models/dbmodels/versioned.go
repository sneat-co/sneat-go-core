package dbmodels

import (
	"github.com/sneat-co/sneat-go-core"
	"time"
)

// Versioned defines an interface for versioned record
type Versioned interface {
	core.Validatable

	// IncreaseVersion returns new record version increased by 1. It also should update UpdatedAt and UpdatedBy fields.
	IncreaseVersion(updatedAt time.Time, updatedBy string) int
}

type WithUpdatedAndVersion struct {
	WithUpdated
	Version int `json:"v" firestore:"v"`
}

func (v *WithUpdatedAndVersion) IncreaseVersion(updatedAt time.Time, updatedBy string) int {
	v.Version++
	v.UpdatedAt = updatedAt
	v.UpdatedBy = updatedBy
	return v.Version
}
