package sneatfb

import (
	"context"
	"errors"
	"fmt"
	"github.com/sneat-co/sneat-go-core/apicore"
	"github.com/sneat-co/sneat-go-core/facade"
	"github.com/sneat-co/sneat-go-core/httpserver"
	"net/http"
	"strings"
)

const authorizationHeaderName = "Authorization"
const bearerPrefix = "Bearer"

// ContextWithFirebaseToken creates a context with a Firebase ContactID token
var ContextWithFirebaseToken = func(r *http.Request, authRequired bool) (ctx context.Context, err error) {
	ctx = r.Context()
	if ctx == nil {
		return ctx, errors.New("request returned nil context")
	}
	authHeader := r.Header.Get(authorizationHeaderName)
	if authHeader != "" || authRequired {
		bearerToken, err := GetBearerToken(authHeader)
		if err != nil {
			return ctx, fmt.Errorf("failed to get bearer token from authorization header: %w", err)
		}
		token, err := NewFirebaseAuthToken(ctx, func() (string, error) {
			return bearerToken, nil
		}, authRequired)
		if err != nil {
			return ctx, fmt.Errorf("failed to get Firebase auth toke: %w", err)
		}
		ctx = NewContextWithFirebaseToken(ctx, token)
		//log.Println("apicore.ContextWithFirebaseToken() is OK:", ctx)
	}
	return ctx, err
}

// NewAuthContext creates new authentication context
var NewAuthContext = func(r *http.Request) (facade.AuthContext, error) {
	fbIDToken := func() (string, error) {
		return GetBearerToken(r.Header.Get(authorizationHeaderName))
	}
	return NewFirebaseAuthContext(fbIDToken), nil
}

func GetBearerToken(authorizationHeader string) (token string, err error) {
	if authorizationHeader == "" {
		return "", facade.ErrNoAuthHeader
	}
	if !strings.HasPrefix(authorizationHeader, bearerPrefix) {
		return "", httpserver.ErrNotABearerToken
	}
	return authorizationHeader[len(bearerPrefix)+1:], nil
}

func init() {
	apicore.NewContextWithToken = ContextWithFirebaseToken
	apicore.NewAuthContext = NewAuthContext
}
