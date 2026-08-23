package slugs_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dal-go/dalgo/adapters/dalgo2memory"
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

// TestClaim_RollbackLeavesNoOrphanClaim verifies
// AC:claim-and-record-commit-together: a caller transaction that claims a
// slug and then fails before commit should roll back, leaving no claim
// document — the slug free.
//
// This is proved against dalgo2memory's WithOptimisticConcurrency() mode
// (dal-go/dalgo v0.65.0): a transaction created in that mode buffers every
// read and write locally and never touches the shared engine until commit,
// which runs only if the callback returns nil (see
// runOptimisticReadwriteTransaction in adapters/dalgo2memory/optimistic.go).
// A callback that returns an error short-circuits before commit is even
// attempted, so Claim's tx.Insert — buffered, never applied — is discarded
// along with everything else the transaction did. This is genuine rollback,
// not merely "the caller saw an error": the assertion below reads the slug
// back from the database afterwards to confirm nothing was ever written.
func TestClaim_RollbackLeavesNoOrphanClaim(t *testing.T) {
	db := dalgo2memory.NewDB(dalgo2memory.WithOptimisticConcurrency())
	ctx := context.Background()

	simulatedErr := errors.New("simulated failure writing the caller's own record")
	err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		if _, err := slugs.Claim(ctx, tx, "test:space", "orphan-check", "space-1", "space"); err != nil {
			return err
		}
		return simulatedErr
	})
	if !errors.Is(err, simulatedErr) {
		t.Fatalf("expected the simulated failure to propagate out of RunReadwriteTransaction, got %v", err)
	}

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
