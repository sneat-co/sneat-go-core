package slugs

import (
	"context"
	"fmt"
	"time"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
	"github.com/dal-go/record/update"
)

// Release tombstones a claim — on rename, or when the target it names is
// deleted — rather than deleting its document, per
// REQ:release-leaves-tombstone. After Release, the slug resolves as
// tombstoned (Resolve reports Tombstoned=true) and can never be claimed by a
// different target: the document still exists, so Claim's insert still fails
// against it.
//
// Release is a single partial-field update (tx.Update), not a
// read-then-rewrite: it sets only ReleasedAt, leaving TargetID and
// TargetKind exactly as they were, which is how a tombstone keeps answering
// "what did this used to point to". That also means Release works cleanly
// under a strict no-reads-after-writes transaction (the shape Firestore
// enforces), since it never reads at all.
//
// Like Claim, tx is caller-supplied: releasing a slug typically happens
// alongside deleting or renaming the record it names, and those two writes
// must commit together.
func Release(ctx context.Context, tx dal.ReadwriteTransaction, namespace Namespace, slug string) error {
	if err := ValidateNamespace(namespace); err != nil {
		return err
	}
	normalised := Normalize(slug)
	if normalised == "" {
		return fmt.Errorf("slugs: slug %q normalises to an empty string", slug)
	}
	return tombstone(ctx, tx, namespace, Slug(normalised), "")
}

// tombstone applies the release update shared by Release and Rename (the
// "tombstone the old slug" half of a rename). successor is the empty string
// for a plain Release and the new slug for a Rename.
func tombstone(ctx context.Context, tx dal.ReadwriteTransaction, namespace Namespace, slug Slug, successor Slug) error {
	updates := []update.Update{
		update.ByFieldName("releasedAt", time.Now().UTC()),
	}
	if successor != "" {
		updates = append(updates, update.ByFieldName("successorSlug", string(successor)))
	}
	key := claimKey(namespace, slug)
	if err := tx.Update(ctx, key, updates); err != nil {
		if record.IsNotFound(err) {
			return fmt.Errorf("%w: %q in namespace %q", ErrSlugNotFound, slug, namespace)
		}
		return err
	}
	return nil
}
