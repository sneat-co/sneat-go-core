// Copyright 2026 Sneat Co.
package firestorepolicy

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"

	"github.com/dal-go/dalgo/access"
	"github.com/dal-go/dalgo/dal"
)

//go:embed root-collections.json
var rootRegistryJSON []byte

// RootRegistry separates frozen legacy observation names from the exact legacy
// roots still eligible for writes. Neither list approves new root placement.
type RootRegistry struct {
	SchemaVersion       int      `json:"schemaVersion"`
	Project             string   `json:"project"`
	Database            string   `json:"database"`
	ApprovedRoots       []string `json:"approvedRoots"`
	LegacyRoots         []string `json:"legacyRoots"`
	LegacyWritableRoots []string `json:"legacyWritableRoots"`
	Extensions          []string `json:"extensions"`
	// SpaceOnlyExtensions restricts registered extensions to their space-owned
	// paths; it does not grant global or user-owned placement.
	SpaceOnlyExtensions  []string `json:"spaceOnlyExtensions,omitempty"`
	LegacySourceSnapshot string   `json:"legacySourceSnapshot"`
	LegacyRemovalGate    string   `json:"legacyRemovalGate"`
}

var newRootAccessPolicy = access.NewPolicy

var segmentName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]*$`)

// LoadRootRegistry validates the complete artifact; unknown fields, duplicate
// names and approved/legacy overlap fail closed.
func LoadRootRegistry(data []byte) (RootRegistry, error) {
	var r RootRegistry
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&r); err != nil {
		return r, fmt.Errorf("root registry: %w", err)
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF {
		return r, fmt.Errorf("root registry: trailing JSON")
	}
	if r.SchemaVersion != 2 || r.Project == "" || r.Database != "(default)" || len(r.ApprovedRoots) == 0 || len(r.Extensions) == 0 || r.LegacySourceSnapshot == "" || r.LegacyRemovalGate == "" {
		return r, fmt.Errorf("root registry: incomplete deployment identity")
	}
	seen := map[string]bool{}
	for _, group := range [][]string{r.ApprovedRoots, r.LegacyRoots} {
		for _, name := range group {
			if !segmentName.MatchString(name) || seen[name] {
				return r, fmt.Errorf("root registry: invalid or duplicate root %q", name)
			}
			seen[name] = true
		}
	}
	if r.LegacyWritableRoots == nil {
		return r, fmt.Errorf("root registry: legacyWritableRoots is required")
	}
	writable := map[string]bool{}
	for _, name := range r.LegacyWritableRoots {
		if !segmentName.MatchString(name) || writable[name] || !containsRoot(r.LegacyRoots, name) {
			return r, fmt.Errorf("root registry: invalid, duplicate or non-legacy writable root %q", name)
		}
		writable[name] = true
	}
	if !containsRoot(r.ApprovedRoots, "ext") {
		return r, fmt.Errorf("root registry: ext namespace must be approved")
	}
	seen = map[string]bool{}
	for _, name := range r.Extensions {
		if !segmentName.MatchString(name) || seen[name] {
			return r, fmt.Errorf("root registry: invalid or duplicate extension %q", name)
		}
		seen[name] = true
	}
	spaceOnly := map[string]bool{}
	for _, name := range r.SpaceOnlyExtensions {
		if !seen[name] || spaceOnly[name] {
			return r, fmt.Errorf("root registry: unknown or duplicate space-only extension %q", name)
		}
		spaceOnly[name] = true
	}
	return r, nil
}

// DefaultRootRegistry returns a fresh registry and its exact artifact digest.
func DefaultRootRegistry() (RootRegistry, string, error) {
	r, err := LoadRootRegistry(rootRegistryJSON)
	sum := sha256.Sum256(rootRegistryJSON)
	return r, hex.EncodeToString(sum[:]), err
}

// RootPolicy guards placement only. Existing caller/tenant authorization must
// remain in force; read operations pass through without adding an access grant.
func (r RootRegistry) RootPolicy() (access.Policy, error) {
	// Revalidate caller-supplied values through the same strict artifact codec.
	data, _ := json.Marshal(r) // RootRegistry contains only JSON scalar/slice fields.
	r, err := LoadRootRegistry(data)
	if err != nil {
		return nil, err
	}
	mutations := access.Insert | access.Set | access.Update | access.Delete
	rules := []access.Rule{}
	for _, root := range append(append([]string{}, r.ApprovedRoots...), r.LegacyWritableRoots...) {
		if root != "ext" {
			rules = append(rules, access.Collection(root, access.Allow(mutations)))
		}
	}
	// A literal extension has higher structural specificity than the wildcard.
	rules = append(rules, access.Under(access.Path("spaces", access.AnyID, "ext", access.AnyID), access.Deny(mutations)), access.Under(access.Path("users", access.AnyID, "ext", access.AnyID), access.Deny(mutations)))
	for _, extension := range r.Extensions {
		rules = append(rules, access.Under(access.Path("spaces", access.AnyID, "ext", extension), access.Allow(mutations)))
		if !containsRoot(r.SpaceOnlyExtensions, extension) {
			rules = append(rules, access.Scope("ext", extension, access.Allow(mutations)), access.Under(access.Path("users", access.AnyID, "ext", extension), access.Allow(mutations)))
		}
	}
	policy, err := newRootAccessPolicy("sneat-firestore-root-placement", rules...)
	if err != nil {
		return nil, err
	}
	return rootPlacementPolicy{Policy: policy}, nil
}

type rootPlacementPolicy struct{ access.Policy }

func (p rootPlacementPolicy) Decide(ctx context.Context, r access.Request) access.Decision {
	if r.Operation == access.Get || r.Operation == access.Exists || r.Operation == access.Query {
		return access.Decision{Allowed: true, Operation: r.Operation, Policy: p.Name(), Effect: "allow", Explanation: "placement policy leaves read authorization to caller policies"}
	}
	return p.Policy.Decide(ctx, r)
}
func (p rootPlacementPolicy) Authorize(ctx context.Context, r access.Request) error {
	d := p.Decide(ctx, r)
	if d.Allowed {
		return nil
	}
	return &access.DeniedError{Decision: d}
}

// ProtectRootCollections attaches a mandatory policy which request-bound
// policies cannot widen. It does not unwrap or replace existing policies.
func ProtectRootCollections(db dal.DB) (dal.DB, error) {
	r, _, err := DefaultRootRegistry()
	if err != nil {
		return nil, err
	}
	policy, err := r.RootPolicy()
	if err != nil {
		return nil, err
	}
	return access.SecureDB(db, access.WithDatabasePolicies(policy))
}

// RootComparison contains metadata names only, never document fields or IDs.
type RootComparison struct {
	Unexpected []string `json:"unexpected"`
	Legacy     []string `json:"legacy"`
}

// CompareRoots distinguishes frozen legacy debt from newly unapproved roots.
func (r RootRegistry) CompareRoots(names []string) RootComparison {
	approved := map[string]bool{}
	legacy := map[string]bool{}
	for _, name := range r.ApprovedRoots {
		approved[name] = true
	}
	for _, name := range r.LegacyRoots {
		legacy[name] = true
	}
	result := RootComparison{Unexpected: []string{}, Legacy: []string{}}
	seen := map[string]bool{}
	for _, name := range names {
		if seen[name] {
			continue
		}
		seen[name] = true
		if approved[name] {
			continue
		}
		if legacy[name] {
			result.Legacy = append(result.Legacy, name)
		} else {
			result.Unexpected = append(result.Unexpected, name)
		}
	}
	sort.Strings(result.Unexpected)
	sort.Strings(result.Legacy)
	return result
}

func containsRoot(names []string, wanted string) bool {
	for _, name := range names {
		if name == wanted {
			return true
		}
	}
	return false
}
