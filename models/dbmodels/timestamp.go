package dbmodels

import (
	"fmt"
	"github.com/sneat-co/sneat-go-core"
	"github.com/strongo/validation"
	"strings"
	"time"
)

var _ core.Validatable = (*ByUser)(nil)

// Timestamp record
type Timestamp struct {
	Time      time.Time `json:"t" firestore:"t"`
	Operation string    `json:"o" firestore:"o"` // e.g. start, paus, resum, finish.
	By        *ByUser   `json:"by,omitempty" firestore:"by,omitempty"`
}

// Validate validates Timestamp record
func (v *Timestamp) Validate() error {
	if v.Time.IsZero() {
		return validation.NewErrRecordIsMissingRequiredField("time")
	}
	if strings.TrimSpace(v.Operation) == "" {
		return validation.NewErrRecordIsMissingRequiredField("operation")
	}
	if v.By != nil {
		if err := v.By.Validate(); err != nil {
			return fmt.Errorf("invalid 'by' field: %w", err)
		}
	}
	return nil
}
