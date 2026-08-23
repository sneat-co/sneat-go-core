package slugs_test

import (
	"context"
	"sync"
	"testing"

	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/slugs"
	"github.com/sneat-co/sneat-go-core/sneatcoretesting"
)

// TestClaim_ConcurrentClaimsYieldOneWinner is meant to verify
// AC:concurrent-claims-yield-one-winner: two transactions racing to claim
// the same slug must settle with exactly one commit, and no observer can
// ever see both holding the claim.
//
// It is SKIPPED, exactly as spec/features/slug-claims/README.md and
// spec/decisions/0001-slug-claims.md require: dalgo2memory — the only
// backend available to this repo's tests — takes a global write lock for
// the whole of RunReadwriteTransaction (db.mu.Lock() in
// dalgo2memory's database.RunReadwriteTransaction, held for the entire
// callback), so two "concurrent" transactions against it can never actually
// interleave — the second always runs to completion only after the first
// has fully finished. A contention test against this adapter would pass
// unconditionally regardless of whether Claim's atomicity is correct, which
// is precisely the "passes regardless" the Feature forbids presenting as
// satisfied.
//
// This is provable only against a backend that can genuinely contend on the
// same write — a real emulator/database, or dalgo2memory once it grows a
// contention mode — which is tracked separately in dal-go/dalgo per the
// Decision, in parallel with this package. The test body below is exactly
// the scenario to un-skip once that lands: two goroutines race
// RunReadwriteTransaction against the same (namespace, slug), and exactly
// one Claim must succeed.
func TestClaim_ConcurrentClaimsYieldOneWinner(t *testing.T) {
	t.Skip("dalgo2memory serialises every RunReadwriteTransaction under one global lock, so two transactions can never actually contend against it — proving this AC needs a real backend or a dalgo2memory contention mode; see the comment on this test and spec/decisions/0001-slug-claims.md")

	db := sneatcoretesting.NewMemoryDB()
	ctx := context.Background()

	const attempts = 2
	targetIDs := []string{"space-1", "space-2"}
	results := make(chan error, attempts)
	var wg sync.WaitGroup
	wg.Add(attempts)
	for i := 0; i < attempts; i++ {
		targetID := targetIDs[i]
		go func() {
			defer wg.Done()
			err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
				_, err := slugs.Claim(ctx, tx, "test:space", "contested-hall", targetID, "space")
				return err
			})
			results <- err
		}()
	}
	wg.Wait()
	close(results)

	successes, taken := 0, 0
	for err := range results {
		switch {
		case err == nil:
			successes++
		case slugs.IsSlugTaken(err):
			taken++
		default:
			t.Errorf("unexpected error: %v", err)
		}
	}
	if successes != 1 {
		t.Errorf("got %d successful claim(s), want exactly 1", successes)
	}
	if taken != attempts-1 {
		t.Errorf("got %d slug-taken failure(s), want %d", taken, attempts-1)
	}
}
