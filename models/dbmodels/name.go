package dbmodels

import (
	"fmt"
	"github.com/strongo/validation"
	"strings"
)

// Name defines names of a person (like First, Last, Middle)
type Name struct {
	Full   string `json:"full,omitempty" firestore:"full,omitempty"`
	First  string `json:"first,omitempty" firestore:"first,omitempty"`
	Middle string `json:"middle,omitempty" firestore:"middle,omitempty"`
	Last   string `json:"last,omitempty" firestore:"last,omitempty"`
	Nick   string `json:"nick,omitempty" firestore:"nick,omitempty"`
}

// Equal returns true if two Name structs are equal
func (v *Name) Equal(v2 *Name) bool {
	return v == nil && v2 == nil || v != nil && v2 != nil && *v == *v2
}

// Title returns full name
func (v *Name) Title() string {
	if v.Full != "" {
		return v.Full
	}
	if v.First != "" && v.Last != "" {
		if v.Middle == "" {
			return fmt.Sprintf("%v %v", v.First, v.Last)
		}
		return fmt.Sprintf("%v %v %v", v.First, v.Middle, v.Last)
	}
	if v.First != "" {
		return v.First
	}
	if v.Last != "" {
		return v.Last
	}
	if v.Middle != "" {
		return v.Middle
	}
	return ""
}

// IsEmpty checks if name is empty
func (v *Name) IsEmpty() bool {
	return v == nil || *v == Name{}
}

// Validate returns error if not valid
func (v *Name) Validate() error {
	if v == nil {
		return nil
	}
	const spaces = "leading or closing spaces"
	if strings.TrimSpace(v.First) != v.First {
		return validation.NewErrBadRecordFieldValue("first", spaces)
	}
	if strings.TrimSpace(v.Last) != v.Last {
		return validation.NewErrBadRecordFieldValue("last", spaces)
	}
	if strings.TrimSpace(v.Middle) != v.Middle {
		return validation.NewErrBadRecordFieldValue("middle", spaces)
	}
	if strings.TrimSpace(v.Full) != v.Full {
		return validation.NewErrBadRecordFieldValue("full", spaces)
	}
	if err := ValidateRequiredName(v); err != nil {
		return err
	}
	return nil
}