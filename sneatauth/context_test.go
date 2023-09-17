package sneatauth

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAuthToken(t *testing.T) {
	parent := context.Background()
	token := AuthTokenFromContext(parent)
	assert.Nil(t, token)
	ctx := NewContextWithAuthToken(parent, nil)
	token = AuthTokenFromContext(ctx)
	assert.Nil(t, token)
	token1 := &Token{UID: "user1"}
	ctx = NewContextWithAuthToken(parent, token1)
	token = AuthTokenFromContext(ctx)
	assert.Same(t, token1, token)
}
