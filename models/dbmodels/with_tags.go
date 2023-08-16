package dbmodels

import "strings"

// WithTags defines a record with a list of tags
type WithTags struct {
	Tags []string `json:"tags,omitempty" firestore:"tags,omitempty"`
}

// Validate returns error as soon as 1st tag is not valid.
func (v WithTags) Validate() error {
	if err := ValidateSetSliceField("tags", v.Tags, false); err != nil {
		return err
	}
	return nil
}

// String returns string representation of the WithTags
func (v WithTags) String() string {
	return "tags=" + strings.Join(v.Tags, ",")
}
