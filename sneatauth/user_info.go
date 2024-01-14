package sneatauth

import (
	"context"
	"fmt"
)

var GetUserInfo = func(ctx context.Context, uid string) (authUser *AuthUserInfo, err error) {
	panic("GetUserInfo is not initialized")
}

type AuthProviderUserInfo struct {
	DisplayName string `json:"displayName,omitempty"`
	Email       string `json:"email,omitempty"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
	PhotoURL    string `json:"photoUrl,omitempty"`
	// In the ProviderUserInfo[] ProviderID can be a short domain name (e.g. google.com),
	// or the identity of an OpenID identity provider.
	// In AuthUserInfo it will return the constant string "firebase".
	ProviderID string `json:"providerId,omitempty"`
	UID        string `json:"rawId,omitempty"`
}

func (v AuthProviderUserInfo) String() string {
	return fmt.Sprintf("AuthUserInfo{ProviderID:%s, UID:%s, DisplayName:%s}",
		v.ProviderID, v.UID, v.DisplayName)
}

// MultiFactorInfo describes a user enrolled second phone factor.
// TODO : convert PhoneNumber to PhoneMultiFactorInfo struct
type MultiFactorInfo struct {
	UID                 string
	DisplayName         string
	EnrollmentTimestamp int64
	FactorID            string
	PhoneNumber         string
}

// MultiFactorSettings describes the multi-factor related user settings.
type MultiFactorSettings struct {
	EnrolledFactors []*MultiFactorInfo
}

// UserMetadata contains additional metadata associated with a user account.
// Timestamps are in milliseconds since epoch.
type UserMetadata struct {
	CreationTimestamp  int64
	LastLogInTimestamp int64
	// The time at which the user was last active (ID token refreshed), or 0 if
	// the user was never active.
	LastRefreshTimestamp int64
}

type AuthUserInfo struct {
	*AuthProviderUserInfo
	CustomClaims           map[string]interface{}
	Disabled               bool
	EmailVerified          bool
	ProviderUserInfo       []*AuthProviderUserInfo
	TokensValidAfterMillis int64 // milliseconds since epoch.
	UserMetadata           *UserMetadata
	TenantID               string
	MultiFactor            *MultiFactorSettings
}

func (v AuthUserInfo) String() string {
	return fmt.Sprintf("AuthUserInfo{UID:%s ProviderID:%s}", v.UID, v.ProviderID)
}
