package sneatauth

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

func validProductFixture(t *testing.T) (string, ProductDelegation) {
	t.Helper()
	bearer, digest, err := NewProductDelegationBearer(bytes.NewReader(bytes.Repeat([]byte{7}, 32)))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	return bearer, ProductDelegation{Family: ProductDelegationFamily, Digest: digest, Issuer: "https://auth.sneat.co" + ProductDelegationPath,
		FirebaseProject: "sneat-eur3-1", UserID: "uid", ClientID: "datatug-cli", Audience: "datatug-cli", Product: "datatug",
		Scopes: []string{"openid", "datatug:projects:read"}, SourceID: "source", IssuedAt: now, ExpiresAt: now.Add(time.Hour)}
}

func TestProductBearerCanonicalDigest(t *testing.T) {
	bearer, principal := validProductFixture(t)
	digest, err := ProductDelegationDigest(bearer)
	want := sha256.Sum256([]byte(bearer))
	if err != nil || digest != hex.EncodeToString(want[:]) || digest != principal.Digest || !ValidProductDelegationDigest(digest) {
		t.Fatalf("canonical digest mismatch: %v", err)
	}
	for _, invalid := range []string{"", "ey.token.jwt", "pd2_" + bearer[4:], bearer + "=", " " + bearer, bearer + "\n", "pd1_" + base64.RawURLEncoding.EncodeToString(make([]byte, 31)), "pd1_" + strings.Repeat("_", 43)} {
		if _, err = ProductDelegationDigest(invalid); !errors.Is(err, ErrInvalidProductDelegation) {
			t.Fatal("noncanonical or foreign family accepted")
		}
	}
	for _, invalid := range []string{"", strings.ToUpper(digest), digest[:62], strings.Repeat("z", 64)} {
		if ValidProductDelegationDigest(invalid) {
			t.Fatal("unsafe digest accepted")
		}
	}
	if bearer, digest, err := NewProductDelegationBearer(strings.NewReader("short")); !errors.Is(err, io.ErrUnexpectedEOF) || bearer != "" || digest != "" {
		t.Fatalf("entropy failure released a credential: %v", err)
	}
	if bearer, digest, err := NewProductDelegationBearer(nil); !errors.Is(err, ErrInvalidProductDelegation) || bearer != "" || digest != "" {
		t.Fatal("nil entropy source accepted")
	}
}

func TestProductRecordClosedValidation(t *testing.T) {
	_, p := validProductFixture(t)
	if p.Validate() != nil || p.ActiveAt(p.IssuedAt) != nil || p.ActiveAt(p.ExpiresAt.Add(-time.Nanosecond)) != nil {
		t.Fatal("valid record refused")
	}
	for name, change := range map[string]func(*ProductDelegation){
		"unknown family":    func(p *ProductDelegation) { p.Family = "browser" },
		"invalid digest":    func(p *ProductDelegation) { p.Digest = "../digest" },
		"wrong issuer path": func(p *ProductDelegation) { p.Issuer = "https://auth.sneat.co" },
		"missing project":   func(p *ProductDelegation) { p.FirebaseProject = "" },
		"tenant":            func(p *ProductDelegation) { p.FirebaseTenant = "tenant" },
		"missing user":      func(p *ProductDelegation) { p.UserID = "" },
		"unsafe user":       func(p *ProductDelegation) { p.UserID = "../user" },
		"invalid UTF-8":     func(p *ProductDelegation) { p.UserID = "\xff" },
		"dot user":          func(p *ProductDelegation) { p.UserID = ".." },
		"control":           func(p *ProductDelegation) { p.UserID = "uid\n" },
		"oversize":          func(p *ProductDelegation) { p.UserID = strings.Repeat("a", 257) },
		"client":            func(p *ProductDelegation) { p.ClientID = "" },
		"audience":          func(p *ProductDelegation) { p.Audience = "sneat-cli" },
		"product":           func(p *ProductDelegation) { p.Product = "" },
		"source":            func(p *ProductDelegation) { p.SourceID = "" },
		"scope absent":      func(p *ProductDelegation) { p.Scopes = nil },
		"scope invalid":     func(p *ProductDelegation) { p.Scopes = []string{" "} },
		"scope duplicate":   func(p *ProductDelegation) { p.Scopes = []string{"openid", "openid"} },
		"issued absent":     func(p *ProductDelegation) { p.IssuedAt = time.Time{} },
		"zero lifetime":     func(p *ProductDelegation) { p.ExpiresAt = p.IssuedAt },
		"long lifetime":     func(p *ProductDelegation) { p.ExpiresAt = p.IssuedAt.Add(time.Hour + time.Nanosecond) },
		"bad revocation":    func(p *ProductDelegation) { at := p.IssuedAt.Add(-time.Second); p.RevokedAt = &at },
	} {
		t.Run(name, func(t *testing.T) {
			value := p.Clone()
			change(&value)
			if !errors.Is(value.Validate(), ErrInvalidProductDelegation) || !errors.Is(value.ActiveAt(p.IssuedAt), ErrInvalidProductDelegation) {
				t.Fatal("invalid stored credential accepted")
			}
		})
	}
	for _, now := range []time.Time{p.IssuedAt.Add(-time.Nanosecond), p.ExpiresAt, p.ExpiresAt.Add(time.Hour)} {
		if p.ActiveAt(now) == nil {
			t.Fatal("credential outside hard lifetime accepted")
		}
	}
	p.RevokedAt = &p.IssuedAt
	if p.ActiveAt(p.IssuedAt) == nil {
		t.Fatal("revoked credential accepted")
	}
	copy := p.Clone()
	copy.Scopes[0] = "broadened"
	*copy.RevokedAt = copy.RevokedAt.Add(time.Minute)
	if p.Scopes[0] != "openid" || !p.RevokedAt.Equal(p.IssuedAt) {
		t.Fatal("clone retained mutable authority")
	}
}

func TestProductRealmIsClosed(t *testing.T) {
	for _, issuer := range []string{"%", "http://auth.sneat.co/product-delegation/v1", "https://Auth.sneat.co/product-delegation/v1", "https://user@auth.sneat.co/product-delegation/v1", "https://auth.sneat.co/product-delegation/v1?x=1", "https://auth.sneat.co/product-delegation/v1#fragment", "https://auth.sneat.co/product%2Ddelegation/v1", "https:///product-delegation/v1", "https://:443/product-delegation/v1", "https://auth.sneat.co/product-delegation/v1?", "https://auth.sneat.co:443/product-delegation/v1", "https://auth.sneat.co:/product-delegation/v1"} {
		if ValidateProductDelegationRealm(issuer, "project", "") == nil {
			t.Fatal("invalid realm accepted")
		}
	}
	// Pinned deviceauth canonicalizes an explicit default/empty port away.
	// Only that resulting identity is accepted; a nondefault port is retained.
	for _, issuer := range []string{"https://auth.sneat.co/product-delegation/v1", "https://auth.sneat.co:8443/product-delegation/v1"} {
		if err := ValidateProductDelegationRealm(issuer, "project", ""); err != nil {
			t.Fatal("canonical client issuer refused", err)
		}
	}
}

func TestOversizedProductBearerRefusedBeforeDecodeAllocation(t *testing.T) {
	oversized := ProductDelegationPrefix + strings.Repeat("A", 4096)
	if allocations := testing.AllocsPerRun(10, func() {
		digest, err := ProductDelegationDigest(oversized)
		if digest != "" || !errors.Is(err, ErrInvalidProductDelegation) {
			t.Fatal("oversized bearer accepted")
		}
	}); allocations != 0 {
		t.Fatal("oversized bearer allocated a decode buffer")
	}
}
