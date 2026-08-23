package slugs

import (
	"context"
	"fmt"

	"github.com/dal-go/dalgo/dal"
)

// Rename claims newSlug and tombstones oldSlug as one operation inside tx,
// per REQ:release-leaves-tombstone's "rename MUST be a single operation that
// claims the new slug and tombstones the old one in the caller's
// transaction, so a rename cannot half-happen."
//
// The two writes happen in an order that matters: the new slug is claimed
// first, and the old slug is only tombstoned if that succeeds. If newSlug is
// already taken, Rename returns ErrSlugTaken and oldSlug is left completely
// untouched — still live, still resolving to its current target. Whether the
// old-slug tombstone write itself succeeds or fails, the caller's
// transaction is what makes the whole thing atomic: Rename does not open or
// manage a transaction of its own, so if either write fails and the caller
// returns that error from its RunReadwriteTransaction callback, the backend
// rolls back both writes together and the rename never partially happened.
//
// The tombstone left on oldSlug records newSlug as its successor, so
// resolving oldSlug afterwards reports a tombstone carrying the new slug
// rather than a bare "not found" — see Resolve.
func Rename(ctx context.Context, tx dal.ReadwriteTransaction, namespace Namespace, oldSlug, newSlug, targetID string, targetKind TargetKind, opts ...Option) (Slug, error) {
	if err := ValidateNamespace(namespace); err != nil {
		return "", err
	}
	normalisedOld := Normalize(oldSlug)
	if normalisedOld == "" {
		return "", fmt.Errorf("slugs: old slug %q normalises to an empty string", oldSlug)
	}

	claimed, err := Claim(ctx, tx, namespace, newSlug, targetID, targetKind, opts...)
	if err != nil {
		return "", err
	}

	if err := tombstone(ctx, tx, namespace, Slug(normalisedOld), claimed); err != nil {
		return "", err
	}
	return claimed, nil
}
