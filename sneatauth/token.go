package sneatauth

// Token contains claims verified by the host authentication adapter.
// Adapters must populate these fields only after verifying the credential;
// client-supplied claims are not authority.
type Token struct {
	UID    string `json:"uid,omitempty"`
	Issuer string `json:"issuer,omitempty"`
	// IsAdmin is true only for an explicitly verified boolean admin claim.
	IsAdmin  bool `json:"isAdmin,omitempty"`
	Original any  `json:"original,omitempty"`
}
