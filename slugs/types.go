package slugs

import (
	"fmt"
	"time"

	"github.com/dal-go/record"
)

// namespacesCollection and claimsCollection together spell out the storage
// path fixed by Decision 0001: /slugs/<namespace>/claims/<slug>. The nested
// shape — namespace as a parent document, claims as its subcollection — is
// what makes Enumerate an ordinary subcollection listing instead of needing a
// secondary index. A flat "/slugs/<namespace>:<slug>" collection was
// considered and declined for exactly this reason; see the Decision's
// "Declined Alternatives".
const (
	namespacesCollection = "slugs"
	claimsCollection     = "claims"
)

// Namespace scopes slug uniqueness. It is an arbitrary caller-supplied
// string, not a fixed concept — see the package doc's "Namespaces are
// opaque". Two claims of the same slug in different namespaces never
// collide: "communitycentrum:space" is a global scope, "bookius:bookingType:
// <spaceID>" is scoped to one space, "communitycentrum:room:<spaceID>" is
// scoped to one venue, and this package cannot tell the difference between
// them — that is the point.
type Namespace string

// Slug is a normalised slug: lowercased, Unicode-normalised, with runs of
// separators collapsed to a single hyphen and none leading or trailing. Every
// exported function that accepts a raw string slug normalises it internally,
// so a Slug value returned by this package (e.g. from Claim or Rename) is
// always already in the form that gets stored and that should appear in a
// URL.
type Slug string

// TargetKind discriminates what a claim's TargetID identifies (e.g. "space",
// "bookingType", "room"). It exists because Resolve reports both in one read
// with no query: a caller routing a public URL needs to know not just which
// record a slug points to, but what kind of record it is, before it can look
// the record up.
type TargetKind string

// ClaimInfo is the read view of a stored claim, returned by Resolve (for one
// slug) and Enumerate (for every claim in a namespace). It is not the wire
// format persisted to storage — see claimData for that — because a released
// claim (a tombstone) exposes a different, richer shape to callers than a
// live one does, and giving both a single reported type is what lets a
// caller distinguish them with a single boolean rather than a type switch.
type ClaimInfo struct {
	// Namespace and Slug identify which claim this is. Enumerate needs both
	// filled in to be useful on its own (a caller iterating results has no
	// other way to know which slug a given entry names); Resolve fills them
	// in too so a ClaimInfo is self-describing regardless of which function
	// produced it.
	Namespace Namespace
	Slug      Slug

	// TargetID and TargetKind identify what the slug points to. These are
	// never cleared by Release or Rename: a tombstone's whole job is to
	// remember what it used to point to, per REQ:release-leaves-tombstone.
	TargetID   string
	TargetKind TargetKind

	// ClaimedAt and ClaimedBy record when and (optionally) by whom the slug
	// was claimed. ClaimedBy is caller-defined and never interpreted by this
	// package — see the Feature's open question on whether only the holder
	// should be allowed to release a claim, deliberately left unresolved.
	ClaimedAt time.Time
	ClaimedBy string

	// Tombstoned reports whether this claim has been released (by Release or
	// by the "old slug" side of a Rename) rather than deleted. A tombstoned
	// claim's document still exists precisely so it can never be claimed by
	// a different target; see the package doc's "Release leaves a
	// tombstone".
	Tombstoned bool

	// ReleasedAt is the zero time when Tombstoned is false, and the moment
	// of release otherwise.
	ReleasedAt time.Time

	// SuccessorSlug is set when Rename tombstoned this slug in favour of a
	// new one, so a caller resolving the old slug can issue a redirect
	// rather than a 404. It is empty for a live claim and for a claim
	// released outright (its target was deleted, not renamed).
	SuccessorSlug Slug
}

// claimData is the JSON/Firestore wire shape stored at
// /slugs/<namespace>/claims/<slug>. It intentionally mirrors the minimal
// shape from Decision 0001 ({targetID, targetKind, claimedAt, claimedBy}) plus
// the two fields a tombstone needs (releasedAt, successorSlug), which stay
// absent on a live claim rather than being pre-declared with zero values —
// ReleasedAt is a pointer so `omitempty` actually omits it (a zero time.Time
// is not "empty" to encoding/json).
type claimData struct {
	TargetID   string     `json:"targetID" firestore:"targetID"`
	TargetKind TargetKind `json:"targetKind" firestore:"targetKind"`
	ClaimedAt  time.Time  `json:"claimedAt" firestore:"claimedAt"`
	ClaimedBy  string     `json:"claimedBy,omitempty" firestore:"claimedBy,omitempty"`

	// ReleasedAt/SuccessorSlug are the tombstone. TargetID/TargetKind above
	// are deliberately left untouched when these are set, so a tombstone
	// still answers "what did this used to point to" — see
	// REQ:release-leaves-tombstone.
	ReleasedAt    *time.Time `json:"releasedAt,omitempty" firestore:"releasedAt,omitempty"`
	SuccessorSlug string     `json:"successorSlug,omitempty" firestore:"successorSlug,omitempty"`
}

func (d *claimData) toInfo(namespace Namespace, slug Slug) *ClaimInfo {
	info := &ClaimInfo{
		Namespace:  namespace,
		Slug:       slug,
		TargetID:   d.TargetID,
		TargetKind: d.TargetKind,
		ClaimedAt:  d.ClaimedAt,
		ClaimedBy:  d.ClaimedBy,
	}
	if d.ReleasedAt != nil {
		info.Tombstoned = true
		info.ReleasedAt = *d.ReleasedAt
		info.SuccessorSlug = Slug(d.SuccessorSlug)
	}
	return info
}

// ValidateNamespace checks that namespace satisfies REQ:namespace-is-opaque's
// storage requirement: non-empty and safe as a document ID. It does not, and
// cannot, validate that the namespace means what the caller thinks it means —
// namespaces are opaque strings by design, so "safe to store" is the only
// universal check available.
func ValidateNamespace(namespace Namespace) error {
	if namespace == "" {
		return fmt.Errorf("slugs: namespace is empty")
	}
	if err := record.ValidateStringID(string(namespace)); err != nil {
		return fmt.Errorf("slugs: namespace %q is not safe as a document ID: %w", namespace, err)
	}
	return nil
}

// namespaceKey builds the parent key for a namespace's claims subcollection:
// /slugs/<namespace>.
func namespaceKey(namespace Namespace) *record.Key {
	return record.NewKeyWithID(namespacesCollection, string(namespace))
}

// claimKey builds the key of one claim document:
// /slugs/<namespace>/claims/<slug>. slug must already be normalised — every
// caller of claimKey in this package normalises first, so that a claim's key
// and the slug reported back to the caller are always the same string.
func claimKey(namespace Namespace, slug Slug) *record.Key {
	return record.NewKeyWithParentAndID(namespaceKey(namespace), claimsCollection, string(slug))
}
