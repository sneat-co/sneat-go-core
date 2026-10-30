package sneatauth

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"net/url"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// CredentialFamily identifies a verified credential, independently of its UID.
type CredentialFamily string

const (
	ProductDelegationFamily      CredentialFamily = "product-delegation.v1"
	ProductDelegationPrefix                       = "pd1_"
	ProductDelegationPath                         = "/product-delegation/v1"
	MaxProductDelegationLifetime                  = time.Hour
)

var ErrInvalidProductDelegation = errors.New("invalid product delegation")

// NewProductDelegationBearer reads exactly 32 bytes from a cryptographically
// secure entropy source. Production callers must use crypto/rand.Reader.
func NewProductDelegationBearer(entropy io.Reader) (bearer, digest string, err error) {
	if entropy == nil {
		return "", "", ErrInvalidProductDelegation
	}
	var secret [32]byte
	if _, err = io.ReadFull(entropy, secret[:]); err != nil {
		return "", "", err
	}
	bearer = ProductDelegationPrefix + base64.RawURLEncoding.EncodeToString(secret[:])
	return bearer, productBearerDigest(bearer), nil
}

// FirebaseIdentity contains host-verified Firebase facts. These facts alone do
// not prove independent human consent or decision eligibility after recovery.
// Consumers must never populate them from request bodies or decoded JWTs.
type FirebaseIdentity struct {
	UID             string
	FirebaseProject string
	FirebaseTenant  string
	AuthTime        time.Time
	IssuedAt        time.Time
	ExpiresAt       time.Time
	SignInProvider  string
}

// ProductDelegation is the authoritative product credential record and the
// typed principal returned by a verifier. It is never a Firebase/browser token.
// Digest is lowercase hex SHA-256 of the complete canonical prefixed bearer.
type ProductDelegation struct {
	Family          CredentialFamily `json:"family" firestore:"family"`
	Digest          string           `json:"digest" firestore:"digest"`
	Issuer          string           `json:"issuer" firestore:"issuer"`
	FirebaseProject string           `json:"firebaseProject" firestore:"firebaseProject"`
	FirebaseTenant  string           `json:"firebaseTenant" firestore:"firebaseTenant"`
	UserID          string           `json:"userId" firestore:"userId"`
	ClientID        string           `json:"clientId" firestore:"clientId"`
	Audience        string           `json:"audience" firestore:"audience"`
	Product         string           `json:"product" firestore:"product"`
	Scopes          []string         `json:"scopes" firestore:"scopes"`
	SourceID        string           `json:"sourceId" firestore:"sourceId"`
	IssuedAt        time.Time        `json:"issuedAt" firestore:"issuedAt"`
	ExpiresAt       time.Time        `json:"expiresAt" firestore:"expiresAt"`
	RevokedAt       *time.Time       `json:"revokedAt,omitempty" firestore:"revokedAt,omitempty"`
}

// ProductDelegationDigest refuses noncanonical, unknown and Firebase tokens.
// It does not authenticate the bearer; the current server record is required.
func ProductDelegationDigest(bearer string) (string, error) {
	secret, ok := strings.CutPrefix(bearer, ProductDelegationPrefix)
	if !ok {
		return "", ErrInvalidProductDelegation
	}
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(secret)
	if err != nil || len(decoded) != 32 || base64.RawURLEncoding.EncodeToString(decoded) != secret {
		return "", ErrInvalidProductDelegation
	}
	return productBearerDigest(bearer), nil
}

func productBearerDigest(bearer string) string {
	digest := sha256.Sum256([]byte(bearer))
	return hex.EncodeToString(digest[:])
}

// ValidateProductDelegationRealm freezes v1 to one configured project and no
// tenants. A future tenant contract requires independent review.
func ValidateProductDelegationRealm(issuer, project, tenant string) error {
	u, err := url.Parse(issuer)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.Host != strings.ToLower(u.Host) ||
		u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.RawPath != "" ||
		u.Path != ProductDelegationPath || u.String() != issuer || tenant != "" || !validCredentialLabel(project) {
		return ErrInvalidProductDelegation
	}
	return nil
}

// Validate checks the closed record shape, not product permissions.
func (p ProductDelegation) Validate() error {
	if p.Family != ProductDelegationFamily || !ValidProductDelegationDigest(p.Digest) ||
		ValidateProductDelegationRealm(p.Issuer, p.FirebaseProject, p.FirebaseTenant) != nil ||
		!validCredentialLabel(p.UserID) || !validCredentialLabel(p.ClientID) || p.Audience != p.ClientID ||
		!validCredentialLabel(p.Product) || !validCredentialLabel(p.SourceID) || len(p.Scopes) == 0 ||
		p.IssuedAt.IsZero() || !p.ExpiresAt.After(p.IssuedAt) || p.ExpiresAt.Sub(p.IssuedAt) > MaxProductDelegationLifetime ||
		(p.RevokedAt != nil && p.RevokedAt.Before(p.IssuedAt)) {
		return ErrInvalidProductDelegation
	}
	for i, scope := range p.Scopes {
		if !validCredentialLabel(scope) || slices.Contains(p.Scopes[:i], scope) {
			return ErrInvalidProductDelegation
		}
	}
	return nil
}

// ActiveAt refuses future-issued, expired and revoked credentials. Call after
// resolving current state; never extend the stored deadline on a refresh.
func (p ProductDelegation) ActiveAt(now time.Time) error {
	if p.Validate() != nil || now.Before(p.IssuedAt) || !now.Before(p.ExpiresAt) || p.RevokedAt != nil {
		return ErrInvalidProductDelegation
	}
	return nil
}

// Clone prevents a caller's retained scopes or timestamp pointer widening a
// later verification result.
func (p ProductDelegation) Clone() ProductDelegation {
	p.Scopes = slices.Clone(p.Scopes)
	if p.RevokedAt != nil {
		timestamp := *p.RevokedAt
		p.RevokedAt = &timestamp
	}
	return p
}

// ValidProductDelegationDigest also protects the Firestore document-ID boundary.
func ValidProductDelegationDigest(digest string) bool {
	decoded, err := hex.DecodeString(digest)
	return err == nil && len(decoded) == sha256.Size && hex.EncodeToString(decoded) == digest
}

func validCredentialLabel(value string) bool {
	return value != "" && value != "." && value != ".." && utf8.ValidString(value) && len(value) <= 256 && !strings.ContainsAny(value, "/\\") &&
		strings.IndexFunc(value, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) < 0
}
