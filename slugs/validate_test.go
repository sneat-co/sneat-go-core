package slugs_test

import (
	"context"
	"testing"

	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/slugs"
	"github.com/sneat-co/sneat-go-core/sneatcoretesting"
)

// TestValidateNamespace pins REQ:namespace-is-opaque's storage-level
// validation: non-empty, and safe as a document ID (no reserved '%' escape
// marker — see record.ValidateStringID).
func TestValidateNamespace(t *testing.T) {
	for _, ns := range []slugs.Namespace{"communitycentrum:space", "bookius:bookingType:space-1", "a"} {
		if err := slugs.ValidateNamespace(ns); err != nil {
			t.Errorf("ValidateNamespace(%q) = %v, want nil", ns, err)
		}
	}
	for _, ns := range []slugs.Namespace{"", "has%percent"} {
		if err := slugs.ValidateNamespace(ns); err == nil {
			t.Errorf("ValidateNamespace(%q) = nil, want an error", ns)
		}
	}
}

// TestClaim_RejectsEmptyNamespaceAndSlug exercises the two defensive
// validation paths that are not any single AC on their own, but that every
// exported entry point relies on: an empty/unsafe namespace, and a slug that
// normalises to nothing (e.g. all punctuation).
func TestClaim_RejectsEmptyNamespaceAndSlug(t *testing.T) {
	db := sneatcoretesting.NewMemoryDB()
	ctx := context.Background()

	err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := slugs.Claim(ctx, tx, "", "main-hall", "space-1", "space")
		return err
	})
	if err == nil {
		t.Error("Claim with an empty namespace should fail")
	}

	err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := slugs.Claim(ctx, tx, "test:space", "!!!", "space-1", "space")
		return err
	})
	if err == nil {
		t.Error("Claim with a slug that normalises to empty should fail")
	}
}

// TestRejectsEmptyNamespaceAndSlug_AllEntryPoints exercises the same
// defensive validation as TestClaim_RejectsEmptyNamespaceAndSlug, but for
// Release, Rename, Resolve and Enumerate — every exported function shares
// the same ValidateNamespace/Normalize guard at its top, and each one's
// guard is its own code path worth pinning independently of the others.
func TestRejectsEmptyNamespaceAndSlug_AllEntryPoints(t *testing.T) {
	db := sneatcoretesting.NewMemoryDB()
	ctx := context.Background()

	if _, err := slugs.Resolve(ctx, db, "", "main-hall"); err == nil {
		t.Error("Resolve with an empty namespace should fail")
	}
	if _, err := slugs.Resolve(ctx, db, "test:space", "!!!"); !slugs.IsSlugNotFound(err) {
		t.Errorf("Resolve with a slug that normalises to empty: IsSlugNotFound(%v) = false, want true", err)
	}
	if _, err := slugs.Enumerate(ctx, db, ""); err == nil {
		t.Error("Enumerate with an empty namespace should fail")
	}

	err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return slugs.Release(ctx, tx, "", "main-hall")
	})
	if err == nil {
		t.Error("Release with an empty namespace should fail")
	}
	err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return slugs.Release(ctx, tx, "test:space", "!!!")
	})
	if err == nil {
		t.Error("Release with a slug that normalises to empty should fail")
	}

	err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := slugs.Rename(ctx, tx, "", "old-hall", "new-hall", "space-1", "space")
		return err
	})
	if err == nil {
		t.Error("Rename with an empty namespace should fail")
	}
	err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := slugs.Rename(ctx, tx, "test:space", "!!!", "new-hall", "space-1", "space")
		return err
	})
	if err == nil {
		t.Error("Rename with an old slug that normalises to empty should fail")
	}
}

// TestRelease_NeverClaimedReportsNotFound documents Release's behaviour when
// asked to tombstone a slug nothing ever claimed: it fails as not-found
// rather than silently creating a tombstone for a claim that never existed.
// This is not itself an AC, but it is the natural consequence of Release
// being a partial-field update (never a read-then-write) that this package
// relies on elsewhere, so it earns a test of its own.
func TestRelease_NeverClaimedReportsNotFound(t *testing.T) {
	db := sneatcoretesting.NewMemoryDB()
	ctx := context.Background()

	err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return slugs.Release(ctx, tx, "test:space", "never-claimed")
	})
	if !slugs.IsSlugNotFound(err) {
		t.Errorf("releasing a never-claimed slug: IsSlugNotFound(%v) = false, want true", err)
	}
}
