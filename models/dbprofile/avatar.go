package dbprofile

import "github.com/strongo/validation"

// Avatar record
type Avatar struct {
	External struct {
		Provider string `json:"provider" firestore:"provider"`
		URL      string `json:"url" firestore:"url"`
	} `json:"external" firestore:"external"`
	Gravatar string `json:"gravatar" firestore:"gravatar"`
}

func (v *Avatar) Equal(v2 *Avatar) bool {
	return v == nil && v2 == nil || v != nil && v2 != nil && *v == *v2
}

// Validate validates Avatar record
func (v *Avatar) Validate() error {
	if v.Gravatar != "" && v.External.URL != "" {
		return validation.NewErrBadRecordFieldValue("ToAvatar", "either `external.url` or `gravatar` should be set, not both")
	}
	if v.External.URL != "" && v.External.Provider == "" {
		return validation.NewErrBadRecordFieldValue("ToAvatar", "If `external.url` is set the `external.provider` also should be set")
	}
	return nil
}
