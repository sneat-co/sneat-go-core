package dbmodels

import (
	"github.com/sneat-co/sneat-go-core"
	"time"
)

// Versioned defines an interface for versioned record
type Versioned interface {
	IncreaseVersion(timestamp time.Time) int
	core.Validatable
}
