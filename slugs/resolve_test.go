package slugs_test

import (
	"context"
	"testing"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
	"github.com/sneat-co/sneat-go-core/slugs"
	"github.com/sneat-co/sneat-go-core/sneatcoretesting"
)

// countingGetter wraps a dal.DB and counts calls to Get, so
// TestResolve_ReturnsTargetInOneRead can prove Resolve performs exactly one
// document read rather than trusting a source-code read alone.
type countingGetter struct {
	dal.DB
	gets int
}

func (g *countingGetter) Get(ctx context.Context, rec record.Record) error {
	g.gets++
	return g.DB.Get(ctx, rec)
}

// TestResolve_ReturnsTargetInOneRead verifies
// AC:resolve-returns-target-in-one-read.
//
// "No query" is enforced structurally, not just by this test: Resolve's
// parameter is a dal.Getter, an interface exposing only Get and Exists, so
// there is no query method reachable from inside Resolve even in principle —
// see Resolve's signature in resolve.go. What this test proves dynamically
// is the other half of the AC: that resolving a claimed slug costs exactly
// one Get call and returns the right target.
func TestResolve_ReturnsTargetInOneRead(t *testing.T) {
	db := sneatcoretesting.NewMemoryDB()
	ctx := context.Background()

	err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := slugs.Claim(ctx, tx, "test:space", "Main Hall", "space-1", "space", slugs.WithClaimedBy("alex"))
		return err
	})
	if err != nil {
		t.Fatalf("claim: %v", err)
	}

	spy := &countingGetter{DB: db}
	info, err := slugs.Resolve(ctx, spy, "test:space", "main hall")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if spy.gets != 1 {
		t.Errorf("Resolve performed %d Get call(s), want exactly 1", spy.gets)
	}
	if info.TargetID != "space-1" || info.TargetKind != "space" {
		t.Errorf("resolved target = (%q, %q), want (space-1, space)", info.TargetID, info.TargetKind)
	}
	if info.ClaimedBy != "alex" {
		t.Errorf("resolved ClaimedBy = %q, want alex", info.ClaimedBy)
	}
	if info.Tombstoned {
		t.Error("a freshly claimed slug must not be reported as tombstoned")
	}
}

// TestResolve_UnclaimedAndTombstonedAreDistinguishable verifies
// AC:unclaimed-and-tombstoned-are-distinguishable: a never-claimed slug
// reports IsSlugNotFound, and a released slug resolves successfully with
// Tombstoned set rather than erroring.
func TestResolve_UnclaimedAndTombstonedAreDistinguishable(t *testing.T) {
	db := sneatcoretesting.NewMemoryDB()
	ctx := context.Background()

	err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		if _, err := slugs.Claim(ctx, tx, "test:space", "old-hall", "space-1", "space"); err != nil {
			return err
		}
		return slugs.Release(ctx, tx, "test:space", "old-hall")
	})
	if err != nil {
		t.Fatalf("claim+release: %v", err)
	}

	if _, err := slugs.Resolve(ctx, db, "test:space", "never-claimed"); !slugs.IsSlugNotFound(err) {
		t.Errorf("resolving a never-claimed slug: IsSlugNotFound(%v) = false, want true", err)
	}

	info, err := slugs.Resolve(ctx, db, "test:space", "old-hall")
	if err != nil {
		t.Fatalf("resolving a released slug must not error: %v", err)
	}
	if !info.Tombstoned {
		t.Error("a released slug must resolve with Tombstoned = true")
	}
	if info.TargetID != "space-1" {
		t.Errorf("a tombstone must still report what it used to point to, got TargetID=%q", info.TargetID)
	}
}
