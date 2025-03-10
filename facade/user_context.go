package facade

import (
	"fmt"
	"strings"
)

type UserContext struct {
	userID string
}

// NewUserContext creates new contextWithUser context
func NewUserContext(userID string) (userCtx *UserContext) {
	if strings.TrimSpace(userID) == "" {
		panic("userID is empty string")
	}
	return &UserContext{userID: userID}
}

// String returns string representation of contextWithUser
func (v *UserContext) String() string {
	return fmt.Sprintf("UserContext{id=%s}", v.userID)
}

// GetUserID returns user userID
func (v *UserContext) GetUserID() string {
	return v.userID
}

func (v *UserContext) Validate() error {
	if strings.TrimSpace(v.userID) == "" {
		return fmt.Errorf("field `UserContext.userID` is empty string")
	}
	return nil
}
