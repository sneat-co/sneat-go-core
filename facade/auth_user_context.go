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

var _ UserContext = (*userContext)(nil)

// userContext implements UserContext
type userContext struct {
	userID string
	context.Context
}

func (v userContext) User() UserContext {
	return v
}

// String returns string representation of userContext
func (v userContext) String() string {
	return fmt.Sprintf("{id=%s}", v.userID)
}

// Validate returns error if userContext is invalid
func (v userContext) Validate() error {
	if strings.TrimSpace(v.userID) == "" {
		return fmt.Errorf("field userContext.userID is empty string")
	}
	return nil
}

// GetUserIDFromContext returns user userID
func (v userContext) GetUserID() string {
	return v.userID
}
