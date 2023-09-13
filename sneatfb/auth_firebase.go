package sneatfb

import (
	"context"
	"firebase.google.com/go/v4/auth"
	"github.com/sneat-co/sneat-go-core/facade"
)

type firebaseAuthContext struct {
	fbIDToken   func() (string, error)
	err         error
	fbAuthToken *auth.Token
	userContext facade.User
}

var _ facade.AuthContext = (*firebaseAuthContext)(nil)

func (v *firebaseAuthContext) User(ctx context.Context, authRequired bool) (facade.User, error) {
	if v.err != nil || v.userContext != nil {
		return v.userContext, v.err
	}
	if err := v.getFbAuthToken(ctx, authRequired); err != nil {
		return v.userContext, v.err
	}
	v.userContext = facade.AuthUser{ID: v.fbAuthToken.UID}
	return v.userContext, nil
}

func (v *firebaseAuthContext) getFbAuthToken(ctx context.Context, authRequired bool) error {
	v.fbAuthToken, v.err = NewFirebaseAuthToken(ctx, v.fbIDToken, authRequired)
	return v.err
}

// NewFirebaseAuthContext creates new context with firebaseIDToken
func NewFirebaseAuthContext(firebaseIDToken func() (string, error)) facade.AuthContext {
	return &firebaseAuthContext{fbIDToken: firebaseIDToken}
}
