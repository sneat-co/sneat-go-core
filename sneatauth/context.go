package sneatauth

import "context"

var authTokenContextKey = "authTokenContextKey"

func NewContextWithAuthToken(parent context.Context, token *Token) context.Context {
	if token == nil && AuthTokenFromContext(parent) == nil {
		return parent
	}
	return context.WithValue(parent, &authTokenContextKey, token)
}

func AuthTokenFromContext(ctx context.Context) *Token {
	token, _ := ctx.Value(&authTokenContextKey).(*Token)
	return token
}
