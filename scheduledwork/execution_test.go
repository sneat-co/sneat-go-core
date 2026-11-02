package scheduledwork

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
	"github.com/sneat-co/sneat-go-core/coretypes"
	"github.com/sneat-co/sneat-go-core/sneatcoretesting"
)

func claimWork(t *testing.T, db dal.DB, ref Ref, generation int64, at time.Time) ClaimRef {
	t.Helper()
	var claim ClaimRef
	if err := db.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		change, value, err := PrepareClaim(ctx, tx, ref, generation, at, time.Minute)
		if err != nil {
			return err
		}
		claim = value
		return change.Apply(ctx)
	}); err != nil {
		t.Fatal(err)
	}
	return claim
}

func TestClaimIntentAndAcknowledgedCompletion(t *testing.T) {
	ctx := context.Background()
	db := sneatcoretesting.NewInMemoryTestDB()
	spec := sample().Spec
	schedule(t, db, spec, 0)
	claim := claimWork(t, db, spec.Ref, 1, spec.RunAt)
	now := func() time.Time { return spec.RunAt.Add(time.Second) }
	for _, intent := range []string{"owning-effect-1", "owning-effect-1"} {
		if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			e, err := PrepareExecution(ctx, tx, claim, now)
			if err != nil {
				return err
			}
			return e.CommitIntent(ctx, intent)
		}); err != nil {
			t.Fatal(err)
		}
	}
	err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		e, err := PrepareExecution(ctx, tx, claim, now)
		if err != nil {
			return err
		}
		return e.CommitIntent(ctx, "different-effect")
	})
	if !errors.Is(err, ErrEffectCommitted) {
		t.Fatal("changed effect identity accepted", err)
	}
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		e, err := PrepareExecution(ctx, tx, claim, now)
		if err != nil {
			return err
		}
		return e.Complete(ctx)
	}); err != nil {
		t.Fatal(err)
	}
	w, err := Read(ctx, db, spec.Ref)
	if err != nil || w.State != Completed || w.EffectKey != "owning-effect-1" || w.Claim != nil || !w.CompletedAt.Equal(now()) {
		t.Fatal("incorrect completion", w, err)
	}
	err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := PrepareExecution(ctx, tx, claim, now)
		return err
	})
	if !errors.Is(err, ErrLeaseLost) {
		t.Fatal("completed action remained executable", err)
	}
}

func TestExpiredLeaseTakeoverFencesOldWorkerAndKeepsEffectIdentity(t *testing.T) {
	ctx := context.Background()
	db := sneatcoretesting.NewInMemoryTestDB()
	spec := sample().Spec
	schedule(t, db, spec, 0)
	first := claimWork(t, db, spec.Ref, 1, spec.RunAt)
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		e, err := PrepareExecution(ctx, tx, first, func() time.Time { return spec.RunAt })
		if err != nil {
			return err
		}
		return e.CommitIntent(ctx, "stable-effect")
	}); err != nil {
		t.Fatal(err)
	}
	err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, _, err := PrepareClaim(ctx, tx, spec.Ref, 1, spec.RunAt.Add(time.Second), time.Minute)
		return err
	})
	if !errors.Is(err, ErrLeaseLost) {
		t.Fatal("unexpired lease stolen", err)
	}
	second := claimWork(t, db, spec.Ref, 1, spec.RunAt.Add(time.Minute))
	if first.Token == second.Token {
		t.Fatal("takeover reused stale token")
	}
	w, _ := Read(ctx, db, spec.Ref)
	if w.EffectKey != "stable-effect" || w.Attempts != 2 {
		t.Fatal("takeover lost intent or attempt history", w)
	}
	err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := PrepareExecution(ctx, tx, first, func() time.Time { return spec.RunAt.Add(time.Minute) })
		return err
	})
	if !errors.Is(err, ErrLeaseLost) {
		t.Fatal("old token remained executable", err)
	}
}

func TestLeaseExpiryDuringBusinessWritesRollsBackBoth(t *testing.T) {
	ctx := context.Background()
	db := sneatcoretesting.NewInMemoryTestDB()
	spec := sample().Spec
	schedule(t, db, spec, 0)
	claim := claimWork(t, db, spec.Ref, 1, spec.RunAt)
	var clock atomic.Int64
	clock.Store(spec.RunAt.UnixMicro())
	now := func() time.Time { return time.UnixMicro(clock.Load()).UTC() }
	key := record.NewKeyWithParentAndID(coretypes.NewSpaceModuleKey(spec.Ref.SpaceID, spec.Ref.Producer), "effects", "business-effect")
	err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		e, err := PrepareExecution(ctx, tx, claim, now)
		if err != nil {
			return err
		}
		if err := tx.Set(ctx, record.NewRecordWithData(key, &map[string]any{"done": true})); err != nil {
			return err
		}
		clock.Store(spec.RunAt.Add(time.Minute).UnixMicro())
		return e.Complete(ctx)
	})
	if !errors.Is(err, ErrLeaseLost) {
		t.Fatal("expired claim committed a business result", err)
	}
	if err := db.Get(ctx, record.NewRecordWithData(key, &map[string]any{})); !record.IsNotFound(err) {
		t.Fatal("business effect survived rollback", err)
	}
	w, _ := Read(ctx, db, spec.Ref)
	if w.State != Leased || w.CompletedAt != nil {
		t.Fatal("failed transaction changed job outcome", w)
	}
}

func TestCancellationCommitsBeforePausedWorkerEffect(t *testing.T) {
	ctx := context.Background()
	db := sneatcoretesting.NewInMemoryTestDB()
	spec := sample().Spec
	schedule(t, db, spec, 0)
	claim := claimWork(t, db, spec.Ref, 1, spec.RunAt)
	ready, release := make(chan struct{}), make(chan struct{})
	result := make(chan error, 1)
	key := record.NewKeyWithParentAndID(coretypes.NewSpaceModuleKey(spec.Ref.SpaceID, spec.Ref.Producer), "effects", "stale-effect")
	go func() {
		result <- db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			e, err := PrepareExecution(ctx, tx, claim, func() time.Time { return spec.RunAt })
			if err != nil {
				return err
			}
			close(ready)
			<-release
			if err := tx.Set(ctx, record.NewRecordWithData(key, &map[string]any{"done": true})); err != nil {
				return err
			}
			return e.Complete(ctx)
		})
	}()
	<-ready
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		c, err := PrepareCancel(ctx, tx, spec.Ref, 1)
		if err != nil {
			return err
		}
		return c.Apply(ctx)
	}); err != nil {
		close(release)
		t.Fatal(err)
	}
	close(release)
	if err := <-result; !errors.Is(err, ErrLeaseLost) {
		t.Fatal("old transaction did not lose to cancellation", err)
	}
	if err := db.Get(ctx, record.NewRecordWithData(key, &map[string]any{})); !record.IsNotFound(err) {
		t.Fatal("obsolete effect survived conflict", err)
	}
	w, _ := Read(ctx, db, spec.Ref)
	if w.State != Cancelled || w.Generation != 2 {
		t.Fatal("cancellation was overwritten", w)
	}
}

func TestClaimAndExecutionFailClosed(t *testing.T) {
	ctx := context.Background()
	spec := sample().Spec
	want := errors.New("unavailable")
	tx := failedReadTx{failure: want}
	for _, lease := range []time.Duration{0, -time.Second, 2 * time.Hour, time.Nanosecond} {
		if c, claim, err := PrepareClaim(ctx, tx, spec.Ref, 1, spec.RunAt, lease); !errors.Is(err, ErrInvalid) || c != nil || claim.Token != "" {
			t.Fatal("invalid lease exposed a claim", c, claim, err)
		}
	}
	if c, _, err := PrepareClaim(ctx, tx, spec.Ref, 1, spec.RunAt, time.Minute); !errors.Is(err, want) || c != nil {
		t.Fatal("failed read created a claim", err)
	}
	handle := ClaimRef{Ref: spec.Ref, Generation: 1, Token: "token"}
	if e, err := PrepareExecution(ctx, tx, handle, nil); !errors.Is(err, ErrInvalid) || e != nil {
		t.Fatal("nil clock accepted")
	}
	if e, err := PrepareExecution(ctx, tx, handle, func() time.Time { return spec.RunAt }); !errors.Is(err, want) || e != nil {
		t.Fatal("failed read exposed execution", err)
	}
	var empty *Execution
	if !errors.Is(empty.Complete(ctx), ErrInvalid) || !errors.Is(empty.CommitIntent(ctx, "key"), ErrInvalid) {
		t.Fatal("nil execution was usable")
	}
	db := sneatcoretesting.NewInMemoryTestDB()
	schedule(t, db, spec, 0)
	for _, generation := range []int64{0, 2} {
		err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			_, _, err := PrepareClaim(ctx, tx, spec.Ref, generation, spec.RunAt, time.Minute)
			return err
		})
		if err == nil {
			t.Fatal("bad generation was claimable")
		}
	}
	err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, _, err := PrepareClaim(ctx, tx, spec.Ref, 1, spec.RunAt.Add(-time.Second), time.Minute)
		return err
	})
	if !errors.Is(err, ErrLeaseLost) {
		t.Fatal("future work was claimable", err)
	}
	claim := claimWork(t, db, spec.Ref, 1, spec.RunAt)
	for _, badKey := range []string{"", " invalid ", "control\nkey"} {
		err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			e, err := PrepareExecution(ctx, tx, claim, func() time.Time { return spec.RunAt })
			if err != nil {
				return err
			}
			return e.CommitIntent(ctx, badKey)
		})
		if !errors.Is(err, ErrInvalid) {
			t.Fatal("invalid external identity accepted", err)
		}
	}
	// Exhausted attempt count cannot wrap back to a fresh claim.
	w := sample()
	w.Attempts = MaxGeneration
	fields, _ := record.DataToMap(w)
	if c, _, err := PrepareClaim(ctx, structReadTx{getter: workGetter{fields: fields}}, spec.Ref, 1, spec.RunAt, time.Minute); !errors.Is(err, ErrInvalid) || c != nil {
		t.Fatal("attempt overflow was claimable", err)
	}
}

type structReadTx struct {
	dal.ReadwriteTransaction
	getter workGetter
}

func (tx structReadTx) Get(ctx context.Context, r record.Record) error { return tx.getter.Get(ctx, r) }

func TestCandidateValidationFailureCannotCommitClaimOrExecution(t *testing.T) {
	ctx := context.Background()
	db := sneatcoretesting.NewInMemoryTestDB()
	spec := sample().Spec
	schedule(t, db, spec, 0)
	want := errors.New("candidate rejected")
	original := validatePreparedWork
	defer func() { validatePreparedWork = original }()
	validatePreparedWork = func(*Work) error { return want }
	err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		c, claim, err := PrepareClaim(ctx, tx, spec.Ref, 1, spec.RunAt, time.Minute)
		if c != nil || claim.Token != "" {
			t.Fatal("invalid candidate exposed a claim")
		}
		return err
	})
	if !errors.Is(err, want) {
		t.Fatal("claim ignored validation failure", err)
	}
	validatePreparedWork = original
	claim := claimWork(t, db, spec.Ref, 1, spec.RunAt)
	validatePreparedWork = func(*Work) error { return want }
	for _, complete := range []bool{false, true} {
		err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			e, err := PrepareExecution(ctx, tx, claim, func() time.Time { return spec.RunAt })
			if err != nil {
				return err
			}
			if complete {
				return e.Complete(ctx)
			}
			return e.CommitIntent(ctx, "stable-key")
		})
		if !errors.Is(err, want) {
			t.Fatal("execution ignored validation failure", err)
		}
		w, err := Read(ctx, db, spec.Ref)
		if err != nil || w.State != Leased || w.EffectKey != "" || w.CompletedAt != nil {
			t.Fatal("failed validation changed stored job", w, err)
		}
	}
}

func TestExternalIntentCommitsBeforeCancellation(t *testing.T) {
	ctx := context.Background()
	db := sneatcoretesting.NewInMemoryTestDB()
	spec := sample().Spec
	schedule(t, db, spec, 0)
	claim := claimWork(t, db, spec.Ref, 1, spec.RunAt)
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		e, err := PrepareExecution(ctx, tx, claim, func() time.Time { return spec.RunAt })
		if err != nil {
			return err
		}
		return e.CommitIntent(ctx, "committed-intent")
	}); err != nil {
		t.Fatal(err)
	}
	err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		c, err := PrepareCancel(ctx, tx, spec.Ref, 1)
		if c != nil {
			t.Fatal("committed intent became cancellable")
		}
		return err
	})
	if !errors.Is(err, ErrEffectCommitted) {
		t.Fatal("cancellation falsely claimed recall", err)
	}
}
