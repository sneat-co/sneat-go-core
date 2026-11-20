package sneatauth

// Token contains claims verified by the host authentication adapter.
// Adapters must populate these fields only after verifying the credential;
// client-supplied claims are not authority.
type Token struct {
	UID    string `json:"uid,omitempty"`
	Issuer string `json:"issuer,omitempty"`
	// IsAdmin is true only for an explicitly verified boolean admin claim.
	IsAdmin bool `json:"isAdmin,omitempty"`
	// Email is the verified credential's email claim, as issued.
	Email string `json:"email,omitempty"`
	// EmailVerified is true only when the identity provider marked Email
	// verified ("email_verified" claim).
	EmailVerified bool `json:"emailVerified,omitempty"`
	// SignInProvider is the provider the user signed in with, e.g.
	// "google.com" (Firebase "firebase.sign_in_provider" claim).
	SignInProvider string `json:"signInProvider,omitempty"`
	Original       any    `json:"original,omitempty"`
}
