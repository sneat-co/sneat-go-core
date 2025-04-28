package facade

import (
	"context"
)

var _ ContextWithUser = (*contextWithUser)(nil)

// contextWithUser implements userContext
type contextWithUser struct {
	user UserContext
	context.Context
}

func (v contextWithUser) User() UserContext {
	return v.user
}

var _ context.Context = &contextWithUser{}
var _ ContextWithUser = &contextWithUser{}

type ContextWithUser interface {
	context.Context
	User() UserContext
}

var userContextKey = "contextWithUser"

func NewContextWithUserID(ctx context.Context, userID string) ContextWithUser {
	userCtx := NewUserContext(userID)
	return NewContextWithUserContext(ctx, userCtx)
}

func NewContextWithUserContext(ctx context.Context, userCtx UserContext) ContextWithUser {
	ctxWithUser := contextWithUser{user: userCtx}
	ctxWithUser.Context = context.WithValue(ctx, &userContextKey, ctxWithUser.user)
	return ctxWithUser
}

func GetUserContext(ctx context.Context) UserContext {
	return ctx.Value(&userContextKey).(UserContext)
}
