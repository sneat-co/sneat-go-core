package facade

import (
	"context"
)

var userIDContextKey = 0

// GetUserIDFromContext gets AuthUserContext ContactID from context
func GetUserIDFromContext(ctx context.Context) string {
	v := ctx.Value(&userIDContextKey)
	if v == nil {
		return ""
	}
	return v.(string)
}

// NewContextWithUserID creates a new context with AuthUserContext ContactID
func NewContextWithUserID(parent context.Context, userID string) context.Context {
	return context.WithValue(parent, &userIDContextKey, userID)
}
