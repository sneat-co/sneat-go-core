package slugs

import (
	"context"
	"fmt"
	"time"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
)

// Claim takes a slug for (targetID, targetKind) inside tx, per
// REQ:claim-is-atomic and REQ:claim-joins-the-caller-transaction.
//
// tx is supplied by the caller, not opened by this package, so that Claim can
// be called alongside — in the same transaction as — the write that creates
// the record the slug names. That is the whole point: a claim without its
// record, or a record without its claim, is exactly the defect this package
// exists to make impossible. A typical call looks like:
//
//	err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
//	    slug, err := slugs.Claim(ctx, tx, "communitycentrum:space", rawSlug, spaceID, "space")
//	    if err != nil {
//	        return err // rolls back the transaction; no claim and no record are left behind
//	    }
//	    space.Slug = slug // REQ: the slug is stored on the record too, for slug -> URL rendering
//	    return tx.Insert(ctx, spaceRecord)
//	})
//
// slug is normalised internally (see Normalize) before it is checked or
// claimed, and the normalised Slug is returned on success so the caller can
// store the exact string that was claimed rather than re-deriving it.
//
// Claim never reads before it writes: the only database operation it performs
// is a single tx.Insert at the claim's fixed document ID, which is what makes
// the claim atomic without depending on a read.
//
// A failed insert is reported as ErrSlugTaken only when the adapter identifies
// it as a duplicate key, via record.IsAlreadyExists. Any other failure is
// returned as itself. The distinction matters to end users: reporting every
// insert failure as "taken" tells someone their chosen name is unavailable
// when the real problem was a deadline or a permission error, and sends them
// off to invent a different name for no reason.
//
// Note the asymmetry: record.IsAlreadyExists returning false is NOT proof that
// the key was free. An adapter that does not classify its duplicate-key errors
// never returns the sentinel, so against one of those a genuine conflict
// surfaces as a plain error rather than ErrSlugTaken. That is the safer
// direction to fail — an honest error beats a confident wrong answer — and it
// is a visible, tracked state rather than a silent one, because dalgo's shared
// conformance suite has an unconditional check for exactly this behaviour.
func Claim(ctx context.Context, tx dal.ReadwriteTransaction, namespace Namespace, slug, targetID string, targetKind TargetKind, opts ...Option) (Slug, error) {
	if err := ValidateNamespace(namespace); err != nil {
		return "", err
	}
	normalised := Normalize(slug)
	if normalised == "" {
		return "", fmt.Errorf("slugs: slug %q normalises to an empty string", slug)
	}

	o := newOptions(opts)
	if isReserved(normalised, o.reservedWords) {
		return "", fmt.Errorf("%w: %q", ErrSlugReserved, normalised)
	}

	data := &claimData{
		TargetID:   targetID,
		TargetKind: targetKind,
		ClaimedAt:  time.Now().UTC(),
		ClaimedBy:  o.claimedBy,
	}
	rec := record.NewRecordWithData(claimKey(namespace, Slug(normalised)), data)
	if err := tx.Insert(ctx, rec); err != nil {
		if record.IsAlreadyExists(err) {
			return "", fmt.Errorf("%w: %q in namespace %q: %w", ErrSlugTaken, normalised, namespace, err)
		}
		return "", fmt.Errorf("slugs: claiming %q in namespace %q: %w", normalised, namespace, err)
	}
	return Slug(normalised), nil
}
