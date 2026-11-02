package scheduledwork

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
	"github.com/sneat-co/sneat-go-core/sneatcoretesting"
)

func retryFixture(t *testing.T) (context.Context, dal.DB, Spec, ClaimRef) {
	t.Helper()
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
		return e.CommitIntent(ctx, "stable-provider-intent")
	}); err != nil {
		t.Fatal(err)
	}
	return ctx, db, spec, claim
}

func TestRetryBackoffAndQuarantinePreserveUncertainIntent(t *testing.T) {
	ctx, db, spec, claim := retryFixture(t)
	at := spec.RunAt.Add(10 * time.Minute)
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		e, err := PrepareExecution(ctx, tx, claim, func() time.Time { return spec.RunAt })
		if err != nil {
			return err
		}
		return e.Retry(ctx, at, "provider-unconfirmed")
	}); err != nil {
		t.Fatal(err)
	}
	w, err := Read(ctx, db, spec.Ref)
	if err != nil || w.State != Pending || w.Claim != nil || w.EffectKey != "stable-provider-intent" || w.Attempts != 1 || w.CompletedAt != nil || !w.NextAttemptAt.Equal(at) {
		t.Fatal("retry lost obligation", w, err)
	}
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, _, err := PrepareClaim(ctx, tx, spec.Ref, 1, spec.RunAt.Add(time.Minute), time.Minute)
		return err
	}); !errors.Is(err, ErrLeaseLost) {
		t.Fatal("backoff bypassed", err)
	}
	second := claimWork(t, db, spec.Ref, 1, at)
	if second.Token == claim.Token {
		t.Fatal("retry reused old token")
	}
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		e, err := PrepareExecution(ctx, tx, second, func() time.Time { return at })
		if err != nil {
			return err
		}
		return e.Quarantine(ctx, "manual-reconciliation")
	}); err != nil {
		t.Fatal(err)
	}
	w, _ = Read(ctx, db, spec.Ref)
	if w.State != Quarantined || w.EffectKey != "stable-provider-intent" || w.Attempts != 2 || w.FailureCode != "manual-reconciliation" || w.CompletedAt != nil {
		t.Fatal("quarantine erased obligation", w)
	}
	resumeAt := at.Add(time.Hour)
	for i := 0; i < 2; i++ {
		if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			change, err := PrepareResume(ctx, tx, spec.Ref, 1, 2, resumeAt)
			if err != nil {
				return err
			}
			return change.Apply(ctx)
		}); err != nil {
			t.Fatal(err)
		}
	}
	w, _ = Read(ctx, db, spec.Ref)
	if w.State != Pending || w.Generation != 1 || w.Attempts != 2 || w.EffectKey != "stable-provider-intent" {
		t.Fatal("resume reset meaning/history", w)
	}
	third := claimWork(t, db, spec.Ref, 1, resumeAt)
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		e, err := PrepareExecution(ctx, tx, third, func() time.Time { return resumeAt })
		if err != nil {
			return err
		}
		return e.Complete(ctx)
	}); err != nil {
		t.Fatal(err)
	}
}

func TestRecoveryRejectsStaleInvalidAndFailedChanges(t *testing.T) {
	ctx, db, spec, claim := retryFixture(t)
	var empty *Execution
	if !errors.Is(empty.Retry(ctx, spec.RunAt, "retry"), ErrInvalid) || !errors.Is(empty.Quarantine(ctx, "retry"), ErrInvalid) {
		t.Fatal("zero execution accepted")
	}
	for _, invalid := range []string{"", strings.Repeat("x", 129), "raw exception details"} {
		for _, quarantine := range []bool{false, true} {
			err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
				e, err := PrepareExecution(ctx, tx, claim, func() time.Time { return spec.RunAt })
				if err != nil {
					return err
				}
				if quarantine {
					return e.Quarantine(ctx, invalid)
				}
				return e.Retry(ctx, spec.RunAt.Add(time.Minute), invalid)
			})
			if !errors.Is(err, ErrInvalid) {
				t.Fatal("unsafe failure classification accepted", err)
			}
		}
	}
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		e, err := PrepareExecution(ctx, tx, claim, func() time.Time { return spec.RunAt })
		if err != nil {
			return err
		}
		return e.Retry(ctx, spec.RunAt, "retry")
	}); !errors.Is(err, ErrInvalid) {
		t.Fatal("hot-loop retry accepted", err)
	}
	before, _ := Read(ctx, db, spec.Ref)
	want := errors.New("candidate invalid")
	original := validatePreparedWork
	defer func() { validatePreparedWork = original }()
	validatePreparedWork = func(*Work) error { return want }
	for _, quarantine := range []bool{false, true} {
		err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			e, err := PrepareExecution(ctx, tx, claim, func() time.Time { return spec.RunAt })
			if err != nil {
				return err
			}
			if quarantine {
				return e.Quarantine(ctx, "reconcile")
			}
			return e.Retry(ctx, spec.RunAt.Add(time.Minute), "retry")
		})
		if !errors.Is(err, want) {
			t.Fatal("recovery validation ignored", err)
		}
		w, _ := Read(ctx, db, spec.Ref)
		if !reflect.DeepEqual(before, w) {
			t.Fatal("failed recovery changed job")
		}
	}
}

func TestResumeRequiresCurrentRecoverableWork(t *testing.T) {
	ctx, db, spec, claim := retryFixture(t)
	at := spec.RunAt.Add(time.Hour)
	if c, err := PrepareResume(ctx, nil, spec.Ref, 1, 1, at); !errors.Is(err, ErrInvalid) || c != nil {
		t.Fatal("nil transaction accepted")
	}
	want := errors.New("storage unavailable")
	if c, err := PrepareResume(ctx, failedReadTx{failure: want}, spec.Ref, 1, 1, at); !errors.Is(err, want) || c != nil {
		t.Fatal("failed read exposed recovery", c, err)
	}
	// Leased work cannot be reset to pending through recovery.
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := PrepareResume(ctx, tx, spec.Ref, 1, 1, at)
		return err
	}); !errors.Is(err, ErrConflict) {
		t.Fatal("active execution reset", err)
	}
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		e, err := PrepareExecution(ctx, tx, claim, func() time.Time { return spec.RunAt })
		if err != nil {
			return err
		}
		return e.Quarantine(ctx, "reconcile")
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := PrepareResume(ctx, tx, spec.Ref, 2, 1, at)
		return err
	}); !errors.Is(err, ErrConflict) {
		t.Fatal("wrong generation resumed", err)
	}
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := PrepareResume(ctx, tx, spec.Ref, 1, 1, spec.RunAt.Add(-time.Second))
		return err
	}); !errors.Is(err, ErrInvalid) {
		t.Fatal("recovery moved work before execution start", err)
	}
	original := validatePreparedWork
	validatePreparedWork = func(*Work) error { return want }
	err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		c, err := PrepareResume(ctx, tx, spec.Ref, 1, 1, at)
		if c != nil {
			t.Fatal("invalid recovery exposed write")
		}
		return err
	})
	validatePreparedWork = original
	if !errors.Is(err, want) {
		t.Fatal("recovery validation ignored", err)
	}
	w, _ := Read(ctx, db, spec.Ref)
	if w.State != Quarantined || w.EffectKey != "stable-provider-intent" {
		t.Fatal("failed recovery changed obligation", w)
	}
	// Counter exhaustion remains recoverable evidence, never an overflow reset.
	w.Attempts = MaxGeneration
	key, _ := spec.Ref.Key()
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Set(ctx, record.NewRecordWithData(key, w))
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := PrepareResume(ctx, tx, spec.Ref, 1, MaxGeneration, at)
		return err
	}); !errors.Is(err, ErrInvalid) {
		t.Fatal("counter exhaustion erased", err)
	}
}

// An old recovery command must not reopen a later reconciliation failure,
// even when its first transaction committed but the caller lost the reply.
func TestResumeLostReplyCannotReopenLaterQuarantine(t *testing.T) {
	ctx, db, spec, first := retryFixture(t)
	quarantine := func(claim ClaimRef, now time.Time) {
		t.Helper()
		if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			e, err := PrepareExecution(ctx, tx, claim, func() time.Time { return now })
			if err != nil {
				return err
			}
			return e.Quarantine(ctx, "reconciliation-needed")
		}); err != nil {
			t.Fatal(err)
		}
	}
	quarantine(first, spec.RunAt)
	at := spec.RunAt.Add(time.Hour)
	resume := func(attempt int64, at time.Time) error {
		return db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			c, err := PrepareResume(ctx, tx, spec.Ref, 1, attempt, at)
			if err != nil {
				return err
			}
			return c.Apply(ctx)
		})
	}
	if err := resume(1, at); err != nil {
		t.Fatal(err)
	}
	second := claimWork(t, db, spec.Ref, 1, at)
	quarantine(second, at)
	before, _ := Read(ctx, db, spec.Ref)
	if err := resume(1, at); !errors.Is(err, ErrConflict) {
		t.Fatal("old recovery reopened newer quarantine", err)
	}
	after, _ := Read(ctx, db, spec.Ref)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("old recovery changed quarantine", before, after)
	}
	if err := resume(2, at.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	after, _ = Read(ctx, db, spec.Ref)
	if after.Attempts != 2 || after.Generation != 1 || after.EffectKey != before.EffectKey {
		t.Fatal("new recovery erased identity", after)
	}
}
