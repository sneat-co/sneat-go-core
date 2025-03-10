package facade

import (
	"context"
)

// AuthContext defines auth context provider that can be used to retrieve user context.
// For example, take Firebase ID token from request header, verify, return user context with a user ID provided by token
// This is implemented in github.com/sneat-co/sneat-go-firebase module.
type AuthContext interface {
	User(ctx context.Context, authRequired bool) (UserContext, error)
}
