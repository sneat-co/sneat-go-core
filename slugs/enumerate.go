package slugs

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
)

// Enumerate lists every claim in namespace — live and tombstoned alike — per
// REQ:namespace-is-opaque's "MUST be able to enumerate every claim in a
// namespace, which admin surfaces and migrations both need." Each returned
// ClaimInfo's Tombstoned field distinguishes a released slug from a live one,
// same as Resolve.
//
// executor can be a dal.DB, a dal.ReadSession, a dal.ReadTransaction or a
// dal.ReadwriteTransaction — whichever the caller already has to hand —
// since all of them satisfy dal.QueryExecutor. This is an ordinary
// subcollection query (/slugs/<namespace>/claims/*) rather than a query over
// a secondary index, which is exactly what the nested storage shape from
// Decision 0001 buys over a flat "namespace:slug" collection.
func Enumerate(ctx context.Context, executor dal.QueryExecutor, namespace Namespace) ([]ClaimInfo, error) {
	if err := ValidateNamespace(namespace); err != nil {
		return nil, err
	}

	source := dal.NewCollectionRef(claimsCollection, "", namespaceKey(namespace))
	query := dal.From(source).NewQuery().SelectIntoRecord(func() record.Record {
		// The key here is a template the executor discards in favour of each
		// row's real stored key (which is what carries the parent/namespace);
		// only Data() from this factory is actually used. An incomplete key
		// avoids having to invent a placeholder slug just to satisfy the
		// factory signature.
		return record.NewRecordWithIncompleteKey(claimsCollection, reflect.String, &claimData{})
	})

	reader, err := executor.ExecuteQueryToRecordsReader(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("slugs: enumerating namespace %q: %w", namespace, err)
	}
	defer func() {
		_ = reader.Close()
	}()

	var claims []ClaimInfo
	for {
		rec, err := reader.Next()
		if err != nil {
			if errors.Is(err, dal.ErrNoMoreRecords) {
				break
			}
			return nil, fmt.Errorf("slugs: enumerating namespace %q: %w", namespace, err)
		}
		slug := Slug(fmt.Sprintf("%v", rec.Key().ID))
		data := rec.Data().(*claimData)
		claims = append(claims, *data.toInfo(namespace, slug))
	}
	return claims, nil
}
