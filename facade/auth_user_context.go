package facade

import (
	"context"
)

// AuthContextObsolete defines auth context interface
// deprecated: use ContextWithUser instead?
type AuthContextObsolete interface {
	User(ctx context.Context, authRequired bool) (UserContext, error)
}
