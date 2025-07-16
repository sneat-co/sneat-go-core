package facade

import (
	"context"
	"github.com/strongo/analytics"
	"github.com/strongo/logus"
)

var _ ContextWithUser = (*contextWithUser)(nil)

// contextWithUser implements userContext
type contextWithUser struct {
	user UserContext
	context.Context
	ua UserAnalytics
}

func (v contextWithUser) User() UserContext {
	return v.user
}

func (v contextWithUser) Analytics() UserAnalytics {
	if v.ua == nil {
		return noAnalytics{}
	}
	return v.ua
}

type noAnalytics struct{}

func (n noAnalytics) Send(msg analytics.Message) {
	logus.Errorf(context.Background(),
		"user context has no analytics: message{event: %s, category: %s}",
		msg.Event(), msg.Category())
}

var _ context.Context = &contextWithUser{}
var _ ContextWithUser = &contextWithUser{}

type ContextWithUser interface {
	context.Context
	User() UserContext
	Analytics() UserAnalytics
}

var userContextKey = "contextWithUser"

func NewContextWithUserID(ctx context.Context, userID string) ContextWithUser {
	userCtx := NewUserContext(userID)
	return NewContextWithUser(ctx, userCtx)
}

func NewContextWithUser(ctx context.Context, userCtx UserContext) ContextWithUser {
	ctxWithUser := contextWithUser{user: userCtx}
	ctxWithUser.Context = context.WithValue(ctx, &userContextKey, ctxWithUser.user)
	return ctxWithUser
}

func NewContextWithUserAndAnalytics(ctx context.Context, userCtx UserContext, ua UserAnalytics) ContextWithUser {
	ctxWithUser := contextWithUser{user: userCtx, ua: ua}
	ctxWithUser.Context = context.WithValue(ctx, &userContextKey, ctxWithUser.user)
	return ctxWithUser
}

func GetUserContext(ctx context.Context) UserContext {
	return ctx.Value(&userContextKey).(UserContext)
}
