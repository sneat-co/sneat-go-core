package facade

import (
	"fmt"
	"strings"
)

type UserContext interface {
	GetUserID() string
}

type userContext struct {
	userID string
}

// NewUserContext creates new contextWithUser context
func NewUserContext(userID string) (userCtx UserContext) {
	if strings.TrimSpace(userID) == "" {
		panic("userID is empty string")
	}
	return &userContext{userID: userID}
}

// String returns string representation of contextWithUser
func (v *userContext) String() string {
	return fmt.Sprintf("userContext{id=%s}", v.userID)
}

// GetUserID returns user userID
func (v *userContext) GetUserID() string {
	return v.userID
}

func (v *userContext) Validate() error {
	if strings.TrimSpace(v.userID) == "" {
		return fmt.Errorf("field `userContext.userID` is empty string")
	}
	return nil
}
