package dbprofile

import "github.com/strongo/validation"

// Avatar record
type Avatar struct {
	Provider string `json:"provider" firestore:"provider"`
	URL      string `json:"url" firestore:"url"`
	Width    int    `json:"width,omitempty" firestore:"width,omitempty"`
	Height   int    `json:"height,omitempty" firestore:"height,omitempty"`
	Size     int    `json:"size,omitempty" firestore:"size,omitempty"`
}

func (v *Avatar) Equal(v2 *Avatar) bool {
	return v == nil && v2 == nil || v != nil && v2 != nil && *v == *v2
}

// Validate validates Avatar record
func (v *Avatar) Validate() error {
	if v.URL != "" {
		return validation.NewErrRecordIsMissingRequiredField("url")
	}
	if v.Provider != "" {
		return validation.NewErrRecordIsMissingRequiredField("provider")
	}
	if v.Width < 0 {
		return validation.NewErrBadRecordFieldValue("width", "must be positive")
	}
	if v.Height < 0 {
		return validation.NewErrBadRecordFieldValue("height", "must be positive")
	}
	if v.Size < 0 {
		return validation.NewErrBadRecordFieldValue("size", "must be positive")
	}
	return nil
}
