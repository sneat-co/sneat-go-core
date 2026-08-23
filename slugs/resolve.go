package slugs

import (
	"context"
	"fmt"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
)

// Resolve looks up (namespace, slug) in a single document read with no
// query, per REQ:resolve. getter can be a dal.DB, a dal.ReadTransaction or a
// dal.ReadwriteTransaction — whichever the caller already has to hand — since
// all three satisfy dal.Getter.
//
// There are exactly three outcomes, and they are deliberately not
// squashed into one shape:
//
//   - The slug was never claimed: Resolve returns (nil, err) with
//     IsSlugNotFound(err) true. A caller routing a public URL turns this into
//     an HTTP 404.
//   - The slug is claimed and live: Resolve returns (info, nil) with
//     info.Tombstoned false. A caller routes to info.TargetID/info.TargetKind.
//   - The slug was released — by Release, or by the old side of a Rename:
//     Resolve returns (info, nil) with info.Tombstoned true and
//     info.SuccessorSlug set when the release was a rename. A caller issues a
//     redirect rather than a 404.
//
// Distinguishing "not found" from "tombstoned" by a field on a successfully
// returned value, rather than by two different errors, mirrors dalgo's own
// record.Record shape (Exists() is not itself an error) and means a caller
// cannot forget to check Tombstoned while handling only the error return.
func Resolve(ctx context.Context, getter dal.Getter, namespace Namespace, slug string) (*ClaimInfo, error) {
	if err := ValidateNamespace(namespace); err != nil {
		return nil, err
	}
	normalised := Normalize(slug)
	if normalised == "" {
		// A slug that normalises to nothing can never have been claimed:
		// Claim rejects it before it ever reaches storage. Reporting
		// not-found here (rather than a different error) keeps Resolve's
		// contract simple: every input is either not-found, tombstoned, or
		// live.
		return nil, fmt.Errorf("%w: %q in namespace %q", ErrSlugNotFound, slug, namespace)
	}

	data := &claimData{}
	rec := record.NewRecordWithData(claimKey(namespace, Slug(normalised)), data)
	if err := getter.Get(ctx, rec); err != nil {
		if record.IsNotFound(err) {
			return nil, fmt.Errorf("%w: %q in namespace %q", ErrSlugNotFound, normalised, namespace)
		}
		return nil, err
	}
	return data.toInfo(namespace, Slug(normalised)), nil
}
