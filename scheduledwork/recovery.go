package scheduledwork

import (
	"context"
	"time"

	"github.com/dal-go/dalgo/dal"
)

// Retry records a producer-classified failure and a strictly later eligible
// instant. It does not choose a business retry policy or change effect identity.
func (e *Execution) Retry(ctx context.Context, at time.Time, code string) error {
	now, err := e.check()
	if err != nil {
		return err
	}
	if !instant(at) || !at.After(now) || !identifier(code) {
		return ErrInvalid
	}
	e.work.State = Pending
	e.work.NextAttemptAt = at.UTC()
	e.work.Claim = nil
	e.work.FailureCode = code
	if err := validatePreparedWork(e.work); err != nil {
		return err
	}
	return e.tx.Set(ctx, e.rec)
}

// Quarantine retains unresolved work and its intent for producer recovery. It
// never records successful completion or authorizes dropping an obligation.
func (e *Execution) Quarantine(ctx context.Context, code string) error {
	if _, err := e.check(); err != nil {
		return err
	}
	if !identifier(code) {
		return ErrInvalid
	}
	e.work.State = Quarantined
	e.work.Claim = nil
	e.work.FailureCode = code
	if err := validatePreparedWork(e.work); err != nil {
		return err
	}
	return e.tx.Set(ctx, e.rec)
}

// PrepareResume is an internal producer-authorized recovery operation, not a
// user endpoint. It preserves generation, subject revision, attempt history and
// effect identity: recovery is another attempt at the same unresolved action.
// attempts binds the command to the observed quarantine episode; a lost-reply
// replay must never reopen a later quarantine of this same generation.
// Old tokens stay invalid; a resumed worker must acquire a fresh lease.
func PrepareResume(ctx context.Context, tx dal.ReadwriteTransaction, ref Ref, generation, attempts int64, at time.Time) (*Change, error) {
	if tx == nil || generation < 1 || generation > MaxGeneration || attempts < 1 || attempts >= MaxGeneration || !instant(at) {
		return nil, ErrInvalid
	}
	rec, w, err := readWork(ctx, tx, ref)
	if err != nil {
		return nil, err
	}
	if w.Generation != generation || w.Attempts != attempts {
		return nil, ErrConflict
	}
	if w.State != Quarantined {
		if w.NextAttemptAt.Equal(at) && (w.State == Pending || w.State == Leased || w.State == Completed) {
			return &Change{tx: tx, generation: generation, noop: true}, nil
		}
		return nil, ErrConflict
	}
	if at.Before(w.Spec.RunAt) || w.Attempts >= MaxGeneration {
		return nil, ErrInvalid
	}
	w.State = Pending
	w.NextAttemptAt = at.UTC()
	if err := validatePreparedWork(w); err != nil {
		return nil, err
	}
	return &Change{tx: tx, rec: rec, generation: generation}, nil
}
