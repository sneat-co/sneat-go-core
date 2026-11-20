package sneatauth

import "strings"

// PlatformAdminEmail is the only platform administrator of every Sneat
// website and API (founder ruling, 2026-11-20). Admin pages, CI admin
// endpoints and operator-only APIs all decide admin access with
// IsPlatformAdmin; none keeps an allow-list of its own.
const PlatformAdminEmail = "alex@sneat.co"

// Sign-in providers whose email claim the admin rule accepts. They are
// Firebase's "firebase.sign_in_provider" values for Google and Microsoft
// sign-in; every other provider (password, email link, phone, custom tokens,
// anonymous) is refused even with a verified matching email.
const (
	SignInProviderGoogle    = "google.com"
	SignInProviderMicrosoft = "microsoft.com"
)

// IsPlatformAdmin is the one Sneat admin rule: HasTrustedEmail(t) and the
// email is PlatformAdminEmail (case-insensitive).
//
// It trusts only the fields of t, which the host authentication adapter fills
// from a verified credential (see Token); it never reads request headers.
func IsPlatformAdmin(t *Token) bool {
	return HasTrustedEmail(t) && strings.EqualFold(strings.TrimSpace(t.Email), PlatformAdminEmail)
}

// HasTrustedEmail reports whether t's email can be used for an admin
// decision: t names a user, has a non-empty email the identity provider
// marked verified, and the user signed in with Google or Microsoft.
func HasTrustedEmail(t *Token) bool {
	if t == nil || t.UID == "" || !t.EmailVerified || strings.TrimSpace(t.Email) == "" {
		return false
	}
	switch t.SignInProvider {
	case SignInProviderGoogle, SignInProviderMicrosoft:
		return true
	default:
		return false
	}
}
