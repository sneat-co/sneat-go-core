package facade

import (
	"context"
)

var userIDContextKey = 0

// GetUserID gets AuthUser ContactID from context
func GetUserID(ctx context.Context) string {
	v := ctx.Value(&userIDContextKey)
	if v == nil {
		return ""
	}
	return v.(string)
}

// NewContextWithUserID creates a new context with AuthUser ContactID
func NewContextWithUserID(parent context.Context, userID string) context.Context {
	return context.WithValue(parent, &userIDContextKey, userID)
}
