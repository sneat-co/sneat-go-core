package dbmodels

import (
	"fmt"
	"github.com/strongo/validation"
	"strings"
)

// WithFlags defines a record with a list of flags
type WithFlags struct {
	Flags []string `json:"flags,omitempty" firestore:"flags,omitempty"`
}

// Validate returns error as soon as 1st flag is not valid.
func (v WithFlags) Validate() error {
	for i, flag := range v.Flags {
		if strings.TrimSpace(flag) == "" {
			return validation.NewErrRecordIsMissingRequiredField(fmt.Sprintf("flags[%v]", i))
		}
	}
	return nil
}

// String returns string representation of the WithTags
func (v WithFlags) String() string {
	return "flags=" + strings.Join(v.Flags, ",")
}
