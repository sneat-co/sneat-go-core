package facade

import (
	"context"
	"fmt"
)

var _ ContextWithUser = (*contextWithUser)(nil)

// contextWithUser implements UserContext
type contextWithUser struct {
	user *UserContext
	context.Context
}

func (v contextWithUser) User() *UserContext {
	return v.user
}

// Validate returns error if contextWithUser is invalid
func (v contextWithUser) Validate() error {
	if err := v.user.Validate(); err != nil {
		return fmt.Errorf("field `contextWithUser.user` is invalid: %w", err)
	}
	if v.Context == nil {
		return fmt.Errorf("field contextWithUser.Context is nil")
	}
	return nil
}

var _ context.Context = &contextWithUser{}
var _ ContextWithUser = &contextWithUser{}

type ContextWithUser interface {
	context.Context
	User() *UserContext
}

var userContextKey = "contextWithUser"

func NewContextWithUser(ctx context.Context, userID string) ContextWithUser {
	ctxWithUser := contextWithUser{user: NewUserContext(userID)}
	ctxWithUser.Context = context.WithValue(ctx, &userContextKey, ctxWithUser.user)
	return ctxWithUser
}

func GetUserContext(ctx context.Context) *UserContext {
	return ctx.Value(&userContextKey).(*UserContext)
}
