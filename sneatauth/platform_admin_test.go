package sneatauth

import "testing"

func TestIsPlatformAdmin(t *testing.T) {
	admin := func(mutate func(*Token)) *Token {
		token := &Token{UID: "u1", Email: PlatformAdminEmail, EmailVerified: true, SignInProvider: SignInProviderGoogle}
		if mutate != nil {
			mutate(token)
		}
		return token
	}
	tests := []struct {
		name  string
		token *Token
		want  bool
	}{
		{name: "nil token", token: nil, want: false},
		{name: "google admin", token: admin(nil), want: true},
		{name: "email case and spaces", token: admin(func(t *Token) { t.Email = "  Alex@Sneat.CO " }), want: true},
		{name: "no UID", token: admin(func(t *Token) { t.UID = "" }), want: false},
		{name: "blank UID", token: admin(func(t *Token) { t.UID = "  " }), want: false},
		{name: "unverified email", token: admin(func(t *Token) { t.EmailVerified = false }), want: false},
		{name: "former gmail admin", token: admin(func(t *Token) { t.Email = "alexander.trakhimenok@gmail.com" }), want: false},
		{name: "other email", token: admin(func(t *Token) { t.Email = "mallory@sneat.co" }), want: false},
		{name: "lookalike email", token: admin(func(t *Token) { t.Email = "alex@sneat.co.evil.example" }), want: false},
		{name: "unicode long s folds to s", token: admin(func(t *Token) { t.Email = "alex@ſneat.co" }), want: false},
		{name: "same-length non-ascii", token: admin(func(t *Token) { t.Email = "alex@sneat.cö"[:len(PlatformAdminEmail)] }), want: false},
		{name: "empty email", token: admin(func(t *Token) { t.Email = "" }), want: false},
		{name: "microsoft provider", token: admin(func(t *Token) { t.SignInProvider = "microsoft.com" }), want: false},
		{name: "password provider", token: admin(func(t *Token) { t.SignInProvider = "password" }), want: false},
		{name: "custom token provider", token: admin(func(t *Token) { t.SignInProvider = "custom" }), want: false},
		{name: "no provider", token: admin(func(t *Token) { t.SignInProvider = "" }), want: false},
		{name: "admin claim alone is not enough", token: &Token{UID: "u1", IsAdmin: true}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPlatformAdmin(tt.token); got != tt.want {
				t.Fatalf("IsPlatformAdmin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasTrustedEmail(t *testing.T) {
	if !HasTrustedEmail(&Token{UID: "u", Email: "someone@example.com", EmailVerified: true, SignInProvider: SignInProviderGoogle}) {
		t.Fatal("verified Google email must be trusted")
	}
	if HasTrustedEmail(&Token{UID: "u", Email: "  ", EmailVerified: true, SignInProvider: SignInProviderGoogle}) {
		t.Fatal("blank email must not be trusted")
	}
}
