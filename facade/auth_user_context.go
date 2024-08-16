package facade

import (
	"context"
	"fmt"
	"strings"
)

// AuthContext defines auth context interface
type AuthContext interface {
	User(ctx context.Context, authRequired bool) (UserContext, error)
}

var _ UserContext = (*AuthUserContext)(nil)

// AuthUserContext implements UserContext
type AuthUserContext struct {
	ID string `json:"id" firestore:"id" dalgo:"id"`
}

// String returns string representation of AuthUserContext
func (v AuthUserContext) String() string {
	return fmt.Sprintf("{id=%s}", v.ID)
}

// Validate returns error if AuthUserContext is invalid
func (v AuthUserContext) Validate() error {
	if strings.TrimSpace(v.ID) == "" {
		return fmt.Errorf("field AuthUserContext.ID is empty string")
	}
	return nil
}

// GetUserIDFromContext returns user ID
func (v AuthUserContext) GetUserID() string {
	return v.ID
}
