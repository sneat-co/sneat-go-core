package sneatfb

import (
	"context"
	"firebase.google.com/go/v4/auth"
	"fmt"
	"github.com/sneat-co/sneat-go-core/facade"
)

func verifyIDToken(ctx context.Context, authClient *auth.Client, idToken string) (token *auth.Token, err error) {
	defer func() {
		if fail := recover(); fail != nil {
			err = fmt.Errorf("%w: %v", facade.ErrUnauthorized, fail)
		}
	}()
	token, err = authClient.VerifyIDToken(ctx, idToken)
	return
}

// NewContextWithFirebaseToken creates a new cotext with a Firebase token
func NewContextWithFirebaseToken(ctx context.Context, token *auth.Token) context.Context {
	return context.WithValue(ctx, &firebaseTokenContextKey, token)
}
