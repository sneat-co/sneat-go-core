package slugs_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/slugs"
	"github.com/sneat-co/sneat-go-core/sneatcoretesting"
)

// TestClaim_SecondClaimIsRefused verifies AC:second-claim-is-refused: the
// first claim of a slug succeeds, and a second attempt at the identical
// normalised slug in the same namespace fails with an error identifiable by
// IsSlugTaken.
func TestClaim_SecondClaimIsRefused(t *testing.T) {
	db := sneatcoretesting.NewMemoryDB()
	ctx := context.Background()

	err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := slugs.Claim(ctx, tx, "test:space", "Main Hall", "space-1", "space")
		return err
	})
	if err != nil {
		t.Fatalf("first claim: unexpected error: %v", err)
	}

	err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := slugs.Claim(ctx, tx, "test:space", "Main Hall", "space-2", "space")
		return err
	})
	if err == nil {
		t.Fatal("second claim: expected an error, got nil")
	}
	if !slugs.IsSlugTaken(err) {
		t.Errorf("second claim: IsSlugTaken(%v) = false, want true", err)
	}
}

// TestClaim_RollbackLeavesNoOrphanClaim is meant to verify
// AC:claim-and-record-commit-together: a caller transaction that claims a
// slug and then fails before commit should roll back, leaving no claim
// document — the slug free.
//
// It is SKIPPED. dalgo2memory, the only backend available to this repo's
// tests, does not implement transaction rollback at all: a write made
// through tx.Insert/tx.Update is applied to the shared in-memory store the
// moment it is called, and returning a non-nil error from the
// RunReadwriteTransaction callback does not undo it afterwards. This was
// confirmed directly against the adapter before writing this test: a
// transaction that inserts a record and then returns an error leaves that
// record present, Exists()==true, once RunReadwriteTransaction returns.
//
// This is a second, previously undocumented instance of the same class of
// gap the Feature already names for AC:concurrent-claims-yield-one-winner —
// that one is the adapter's global lock preventing two transactions from
// ever contending; this one is that the adapter has no commit/rollback
// boundary at all, so there is nothing for a returned error to roll back.
// Both stem from dalgo2memory being an intentionally minimal in-memory
// stand-in rather than a faithful transactional emulator.
//
// What this package can and does guarantee today is structural, not
// runtime: Claim/Release/Rename all take a dal.ReadwriteTransaction, never a
// dal.DB, so there is no code path in this package that could open or commit
// a transaction of its own — see the exported function signatures. Proving
// the other half (that a backend's failed commit actually discards the
// prior writes) needs either a real transactional backend (a Firestore
// emulator, a SQL adapter) or dal-go/dalgo growing rollback support for
// dalgo2memory, which the spec already tracks alongside the concurrency-mode
// work.
func TestClaim_RollbackLeavesNoOrphanClaim(t *testing.T) {
	t.Skip("dalgo2memory has no transaction rollback: a write made via tx.Insert/tx.Update is never undone when the transaction callback returns an error, so this AC cannot be proven against it — see the comment on this test")

	db := sneatcoretesting.NewMemoryDB()
	ctx := context.Background()

	_ = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		if _, err := slugs.Claim(ctx, tx, "test:space", "orphan-check", "space-1", "space"); err != nil {
			return err
		}
		return errors.New("simulated failure writing the caller's own record")
	})

	if _, err := slugs.Resolve(ctx, db, "test:space", "orphan-check"); !slugs.IsSlugNotFound(err) {
		t.Errorf("resolving after a rolled-back claim: IsSlugNotFound(%v) = false, want true", err)
	}
}

// TestClaim_NamespacesAreIndependent verifies AC:namespaces-are-independent:
// the same slug claimed in two different namespaces succeeds in both and
// each resolves to its own target.
func TestClaim_NamespacesAreIndependent(t *testing.T) {
	db := sneatcoretesting.NewMemoryDB()
	ctx := context.Background()

	err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		if _, err := slugs.Claim(ctx, tx, "a:space", "main-hall", "space-a", "space"); err != nil {
			return err
		}
		_, err := slugs.Claim(ctx, tx, "b:space", "main-hall", "space-b", "space")
		return err
	})
	if err != nil {
		t.Fatalf("claiming the same slug in two namespaces: %v", err)
	}

	infoA, err := slugs.Resolve(ctx, db, "a:space", "main-hall")
	if err != nil {
		t.Fatalf("resolve a:space: %v", err)
	}
	if infoA.TargetID != "space-a" {
		t.Errorf("a:space resolved to %q, want space-a", infoA.TargetID)
	}

	infoB, err := slugs.Resolve(ctx, db, "b:space", "main-hall")
	if err != nil {
		t.Fatalf("resolve b:space: %v", err)
	}
	if infoB.TargetID != "space-b" {
		t.Errorf("b:space resolved to %q, want space-b", infoB.TargetID)
	}
}

// TestClaim_EquivalentInputsNormaliseTogether verifies
// AC:equivalent-inputs-normalise-together using the Feature's own example: a
// slug claimed as "St--Marys Village Hall " refuses a second claim spelled
// "st-marys-village-hall", and both forms resolve to the same target.
func TestClaim_EquivalentInputsNormaliseTogether(t *testing.T) {
	db := sneatcoretesting.NewMemoryDB()
	ctx := context.Background()

	err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := slugs.Claim(ctx, tx, "test:space", "St--Marys Village Hall ", "space-1", "space")
		return err
	})
	if err != nil {
		t.Fatalf("first claim: %v", err)
	}

	err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := slugs.Claim(ctx, tx, "test:space", "st-marys-village-hall", "space-2", "space")
		return err
	})
	if !slugs.IsSlugTaken(err) {
		t.Fatalf("second claim of the equivalent slug: IsSlugTaken(%v) = false, want true", err)
	}

	infoRaw, err := slugs.Resolve(ctx, db, "test:space", "St--Marys Village Hall ")
	if err != nil {
		t.Fatalf("resolve raw form: %v", err)
	}
	infoNormalised, err := slugs.Resolve(ctx, db, "test:space", "st-marys-village-hall")
	if err != nil {
		t.Fatalf("resolve normalised form: %v", err)
	}
	if infoRaw.TargetID != "space-1" || infoNormalised.TargetID != "space-1" {
		t.Errorf("both forms should resolve to space-1, got %q and %q", infoRaw.TargetID, infoNormalised.TargetID)
	}
}

// TestClaim_ReservedSlugIsRefusedDistinctly verifies
// AC:reserved-slug-is-refused-distinctly: a base reserved word is refused
// with IsSlugReserved (and never IsSlugTaken), the reserved-word rejection
// never reaches storage, and a namespace-specific extra reserved word
// behaves identically.
func TestClaim_ReservedSlugIsRefusedDistinctly(t *testing.T) {
	db := sneatcoretesting.NewMemoryDB()
	ctx := context.Background()

	err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := slugs.Claim(ctx, tx, "test:space", "new", "space-1", "space")
		return err
	})
	if !slugs.IsSlugReserved(err) {
		t.Fatalf("claiming a base-reserved word: IsSlugReserved(%v) = false, want true", err)
	}
	if slugs.IsSlugTaken(err) {
		t.Errorf("a reserved-word rejection must not also read as IsSlugTaken(%v)", err)
	}
	if _, err := slugs.Resolve(ctx, db, "test:space", "new"); !slugs.IsSlugNotFound(err) {
		t.Errorf("a rejected reserved-word claim must not have written anything: %v", err)
	}

	err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := slugs.Claim(ctx, tx, "test:bookingType", "default", "bt-1", "bookingType", slugs.WithReservedWords("default"))
		return err
	})
	if !slugs.IsSlugReserved(err) {
		t.Fatalf("claiming a namespace-extended reserved word: IsSlugReserved(%v) = false, want true", err)
	}

	// A namespace that did not extend the list with "default" is unaffected:
	// the extension is per-call, not global.
	err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := slugs.Claim(ctx, tx, "test:other", "default", "o-1", "other")
		return err
	})
	if err != nil {
		t.Errorf("\"default\" should not be reserved outside the namespace that extended the list: %v", err)
	}
}
