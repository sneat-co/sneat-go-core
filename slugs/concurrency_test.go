package slugs_test

import (
	"context"
	"testing"

	"github.com/dal-go/dalgo/adapters/dalgo2memory"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/slugs"
)

// TestClaim_ConcurrentClaimsYieldOneWinner verifies
// AC:concurrent-claims-yield-one-winner: two transactions racing to claim
// the same slug must settle with exactly one commit, and no observer can
// ever see both holding the claim.
//
// This is proved against dalgo2memory's WithOptimisticConcurrency() mode
// (dal-go/dalgo v0.65.0): unlike the package's default mode, that mode lets
// two RunReadwriteTransaction calls genuinely interleave — each buffers its
// reads and writes locally and touches no shared storage until commit, where
// it is validated against every key it touched. Claim's only operation is a
// single tx.Insert, so racing it against itself races Insert's own
// touch-then-stage step; forcing both transactions past that step (via
// bothClaimed/proceed) before either is allowed to return from its callback
// and commit reproduces the exact interleaving
// TestOptimisticConcurrencyGetThenInsertSameKeyExactlyOneWins in
// adapters/dalgo2memory/optimistic_test.go (dal-go/dalgo) demonstrates for a
// plain Get-then-Insert: with neither side committed yet, both Inserts
// succeed locally, and the race is decided only at commit, where exactly one
// wins and the other fails with ErrTransactionConflict. A plain channel
// suffices for the determinism, with no dedicated test hook needed, because
// a paused transaction callback holds no lock while it waits — see that
// test's doc comment for the full argument.
func TestClaim_ConcurrentClaimsYieldOneWinner(t *testing.T) {
	db := dalgo2memory.NewDB(dalgo2memory.WithOptimisticConcurrency())
	ctx := context.Background()

	const attempts = 2
	targetIDs := []string{"space-1", "space-2"}
	bothClaimed := make(chan struct{}, attempts)
	proceed := make(chan struct{})
	type outcome struct {
		targetID string
		err      error
	}
	results := make(chan outcome, attempts)

	for i := 0; i < attempts; i++ {
		targetID := targetIDs[i]
		go func() {
			err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
				if _, err := slugs.Claim(ctx, tx, "test:space", "contested-hall", targetID, "space"); err != nil {
					return err
				}
				// Both sides' Insert has now succeeded against this
				// transaction's own local view (neither has committed), so the
				// commit-time optimistic check is what has to decide the
				// winner — not the order these goroutines happened to run in.
				bothClaimed <- struct{}{}
				<-proceed
				return nil
			})
			results <- outcome{targetID: targetID, err: err}
		}()
	}

	// Wait for both transactions to have staged their Insert locally before
	// letting either return from its callback and attempt to commit.
	<-bothClaimed
	<-bothClaimed
	close(proceed)

	first, second := <-results, <-results

	var winner string
	successes, conflicts := 0, 0
	for _, o := range []outcome{first, second} {
		switch {
		case o.err == nil:
			successes++
			winner = o.targetID
		case slugs.IsSlugTaken(o.err) || dalgo2memory.IsTransactionConflict(o.err):
			conflicts++
		default:
			t.Errorf("unexpected error for %s: %v", o.targetID, o.err)
		}
	}
	if successes != 1 {
		t.Errorf("got %d successful claim(s), want exactly 1", successes)
	}
	if conflicts != attempts-1 {
		t.Errorf("got %d taken/conflict failure(s), want %d", conflicts, attempts-1)
	}

	// Assert the actual end state, not merely that an error was returned: the
	// slug must resolve to whichever target actually won, and to nothing
	// else.
	info, err := slugs.Resolve(ctx, db, "test:space", "contested-hall")
	if err != nil {
		t.Fatalf("resolving the contested slug: %v", err)
	}
	if info.TargetID != winner {
		t.Errorf("contested slug resolved to %q, want the actual winner %q", info.TargetID, winner)
	}
}
