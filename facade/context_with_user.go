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

// spaceAccessAuthorizedByCallerContext is the UserContext behind
// SpaceAccessAuthorizedByCaller: a deliberately empty user ID that opts the
// current DAL transaction out of dal4spaceus's space-membership gate. See
// SpaceAccessAuthorizedByCaller for the full contract before using this type
// directly.
type spaceAccessAuthorizedByCallerContext struct{}

func (spaceAccessAuthorizedByCallerContext) GetUserID() string { return "" }

// SpaceAccessAuthorizedByCaller returns a facade.ContextWithUser whose
// GetUserID() is deliberately empty, wrapping ctx unchanged in every other
// respect.
//
// dal4spaceus's space-membership gate only runs when GetUserID() is
// non-empty: an empty user ID is that gate's deliberate opt-out, not a
// missing check. Running a space worker under the context returned here
// means the DAL's own "is this caller a member of the space" check will NOT
// run for that transaction.
//
// This is NOT anonymous or unauthenticated access. Call it only after the
// caller has already authorised the operation by its own proof -- because
// the DAL is intentionally trusting the facade layer to have done that
// check already, not because no check is needed. The known legitimate uses
// are:
//
//   - a maintenance / fixer job running with no user at all;
//   - a user joining a space by invite, who by definition is not yet a
//     member and would otherwise fail the ordinary membership gate before
//     the invite proof (ID + PIN, or equivalent) is even checked;
//   - moderation, such as removing a user from a space for spam, performed
//     by an actor who is not themselves a member of that space.
//
// The caller MUST keep its own authenticated user ID (or "no user", for a
// maintenance job) in a separate variable and keep using it for business
// logic -- audit trails, "who did this", any authorisation decision made
// above the DAL. This helper only neutralises the DAL's membership check;
// it does not erase who is actually acting.
//
// Using this to skip an authorisation you have not actually performed is a
// security bug: it silences the DAL's last line of defense without
// supplying the proof that defense exists to enforce.
func SpaceAccessAuthorizedByCaller(ctx context.Context) ContextWithUser {
	return NewContextWithUser(ctx, spaceAccessAuthorizedByCallerContext{})
}

func GetUserContext(ctx context.Context) UserContext {
	return ctx.Value(&userContextKey).(UserContext)
}
