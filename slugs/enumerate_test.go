package slugs_test

import (
	"context"
	"sort"
	"testing"

	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/slugs"
	"github.com/sneat-co/sneat-go-core/sneatcoretesting"
)

// TestEnumerate_NamespaceCanBeEnumerated verifies
// AC:namespace-can-be-enumerated: every claim in a namespace comes back,
// tombstones are distinguishable from live claims, and a claim living in a
// different namespace is excluded.
func TestEnumerate_NamespaceCanBeEnumerated(t *testing.T) {
	db := sneatcoretesting.NewMemoryDB()
	ctx := context.Background()

	err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		if _, err := slugs.Claim(ctx, tx, "test:space", "main-hall", "space-1", "space"); err != nil {
			return err
		}
		if _, err := slugs.Claim(ctx, tx, "test:space", "small-room", "space-2", "space"); err != nil {
			return err
		}
		if err := slugs.Release(ctx, tx, "test:space", "small-room"); err != nil {
			return err
		}
		// Lives in a different namespace: must not appear in the listing below.
		_, err := slugs.Claim(ctx, tx, "other:space", "main-hall", "space-3", "space")
		return err
	})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	claims, err := slugs.Enumerate(ctx, db, "test:space")
	if err != nil {
		t.Fatalf("enumerate: %v", err)
	}
	if len(claims) != 2 {
		t.Fatalf("Enumerate returned %d claim(s), want 2: %+v", len(claims), claims)
	}

	sort.Slice(claims, func(i, j int) bool { return claims[i].Slug < claims[j].Slug })

	if claims[0].Slug != "main-hall" || claims[0].Tombstoned || claims[0].TargetID != "space-1" {
		t.Errorf("claims[0] = %+v, want live main-hall -> space-1", claims[0])
	}
	if claims[1].Slug != "small-room" || !claims[1].Tombstoned || claims[1].TargetID != "space-2" {
		t.Errorf("claims[1] = %+v, want tombstoned small-room (still pointing at space-2)", claims[1])
	}

	// The other namespace's claim was never returned above; confirm it is
	// still independently resolvable, so "excluded from this listing" is not
	// hiding "never actually written".
	if _, err := slugs.Resolve(ctx, db, "other:space", "main-hall"); err != nil {
		t.Errorf("other:space's own claim should resolve fine: %v", err)
	}
}
