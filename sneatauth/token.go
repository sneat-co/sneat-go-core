package sneatauth

// Token represents an authentication token.
type Token struct {
	UID      string `json:"uid,omitempty"`
	Original any    `json:"original,omitempty"`
}
