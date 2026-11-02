package scheduledwork

import (
	"context"
	"crypto/rand"
	"errors"
	"strings"
	"time"
	"unicode"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
)

var ErrLeaseLost = errors.New("scheduled work claim is no longer current")

// ClaimRef is internal routing metadata, not a user authorization capability.
type ClaimRef struct {
	Ref        Ref
	Generation int64
	Token      string
}

// PrepareClaim runs in the dispatcher's read phase. The caller supplies a fresh
// server instant from INSIDE its transaction callback, never a client clock.
// Every lease takeover creates a new opaque token and retains effect identity.
func PrepareClaim(ctx context.Context, tx dal.ReadwriteTransaction, ref Ref, generation int64, now time.Time, lease time.Duration) (*Change, ClaimRef, error) {
	var empty ClaimRef
	if tx == nil || generation < 1 || generation > MaxGeneration || !instant(now) || lease <= 0 || lease > time.Hour || lease%time.Microsecond != 0 {
		return nil, empty, ErrInvalid
	}
	rec, w, err := readWork(ctx, tx, ref)
	if err != nil {
		return nil, empty, err
	}
	if w.Generation != generation {
		return nil, empty, ErrConflict
	}
	if w.Attempts >= MaxGeneration {
		return nil, empty, ErrInvalid
	}
	if now.Before(w.NextAttemptAt) || (w.State != Pending && w.State != Leased) || (w.State == Leased && now.Before(w.Claim.Until)) {
		return nil, empty, ErrLeaseLost
	}
	token := rand.Text()
	w.State = Leased
	w.Attempts++
	w.Claim = &Claim{Token: token, At: now.UTC(), Until: now.Add(lease).UTC()}
	if err := validatePreparedWork(w); err != nil {
		return nil, empty, err
	}
	return &Change{tx: tx, rec: rec, generation: generation}, ClaimRef{Ref: ref, Generation: generation, Token: token}, nil
}

// Execution captures the SAME transaction that read the current claim. The
// owning handler must read its own business state, perform its writes, then
// return CommitIntent/Complete's error from that transaction. This binds stale
// claim rejection to rollback of the business mutation as well.
type Execution struct {
	tx    dal.ReadwriteTransaction
	rec   record.Record
	work  *Work
	claim ClaimRef
	now   func() time.Time
}

func PrepareExecution(ctx context.Context, tx dal.ReadwriteTransaction, claim ClaimRef, clock func() time.Time) (*Execution, error) {
	if tx == nil || clock == nil || claim.Generation < 1 || claim.Generation > MaxGeneration || !identifier(claim.Token) {
		return nil, ErrInvalid
	}
	rec, w, err := readWork(ctx, tx, claim.Ref)
	if err != nil {
		return nil, err
	}
	e := &Execution{tx: tx, rec: rec, work: w, claim: claim, now: clock}
	if _, err := e.check(); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *Execution) check() (time.Time, error) {
	if e == nil || e.tx == nil || e.work == nil || e.now == nil {
		return time.Time{}, ErrInvalid
	}
	now := e.now().UTC().Truncate(time.Microsecond)
	w := e.work
	if w.Generation != e.claim.Generation || w.State != Leased || w.Claim == nil || w.Claim.Token != e.claim.Token || now.Before(w.Claim.At) || !now.Before(w.Claim.Until) {
		return time.Time{}, ErrLeaseLost
	}
	return now, nil
}

// CommitIntent is called in the transaction that writes the producer's durable
// external-effect intent. Actual provider IO follows successful commit. An
// uncertain attempt can repeat only the identical owning idempotency identity.
func (e *Execution) CommitIntent(ctx context.Context, effectKey string) error {
	if _, err := e.check(); err != nil {
		return err
	}
	if effectKey == "" || len(effectKey) > 256 || strings.TrimSpace(effectKey) != effectKey || strings.IndexFunc(effectKey, unicode.IsControl) >= 0 {
		return ErrInvalid
	}
	if e.work.EffectKey != "" && e.work.EffectKey != effectKey {
		return ErrEffectCommitted
	}
	e.work.EffectKey = effectKey
	if err := validatePreparedWork(e.work); err != nil {
		return err
	}
	return e.tx.Set(ctx, e.rec)
}

// Complete belongs in the successful internal business transaction, or in the
// producer's transaction persisting a real external acknowledgment. Enqueue
// acknowledgment alone does not authorize successful business completion.
func (e *Execution) Complete(ctx context.Context) error {
	now, err := e.check()
	if err != nil {
		return err
	}
	e.work.State = Completed
	e.work.CompletedAt = &now
	e.work.Claim = nil
	if err := validatePreparedWork(e.work); err != nil {
		return err
	}
	return e.tx.Set(ctx, e.rec)
}
