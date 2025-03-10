package facade

import (
	"context"
	"strings"
)

// UserContext defines an interface for a userContext context that provides userContext userID.
type UserContext interface {
	GetUserID() string
}

// NewUserContext creates new userContext context
// deprecated: use NewContextWithUser instead
func NewUserContext(id string) (userCtx UserContext) {
	if strings.TrimSpace(id) == "" {
		panic("userContext id is empty string")
	}
	userCtx = userContext{userID: id}
	return
}

var _ context.Context = &userContext{}

type ContextWithUser interface {
	context.Context
	User() UserContext
}

var userContextKey = "user"

func NewContextWithUser(ctx context.Context, userID string) ContextWithUser {
	userCtx := userContext{userID: userID}
	userCtx.Context = context.WithValue(ctx, &userContextKey, &userCtx)
	return userCtx
}

func GetUserContext(ctx context.Context) UserContext {
	return ctx.Value(&userContextKey).(UserContext)
}
