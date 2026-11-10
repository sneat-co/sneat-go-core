package sneatauth

import (
	"errors"
	"net/http"
	"strings"
)

var (
	// ErrInvalidExternalAuthConfig indicates invalid policy or wrapper configuration,
	// such as an empty or invalid reserved UID set, or a nil/zero policy or verifier.
	ErrInvalidExternalAuthConfig = errors.New("invalid external auth configuration")

	// ErrReservedExternalUID indicates that an externally presented token matched
	// a reserved external UID and was denied.
	ErrReservedExternalUID = errors.New("reserved external UID")

	// ErrInvalidExternalUID indicates that required external authentication produced
	// no token, or that a presented token has an empty or whitespace-only UID.
	ErrInvalidExternalUID = errors.New("invalid external UID")
)

// ExternalTokenVerifier verifies an external HTTP request and returns an authenticated Token.
// It matches the function signature used at the external token-getter boundary (e.g. apicore.GetAuthTokenFromHttpRequest).
type ExternalTokenVerifier func(r *http.Request, authRequired bool) (*Token, error)

// ReservedExternalUIDPolicy enforces an immutable set of reserved external UIDs
// that must never be authenticated by external credentials.
type ReservedExternalUIDPolicy struct {
	reserved map[string]struct{}
}

// NewReservedExternalUIDPolicy creates an immutable ReservedExternalUIDPolicy from an explicit
// non-empty slice of reserved external UIDs.
//
// It rejects nil or empty slices, as well as any empty or whitespace-only UID entries,
// returning ErrInvalidExternalAuthConfig. Duplicate entries are deduplicated.
// The input slice is defensively copied into immutable private state so that subsequent
// caller mutations do not affect the policy. Nonblank IDs are preserved exactly as supplied
// (case-sensitive, preserving prefixes and whitespace-wrapped nonblank characters without normalization).
func NewReservedExternalUIDPolicy(reservedUIDs []string) (*ReservedExternalUIDPolicy, error) {
	if len(reservedUIDs) == 0 {
		return nil, ErrInvalidExternalAuthConfig
	}
	reserved := make(map[string]struct{}, len(reservedUIDs))
	for _, uid := range reservedUIDs {
		if strings.TrimSpace(uid) == "" {
			return nil, ErrInvalidExternalAuthConfig
		}
		reserved[uid] = struct{}{}
	}
	return &ReservedExternalUIDPolicy{reserved: reserved}, nil
}

// CheckUID checks whether a given external UID is valid and not reserved.
//
// A nil or zero-value policy fails closed and returns ErrInvalidExternalAuthConfig.
// If the UID is empty or contains only whitespace, ErrInvalidExternalUID is returned.
// If the UID exactly matches a reserved UID, ErrReservedExternalUID is returned.
// Otherwise it returns nil.
func (p *ReservedExternalUIDPolicy) CheckUID(uid string) error {
	if p == nil || len(p.reserved) == 0 {
		return ErrInvalidExternalAuthConfig
	}
	if strings.TrimSpace(uid) == "" {
		return ErrInvalidExternalUID
	}
	if _, ok := p.reserved[uid]; ok {
		return ErrReservedExternalUID
	}
	return nil
}

// WrapExternalTokenVerifier wraps an ExternalTokenVerifier with the given ReservedExternalUIDPolicy.
//
// It fails closed and returns (nil, ErrInvalidExternalAuthConfig) if the verifier is nil or if the policy
// is nil or has a zero-value (empty reserved set). It never returns a callable wrapper on construction error.
//
// The returned wrapper forwards the identical request pointer and authRequired flag to the captured verifier
// exactly once:
//   - If the verifier returns an error, nil token and that same error object are returned (any accompanying token is discarded).
//   - If the verifier returns nil token and nil error with optional authentication (authRequired == false), nil, nil is returned.
//   - If authentication is required (authRequired == true) and the verifier returns nil token and nil error, ErrInvalidExternalUID is returned.
//   - Any non-nil token with an empty or whitespace-only UID is rejected with ErrInvalidExternalUID.
//   - A token whose UID matches a reserved UID is rejected with ErrReservedExternalUID, regardless of claims.
//   - On success, the exact token pointer returned by the verifier is returned with its fields preserved.
//   - Each call evaluates the current returned token independently (no caching).
func WrapExternalTokenVerifier(verifier ExternalTokenVerifier, policy *ReservedExternalUIDPolicy) (ExternalTokenVerifier, error) {
	if verifier == nil {
		return nil, ErrInvalidExternalAuthConfig
	}
	if policy == nil || len(policy.reserved) == 0 {
		return nil, ErrInvalidExternalAuthConfig
	}
	return func(r *http.Request, authRequired bool) (*Token, error) {
		tok, err := verifier(r, authRequired)
		if err != nil {
			return nil, err
		}
		if tok == nil {
			if authRequired {
				return nil, ErrInvalidExternalUID
			}
			return nil, nil
		}
		if err := policy.CheckUID(tok.UID); err != nil {
			return nil, err
		}
		return tok, nil
	}, nil
}

// WrapVerifier is a convenience method on ReservedExternalUIDPolicy delegating to WrapExternalTokenVerifier.
func (p *ReservedExternalUIDPolicy) WrapVerifier(verifier ExternalTokenVerifier) (ExternalTokenVerifier, error) {
	return WrapExternalTokenVerifier(verifier, p)
}
