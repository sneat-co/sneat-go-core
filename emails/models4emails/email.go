package models4emails

import (
	"github.com/strongo/validation"
	"strings"
	"time"
)

// Email record
type Email struct {
	To      string
	From    string
	Created time.Time
	Queued  *time.Time
	Sent    *time.Time
	Type    string // E.g. 'personal-invite'
	Subject string
	Body    struct {
		Text string
		HTML string
	}
}

// Validate returns error if not valid
func (v Email) Validate() error {
	if strings.TrimSpace(v.To) == "" {
		return validation.NewErrRecordIsMissingRequiredField("to")
	}
	if strings.TrimSpace(v.From) == "" {
		return validation.NewErrRecordIsMissingRequiredField("from")
	}
	if strings.TrimSpace(v.Subject) == "" {
		return validation.NewErrRecordIsMissingRequiredField("subject")
	}
	if strings.TrimSpace(v.Type) == "" {
		return validation.NewErrRecordIsMissingRequiredField("type")
	}
	if v.Created.IsZero() {
		return validation.NewErrRecordIsMissingRequiredField("created")
	}
	if strings.TrimSpace(v.Body.Text) == "" && strings.TrimSpace(v.Body.HTML) == "" {
		return validation.NewErrRecordIsMissingRequiredField("body")
	}
	return nil
}
