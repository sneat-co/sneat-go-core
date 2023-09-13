package sneatfb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"fmt"
	"github.com/sneat-co/sneat-go-core/facade"
	"log"
)

var firebaseTokenContextKey = "firebaseToken"

// FirebaseTokenFromContext gets Firebase token from context
func FirebaseTokenFromContext(ctx context.Context) *auth.Token {
	v := ctx.Value(&firebaseTokenContextKey)
	if v == nil {
		return nil
	}
	return v.(*auth.Token)
}

// NewFirebaseAuthToken creates Firebase authentication token
var NewFirebaseAuthToken = newFirebaseAuthToken

// NewFirebaseAuthToken creates a new Firebase Auth Token
func newFirebaseAuthToken(ctx context.Context, fbIDToken func() (string, error), authRequired bool) (*auth.Token, error) {
	if ctx == nil {
		panic("NewFirebaseAuthToken(ctx=nil)")
	}
	idToken, err := fbIDToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}
	fbApp, err := firebase.NewApp(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new firebase app: %w", err)
	}
	if idToken == "" {
		if authRequired {
			return nil, fmt.Errorf("%w: authentication is required but request is missing Firebase idToken", facade.ErrUnauthorized)
		}
		return nil, nil
	}
	fbAuth, err := fbApp.Auth(ctx)
	if err != nil && authRequired {
		return nil, fmt.Errorf("failed to create Firebase auth client: %w", err)
	}
	var token *auth.Token
	token, err = verifyIDToken(ctx, fbAuth, idToken)
	//isDemoProject := strings.HasPrefix(googleCloudProjectID, "demo")
	if err != nil {
		const m = "failed to verify Firebase ContactID idToken: %v\nidToken: %v"
		if authRequired {
			return token, fmt.Errorf(m, err, idToken)
		}
		log.Printf(m, err, idToken)
	}
	if token == nil {
		if authRequired {
			return nil, errors.New("firebase.auth.Client.VerifyIDToken() returned nil error and nil idToken")
		}
	} else if token.UID == "" {
		if authRequired {
			s := new(bytes.Buffer)
			_ = json.NewEncoder(s).Encode(token)
			return nil, fmt.Errorf("no UserID, decoded Token: %v\n\n encoded idToken: %v", s.String(), idToken)
		}
	}
	return token, nil
}
