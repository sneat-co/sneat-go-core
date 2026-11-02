package scheduledwork

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"slices"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
)

var (
	ErrConflict        = errors.New("scheduled work generation conflict")
	ErrEffectCommitted = errors.New("external intent already committed; reconcile before replacing work")
	ErrCompleted       = errors.New("scheduled action already completed")
)

// Keep candidate validation observable in fault tests: future model constraints
// must fail before a producer is handed a writable transaction participant.
var validatePreparedWork = (*Work).Validate

// objectFields prevents case-folded aliases and unknown future metadata from
// changing meaning through the shared JSON-backed typed record converter.
func objectFields(data []byte, nullable []string, allowed ...string) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if json.Unmarshal(data, &fields) != nil || fields == nil {
		return nil, ErrInvalid
	}
	for key, value := range fields {
		if !slices.Contains(allowed, key) {
			return nil, ErrInvalid
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) && !slices.Contains(nullable, key) {
			return nil, ErrInvalid
		}
	}
	return fields, nil
}

func decodeWork(fields map[string]any) (*Work, error) {
	encoded, err := json.Marshal(fields)
	if err != nil {
		return nil, err
	}
	top, err := objectFields(encoded, []string{"claim", "completedAt"}, "spec", "generation", "state", "nextAttemptAt", "attempts", "claim", "effectKey", "completedAt", "failureCode")
	if err != nil {
		return nil, err
	}
	spec, err := objectFields(top["spec"], nil, "ref", "handler", "subjectRevision", "runAt")
	if err != nil {
		return nil, err
	}
	if _, err := objectFields(spec["ref"], nil, "spaceID", "producer", "subjectID", "action"); err != nil {
		return nil, err
	}
	if raw := top["claim"]; len(raw) > 0 && !bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		if _, err := objectFields(raw, nil, "token", "at", "until"); err != nil {
			return nil, err
		}
	}
	w := new(Work)
	if err := record.MapToData(w, fields); err != nil {
		return nil, err
	}
	if err := w.Validate(); err != nil {
		return nil, err
	}
	return w, nil
}

func readWork(ctx context.Context, reader dal.Getter, ref Ref) (record.Record, *Work, error) {
	key, err := ref.Key()
	if err != nil {
		return nil, nil, err
	}
	fields := make(map[string]any)
	raw := record.NewRecordWithData(key, &fields)
	if err := reader.Get(ctx, raw); err != nil {
		return nil, nil, err
	}
	w, err := decodeWork(fields)
	if err != nil {
		return nil, nil, err
	}
	if w.Spec.Ref != ref {
		return nil, nil, ErrInvalid
	}
	return record.NewRecordWithData(key, w).SetError(nil), w, nil
}

// Read validates stored shape and binds it to the requested owning key. It is
// an internal backend read, not a client authorization or public read endpoint.
func Read(ctx context.Context, reader dal.Getter, ref Ref) (*Work, error) {
	_, w, err := readWork(ctx, reader, ref)
	return w, err
}

// Change captures its originating transaction. Apply must run during that same
// transaction's write phase; there is no API accepting a different transaction.
// The caller first finishes ALL business/scheduler reads, then applies writes.
type Change struct {
	tx         dal.ReadwriteTransaction
	rec        record.Record
	generation int64
	insert     bool
	noop       bool
}

func (c *Change) Generation() int64 { return c.generation }
func (c *Change) Apply(ctx context.Context) error {
	if c == nil || c.tx == nil {
		return ErrInvalid
	}
	if c.noop {
		return nil
	}
	if c.insert {
		return c.tx.Insert(ctx, c.rec)
	}
	return c.tx.Set(ctx, c.rec)
}

// PrepareSchedule is a transaction participant, never a transaction starter or
// external enqueue. The producer must authorize its business mutation first.
// Expected generation0 creates; an exact generation+1/spec retry acknowledges
// the existing action without resetting attempts, claim, result or timestamps.
func PrepareSchedule(ctx context.Context, tx dal.ReadwriteTransaction, spec Spec, expected int64) (*Change, error) {
	if tx == nil || expected < 0 || expected >= MaxGeneration {
		return nil, ErrInvalid
	}
	if err := spec.Validate(); err != nil {
		return nil, err
	}
	spec.RunAt = spec.RunAt.UTC()
	_, existing, err := readWork(ctx, tx, spec.Ref)
	insert := record.IsNotFound(err)
	if err != nil && !insert {
		return nil, err
	}
	if insert {
		if expected != 0 {
			return nil, ErrConflict
		}
	} else {
		if existing.Generation == expected+1 && existing.Spec.Ref == spec.Ref && existing.Spec.Handler == spec.Handler && existing.Spec.SubjectRevision == spec.SubjectRevision && existing.Spec.RunAt.Equal(spec.RunAt) && existing.State != Cancelled {
			return &Change{tx: tx, generation: existing.Generation, noop: true}, nil
		}
		if existing.Generation != expected {
			return nil, ErrConflict
		}
		if existing.EffectKey != "" && existing.State != Completed {
			return nil, ErrEffectCommitted
		}
	}
	w := &Work{Spec: spec, Generation: expected + 1, State: Pending, NextAttemptAt: spec.RunAt}
	if err := validatePreparedWork(w); err != nil {
		return nil, err
	}
	key, _ := spec.Ref.Key() // Spec validation already proved the key inputs.
	rec := record.NewRecordWithData(key, w)
	return &Change{tx: tx, rec: rec, generation: w.Generation, insert: insert}, nil
}

// PrepareCancel invalidates pending/leased generations before an effect intent
// commits. Cancellation of an already dispatched/completed effect is explicit.
func PrepareCancel(ctx context.Context, tx dal.ReadwriteTransaction, ref Ref, expected int64) (*Change, error) {
	if tx == nil || expected < 1 || expected >= MaxGeneration {
		return nil, ErrInvalid
	}
	rec, w, err := readWork(ctx, tx, ref)
	if err != nil {
		return nil, err
	}
	if w.Generation == expected+1 && w.State == Cancelled {
		return &Change{tx: tx, generation: w.Generation, noop: true}, nil
	}
	if w.Generation != expected {
		return nil, ErrConflict
	}
	if w.State == Completed {
		return nil, ErrCompleted
	}
	if w.EffectKey != "" {
		return nil, ErrEffectCommitted
	}
	w.Generation++
	w.State = Cancelled
	w.Claim = nil
	w.CompletedAt = nil
	if err := validatePreparedWork(w); err != nil {
		return nil, err
	}
	return &Change{tx: tx, rec: rec, generation: w.Generation}, nil
}
