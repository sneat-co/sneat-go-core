package sneatauth

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetUserInfo(t *testing.T) {
	assert.Panics(t, func() {
		_, _ = GetUserInfo(context.Background(), "u1")
	})
}

func TestUserInfo_String(t *testing.T) {
	providerInfo := AuthProviderUserInfo{ProviderID: "p1", UID: "u1", DisplayName: "User 1"}
	assert.NotEmpty(t, providerInfo.String())

	authInfo := AuthUserInfo{AuthProviderUserInfo: &providerInfo}
	assert.NotEmpty(t, authInfo.String())
}
