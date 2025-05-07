package dbo4sharing

import "github.com/strongo/validation"

type To struct {
	Spaces map[string]Shared      `json:"spaces" firestore:"spaces"`
	Users  map[string]Permissions `json:"users" firestore:"users"`
}

func (v To) Validate() error {
	for id, shared := range v.Spaces {
		if err := shared.Validate(); err != nil {
			return validation.NewErrBadRecordFieldValue("spaces["+id+"]", err.Error())
		}
	}
	for id, permissions := range v.Users {
		if err := permissions.Validate(); err != nil {
			return validation.NewErrBadRecordFieldValue("users["+id+"]", err.Error())
		}
	}
	return nil
}

type Shared struct {
	ID          string      `json:"id" firestore:"id"` // This an ID of the shared item in the receiver space
	Permissions Permissions `json:"permissions" firestore:"permissions"`
}

func (v Shared) Validate() error {
	if v.ID == "" {
		return validation.NewErrRecordIsMissingRequiredField("id")
	}
	if len(v.Permissions) == 0 {
		return validation.NewErrRecordIsMissingRequiredField("permissions")
	}

	if err := v.Permissions.Validate(); err != nil {
		return validation.NewErrBadRecordFieldValue("permissions", err.Error())
	}
	return nil
}
