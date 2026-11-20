package sneatauth

import "strings"

// PlatformAdminEmail is the only platform administrator of every Sneat
// website and API (founder ruling, 2026-11-20). Admin pages, CI admin
// endpoints and operator-only APIs all decide admin access with
// IsPlatformAdmin; none keeps an allow-list of its own.
const PlatformAdminEmail = "alex@sneat.co"

// SignInProviderGoogle is Firebase's "firebase.sign_in_provider" value for
// Google sign-in, the one provider whose email the admin rule accepts.
//
// Google asserts email_verified only for a mailbox it controls. Microsoft is
// deliberately not accepted: Firebase reports its email as unverified, and an
// Entra ID tenant that is not pinned lets any tenant admin set a user's email
// to an arbitrary address (the "nOAuth" class). Every other provider
// (password, email link, phone, custom tokens, anonymous) is refused even with
// a verified matching email.
const SignInProviderGoogle = "google.com"

// IsPlatformAdmin is the one Sneat admin rule: HasTrustedEmail(t) and the
// email is PlatformAdminEmail, compared ASCII case-insensitively.
//
// It trusts only the fields of t, which the host authentication adapter fills
// from a verified credential (see Token); it never reads request headers.
func IsPlatformAdmin(t *Token) bool {
	return HasTrustedEmail(t) && IsPlatformAdminEmail(t.Email)
}

// IsPlatformAdminEmail reports whether email is PlatformAdminEmail, ignoring
// surrounding spaces and ASCII letter case. Any non-ASCII byte fails, so
// Unicode case folding (e.g. U+017F "ſ" folding to "s") cannot match. It
// says nothing about whether email is verified; see IsPlatformAdmin.
func IsPlatformAdminEmail(email string) bool {
	email = strings.TrimSpace(email)
	if len(email) != len(PlatformAdminEmail) {
		return false
	}
	for i := 0; i < len(email); i++ {
		c := email[i]
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		if c != PlatformAdminEmail[i] {
			return false
		}
	}
	return true
}

// HasTrustedEmail reports whether t's email can be used for an admin
// decision: t names a user, has a non-empty email the identity provider
// marked verified, and the user signed in with Google.
func HasTrustedEmail(t *Token) bool {
	return t != nil && strings.TrimSpace(t.UID) != "" && t.EmailVerified &&
		strings.TrimSpace(t.Email) != "" && t.SignInProvider == SignInProviderGoogle
}
