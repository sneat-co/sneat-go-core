package dbmodels

import (
	"github.com/strongo/slice"
)

// WithRoles defines a record with a list of roles
type WithRoles struct {
	Roles []string `json:"roles,omitempty" firestore:"roles,omitempty"`
}

// HasRole checks if a members has a given role
func (v WithRoles) HasRole(role string) bool {
	return slice.Index(v.Roles, role) >= 0
}

func (v WithRoles) AddRole(role string) ( /* u dal.Update - does not make sense to return update as field unknown */ ok bool) {
	if v.HasRole(role) {
		return false
	}
	v.Roles = append(v.Roles, role)
	return true
}

// Validate returns error as soon as 1st role is not valid.
func (v WithRoles) Validate() error {
	if err := ValidateSetSliceField("roles", v.Roles, true); err != nil {
		return err
	}
	return nil
}
