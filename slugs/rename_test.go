package slugs_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/slugs"
	"github.com/sneat-co/sneat-go-core/sneatcoretesting"
)

// TestRename_RenamedSlugIsNotReusable verifies
// AC:renamed-slug-is-not-reusable: after a rename, the old slug cannot be
// claimed by a different target, and resolving it reports a tombstone
// naming the new slug.
func TestRename_RenamedSlugIsNotReusable(t *testing.T) {
	db := sneatcoretesting.NewMemoryDB()
	ctx := context.Background()

	err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := slugs.Claim(ctx, tx, "test:space", "st-marys-village-hall", "space-1", "space")
		return err
	})
	if err != nil {
		t.Fatalf("initial claim: %v", err)
	}

	err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := slugs.Rename(ctx, tx, "test:space", "st-marys-village-hall", "st-marys-hall", "space-1", "space")
		return err
	})
	if err != nil {
		t.Fatalf("rename: %v", err)
	}

	// A different target tries to claim the old, now-tombstoned slug.
	err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := slugs.Claim(ctx, tx, "test:space", "st-marys-village-hall", "space-2", "space")
		return err
	})
	if !slugs.IsSlugTaken(err) {
		t.Fatalf("re-claiming a tombstoned slug: IsSlugTaken(%v) = false, want true", err)
	}

	info, err := slugs.Resolve(ctx, db, "test:space", "st-marys-village-hall")
	if err != nil {
		t.Fatalf("resolving the tombstoned old slug must not error: %v", err)
	}
	if !info.Tombstoned {
		t.Fatal("old slug must resolve as tombstoned")
	}
	if info.SuccessorSlug != "st-marys-hall" {
		t.Errorf("tombstone successor = %q, want st-marys-hall", info.SuccessorSlug)
	}

	newInfo, err := slugs.Resolve(ctx, db, "test:space", "st-marys-hall")
	if err != nil {
		t.Fatalf("resolving the new slug: %v", err)
	}
	if newInfo.TargetID != "space-1" || newInfo.Tombstoned {
		t.Errorf("new slug should resolve live to space-1, got TargetID=%q Tombstoned=%v", newInfo.TargetID, newInfo.Tombstoned)
	}
}

// TestRename_NewSlugTakenLeavesOldUntouched pins the ordering guarantee
// documented on Rename: if the new slug is already claimed, Rename fails as
// taken and the old slug is never touched. Unlike the atomicity AC below,
// this is genuinely provable today — it only requires that Rename does not
// attempt the old-slug tombstone once the new-slug claim has failed, which
// has nothing to do with the backend's rollback support.
func TestRename_NewSlugTakenLeavesOldUntouched(t *testing.T) {
	db := sneatcoretesting.NewMemoryDB()
	ctx := context.Background()

	err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		if _, err := slugs.Claim(ctx, tx, "test:space", "old-hall", "space-1", "space"); err != nil {
			return err
		}
		_, err := slugs.Claim(ctx, tx, "test:space", "new-hall", "space-2", "space")
		return err
	})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := slugs.Rename(ctx, tx, "test:space", "old-hall", "new-hall", "space-1", "space")
		return err
	})
	if !slugs.IsSlugTaken(err) {
		t.Fatalf("renaming onto an already-claimed slug: IsSlugTaken(%v) = false, want true", err)
	}

	info, err := slugs.Resolve(ctx, db, "test:space", "old-hall")
	if err != nil {
		t.Fatalf("old slug should still resolve: %v", err)
	}
	if info.Tombstoned {
		t.Error("old slug must not be tombstoned when the rename's new-slug claim failed")
	}
}

// TestRename_IsAtomic is meant to verify AC:rename-is-atomic: if the
// caller's transaction fails after Rename has performed both writes, rolling
// back must leave the old slug live and the new slug unclaimed.
//
// It is SKIPPED for the same reason as
// TestClaim_RollbackLeavesNoOrphanClaim in claim_test.go (see that test's
// comment for the full explanation): dalgo2memory has no transaction
// rollback, so a rename's two writes are never undone once made, regardless
// of what the callback returns afterwards. This is a second, previously
// undocumented instance of the same backend gap the Feature already names
// for AC:concurrent-claims-yield-one-winner. Proving this AC needs a real
// transactional backend or rollback support landing in dalgo2memory.
func TestRename_IsAtomic(t *testing.T) {
	t.Skip("dalgo2memory has no transaction rollback, so a rename's writes are never undone when the transaction callback returns an error — see the comment on this test and TestClaim_RollbackLeavesNoOrphanClaim")

	db := sneatcoretesting.NewMemoryDB()
	ctx := context.Background()

	err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := slugs.Claim(ctx, tx, "test:space", "st-marys-village-hall", "space-1", "space")
		return err
	})
	if err != nil {
		t.Fatalf("initial claim: %v", err)
	}

	_ = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		if _, err := slugs.Rename(ctx, tx, "test:space", "st-marys-village-hall", "st-marys-hall", "space-1", "space"); err != nil {
			return err
		}
		return errors.New("simulated failure after both writes")
	})

	oldInfo, err := slugs.Resolve(ctx, db, "test:space", "st-marys-village-hall")
	if err != nil || oldInfo.Tombstoned {
		t.Errorf("after rollback the old slug should still resolve live: info=%+v err=%v", oldInfo, err)
	}
	if _, err := slugs.Resolve(ctx, db, "test:space", "st-marys-hall"); !slugs.IsSlugNotFound(err) {
		t.Errorf("after rollback the new slug should be unclaimed: %v", err)
	}
}
