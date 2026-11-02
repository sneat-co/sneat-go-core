package scheduledwork

import (
	"context"
	"errors"
	"maps"
	"reflect"
	"testing"
	"time"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
	"github.com/sneat-co/sneat-go-core/coretypes"
	"github.com/sneat-co/sneat-go-core/sneatcoretesting"
)

func schedule(t *testing.T, db dal.DB, spec Spec, expected int64) int64 {
	t.Helper()
	var generation int64
	if err := db.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		change, err := PrepareSchedule(ctx, tx, spec, expected)
		if err != nil {
			return err
		}
		generation = change.Generation()
		return change.Apply(ctx)
	}); err != nil {
		t.Fatal(err)
	}
	return generation
}

func TestScheduleReplacementCancellationAndExactReplay(t *testing.T) {
	ctx := context.Background()
	db := sneatcoretesting.NewInMemoryTestDB()
	spec := sample().Spec
	if schedule(t, db, spec, 0) != 1 {
		t.Fatal("wrong first generation")
	}
	// Equal instants in a different offset remain the identical request.
	local := spec
	local.RunAt = local.RunAt.In(time.FixedZone("offset", 3600))
	if schedule(t, db, local, 0) != 1 {
		t.Fatal("timezone representation changed replay")
	}
	spec.SubjectRevision = "2"
	spec.RunAt = spec.RunAt.Add(time.Hour)
	if schedule(t, db, spec, 1) != 2 {
		t.Fatal("replacement did not invalidate old generation")
	}
	err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := PrepareSchedule(ctx, tx, spec, 0)
		return err
	})
	if !errors.Is(err, ErrConflict) {
		t.Fatal("stale schedule accepted", err)
	}
	for i := 0; i < 2; i++ {
		err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			change, err := PrepareCancel(ctx, tx, spec.Ref, 2)
			if err != nil {
				return err
			}
			if change.Generation() != 3 {
				t.Fatal("wrong cancellation generation")
			}
			return change.Apply(ctx)
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	w, err := Read(ctx, db, spec.Ref)
	if err != nil || w.State != Cancelled || w.Generation != 3 {
		t.Fatal("cancel not durable", w, err)
	}
	err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := PrepareSchedule(ctx, tx, spec, 1)
		return err
	})
	if !errors.Is(err, ErrConflict) {
		t.Fatal("old schedule resurrected cancelled work", err)
	}
}

type failScheduleTx struct{ dal.ReadwriteTransaction }

func (tx failScheduleTx) Insert(ctx context.Context, r record.Record, opts ...dal.InsertOption) error {
	if r.Key().Collection() == Collection {
		return errors.New("scheduled write failed")
	}
	return tx.ReadwriteTransaction.Insert(ctx, r, opts...)
}

func TestBusinessAndScheduleRollbackTogether(t *testing.T) {
	ctx := context.Background()
	db := sneatcoretesting.NewInMemoryTestDB()
	spec := sample().Spec
	businessKey := record.NewKeyWithParentAndID(coretypes.NewSpaceModuleKey(spec.Ref.SpaceID, spec.Ref.Producer), "subjects", spec.Ref.SubjectID)
	err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, base dal.ReadwriteTransaction) error {
		tx := failScheduleTx{base}
		change, err := PrepareSchedule(ctx, tx, spec, 0)
		if err != nil {
			return err
		}
		if err := tx.Insert(ctx, record.NewRecordWithData(businessKey, &map[string]any{"title": "Promise"})); err != nil {
			return err
		}
		return change.Apply(ctx)
	})
	if err == nil {
		t.Fatal("missing injected schedule failure")
	}
	if err := db.Get(ctx, record.NewRecordWithData(businessKey, &map[string]any{})); !record.IsNotFound(err) {
		t.Fatal("business mutation escaped rollback", err)
	}
	if _, err := Read(ctx, db, spec.Ref); !record.IsNotFound(err) {
		t.Fatal("schedule escaped rollback", err)
	}
}

func TestCannotReplaceUncertainEffectOrCancelCompletedAction(t *testing.T) {
	for _, state := range []State{Leased, Completed} {
		ctx := context.Background()
		db := sneatcoretesting.NewInMemoryTestDB()
		w := sample()
		w.State = state
		w.Attempts = 1
		w.EffectKey = "stable-effect"
		if state == Leased {
			w.Claim = &Claim{Token: "token", At: w.Spec.RunAt, Until: w.Spec.RunAt.Add(time.Minute)}
		} else {
			w.CompletedAt = &w.Spec.RunAt
		}
		key, _ := w.Spec.Ref.Key()
		if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			return tx.Insert(ctx, record.NewRecordWithData(key, &w))
		}); err != nil {
			t.Fatal(err)
		}
		// Exact create replay acknowledges original execution without erasing it.
		if schedule(t, db, w.Spec, 0) != 1 {
			t.Fatal("replay changed uncertain execution")
		}
		err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			_, err := PrepareCancel(ctx, tx, w.Spec.Ref, 1)
			return err
		})
		want := ErrEffectCommitted
		if state == Completed {
			want = ErrCompleted
		}
		if !errors.Is(err, want) {
			t.Fatal(state, err)
		}
		if state == Leased {
			changed := w.Spec
			changed.SubjectRevision = "2"
			err = db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
				_, err := PrepareSchedule(ctx, tx, changed, 1)
				return err
			})
			if !errors.Is(err, ErrEffectCommitted) {
				t.Fatal("replacement erased external uncertainty", err)
			}
		}
	}
}

type workGetter struct {
	dal.Getter
	fields map[string]any
	err    error
}

func (g workGetter) Get(_ context.Context, r record.Record) error {
	if g.err != nil {
		return g.err
	}
	*r.Data().(*map[string]any) = maps.Clone(g.fields)
	r.SetError(nil)
	return nil
}

func TestStoredWorkRejectsAliasesAndWrongOwningIdentity(t *testing.T) {
	w := sample()
	fields, err := record.DataToMap(w)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Read(context.Background(), workGetter{fields: fields}, w.Spec.Ref); err != nil {
		t.Fatal("valid original fields rejected", err)
	}
	for _, change := range []func(map[string]any){
		func(f map[string]any) { f["Generation"] = int64(99) },
		func(f map[string]any) { f["future"] = true },
		func(f map[string]any) { f["spec"] = nil },
		func(f map[string]any) { f["spec"] = false },
		func(f map[string]any) { f["spec"] = map[string]any{"ref": map[string]any{"SpaceID": "alias"}} },
		func(f map[string]any) { f["claim"] = map[string]any{"Token": "alias"} },
		func(f map[string]any) { f["generation"] = "not an integer" },
		func(f map[string]any) { f["state"] = "future" },
		func(f map[string]any) { f["bad"] = make(chan int) },
	} {
		f := maps.Clone(fields)
		change(f)
		if value, err := Read(context.Background(), workGetter{fields: f}, w.Spec.Ref); err == nil || value != nil {
			t.Fatal("invalid stored work became actionable", value, err)
		}
	}
	other := w.Spec.Ref
	other.SpaceID = "another"
	if value, err := Read(context.Background(), workGetter{fields: fields}, other); !errors.Is(err, ErrInvalid) || value != nil {
		t.Fatal("cross-Space data accepted", value, err)
	}
	if value, err := Read(context.Background(), workGetter{fields: fields}, Ref{}); !errors.Is(err, ErrInvalid) || value != nil {
		t.Fatal("global fallback accepted", value, err)
	}
}

type failedReadTx struct {
	dal.ReadwriteTransaction
	failure error
}

func (tx failedReadTx) Get(context.Context, record.Record) error { return tx.failure }

func TestPreparationFailuresNeverReturnWritableChange(t *testing.T) {
	ctx := context.Background()
	spec := sample().Spec
	var empty *Change
	if !errors.Is(empty.Apply(ctx), ErrInvalid) || !errors.Is((&Change{}).Apply(ctx), ErrInvalid) {
		t.Fatal("zero transaction participant was usable")
	}
	for _, expected := range []int64{-1, 0, MaxGeneration} {
		if c, err := PrepareSchedule(ctx, nil, spec, expected); !errors.Is(err, ErrInvalid) || c != nil {
			t.Fatal("nil transaction accepted")
		}
		if c, err := PrepareCancel(ctx, nil, spec.Ref, expected); !errors.Is(err, ErrInvalid) || c != nil {
			t.Fatal("nil cancellation transaction accepted")
		}
	}
	want := errors.New("storage unavailable")
	tx := failedReadTx{failure: want}
	if c, err := PrepareSchedule(ctx, tx, spec, 0); !errors.Is(err, want) || c != nil {
		t.Fatal("read failure yielded a schedule", c, err)
	}
	if c, err := PrepareCancel(ctx, tx, spec.Ref, 1); !errors.Is(err, want) || c != nil {
		t.Fatal("read failure yielded cancellation", c, err)
	}
	bad := spec
	bad.Ref.SpaceID = ""
	if c, err := PrepareSchedule(ctx, tx, bad, 0); !errors.Is(err, ErrInvalid) || c != nil {
		t.Fatal("invalid spec accepted")
	}
	db := sneatcoretesting.NewInMemoryTestDB()
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := PrepareSchedule(ctx, tx, spec, 1)
		return err
	}); !errors.Is(err, ErrConflict) {
		t.Fatal("missing prior generation accepted", err)
	}
	schedule(t, db, spec, 0)
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, err := PrepareCancel(ctx, tx, spec.Ref, 2)
		return err
	}); !errors.Is(err, ErrConflict) {
		t.Fatal("stale cancellation accepted", err)
	}
	original := validatePreparedWork
	defer func() { validatePreparedWork = original }()
	validatePreparedWork = func(*Work) error { return want }
	for _, cancel := range []bool{false, true} {
		err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			var c *Change
			var err error
			if cancel {
				c, err = PrepareCancel(ctx, tx, spec.Ref, 1)
			} else {
				c, err = PrepareSchedule(ctx, tx, spec, 1)
			}
			if c != nil {
				t.Fatal("validation failure exposed writable state")
			}
			return err
		})
		if !errors.Is(err, want) {
			t.Fatal("candidate validation error was ignored", err)
		}
		w, err := Read(ctx, db, spec.Ref)
		if err != nil || w.Generation != 1 || w.State != Pending {
			t.Fatal("validation failure changed stored work", w, err)
		}
	}
}

func TestNullScalarMetadataCannotEraseExternalIntent(t *testing.T) {
	for _, field := range []string{"effectKey", "attempts", "failureCode"} {
		t.Run(field, func(t *testing.T) {
			ctx := context.Background()
			db := sneatcoretesting.NewInMemoryTestDB()
			w := sample()
			w.State = Leased
			w.Attempts = 1
			w.EffectKey = "uncertain-effect"
			w.Claim = &Claim{Token: "claim", At: w.Spec.RunAt, Until: w.Spec.RunAt.Add(time.Minute)}
			fields, err := record.DataToMap(w)
			if err != nil {
				t.Fatal(err)
			}
			fields[field] = nil
			key, _ := w.Spec.Ref.Key()
			if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
				return tx.Insert(ctx, record.NewRecordWithData(key, &fields))
			}); err != nil {
				t.Fatal(err)
			}
			if value, err := Read(ctx, db, w.Spec.Ref); !errors.Is(err, ErrInvalid) || value != nil {
				t.Fatal("null scalar became usable", value, err)
			}
			for _, cancel := range []bool{false, true} {
				err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
					var c *Change
					var err error
					if cancel {
						c, err = PrepareCancel(ctx, tx, w.Spec.Ref, 1)
					} else {
						next := w.Spec
						next.SubjectRevision = "2"
						c, err = PrepareSchedule(ctx, tx, next, 1)
					}
					if c != nil {
						t.Fatal("null scalar exposed a writable change")
					}
					return err
				})
				if !errors.Is(err, ErrInvalid) {
					t.Fatal("null scalar bypassed intent guard", err)
				}
			}
			stored := make(map[string]any)
			if err := db.Get(ctx, record.NewRecordWithData(key, &stored)); err != nil || !reflect.DeepEqual(stored, fields) {
				t.Fatal("rejected null metadata changed storage", stored, err)
			}
		})
	}
	// Pointer nulls are deliberately equivalent to absence in pending work.
	w := sample()
	fields, _ := record.DataToMap(w)
	fields["claim"], fields["completedAt"] = nil, nil
	if value, err := Read(context.Background(), workGetter{fields: fields}, w.Spec.Ref); err != nil || value == nil {
		t.Fatal("nullable pointer fields rejected", value, err)
	}
}
