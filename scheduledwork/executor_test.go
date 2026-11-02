package scheduledwork

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
	"github.com/sneat-co/sneat-go-core/coretypes"
	"github.com/sneat-co/sneat-go-core/sneatcoretesting"
	"github.com/strongo/delaying"
)

type delayedClaimWorker = func(context.Context, string, string, string, int64, string) error

type executorHarness struct {
	worker       delayedClaimWorker
	queued       []any
	params       delaying.Params
	queueFailure error
}

func (h *executorHarness) register(key string, implementation any) delaying.Delayer {
	h.worker = implementation.(delayedClaimWorker)
	return delaying.NewDelayer(key, implementation, func(_ context.Context, params delaying.Params, args ...any) error {
		h.params = params
		h.queued = append([]any(nil), args...)
		return h.queueFailure
	}, func(context.Context, delaying.Params, ...[]any) error { return errors.New("unexpected batch enqueue") })
}
func (h *executorHarness) deliver(ctx context.Context) error {
	return h.worker(ctx, h.queued[0].(string), h.queued[1].(string), h.queued[2].(string), h.queued[3].(int64), h.queued[4].(string))
}

type executorBusinessCounter struct {
	Applied int64 `json:"applied"`
}

func TestExistingDelayedExecutorDeliversThroughOwningTransactionFence(t *testing.T) {
	for _, cancelled := range []bool{false, true} {
		t.Run(map[bool]string{false: "complete-once", true: "cancel-before-delivery"}[cancelled], func(t *testing.T) {
			ctx := context.Background()
			db := sneatcoretesting.NewInMemoryTestDB()
			spec := sample().Spec
			schedule(t, db, spec, 0)
			claim := claimWork(t, db, spec.Ref, 1, spec.RunAt)
			businessKey := record.NewKeyWithParentAndID(coretypes.NewSpaceModuleKey(spec.Ref.SpaceID, "core"), "schedulerTestEntries", "business")
			run := func(ctx context.Context, claim ClaimRef) error {
				err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
					e, err := PrepareExecution(ctx, tx, claim, func() time.Time { return spec.RunAt.Add(time.Second) })
					if err != nil {
						return err
					}
					// The owning handler authorizes its real business/handler/revision. The
					// transport never substitutes argument validation for this transaction.
					current, err := Read(ctx, tx, claim.Ref)
					if err != nil {
						return err
					}
					if current.Spec.Handler != spec.Handler || current.Spec.SubjectRevision != spec.SubjectRevision {
						return ErrConflict
					}
					count := new(executorBusinessCounter)
					rec := record.NewRecordWithData(businessKey, count)
					if err := tx.Get(ctx, rec); err != nil && !record.IsNotFound(err) {
						return err
					}
					count.Applied++
					if err := tx.Set(ctx, rec); err != nil {
						return err
					}
					return e.Complete(ctx)
				})
				// This no-op is an explicit owning-handler policy for obsolete deliveries.
				if errors.Is(err, ErrLeaseLost) {
					return nil
				}
				return err
			}
			h := new(executorHarness)
			executor, err := RegisterExecutor(h.register, spec.Ref.Producer, spec.Handler, "commitius", run)
			if err != nil {
				t.Fatal(err)
			}
			if err := executor.Enqueue(ctx, claim); err != nil {
				t.Fatal(err)
			}
			if h.params.Queue() != "commitius" || h.params.Path() != spec.Handler || h.params.Delay() != 0 || !reflect.DeepEqual(h.queued, []any{string(spec.Ref.SpaceID), spec.Ref.SubjectID, spec.Ref.Action, claim.Generation, claim.Token}) {
				t.Fatal("wrong shared runtime payload", h.params, h.queued)
			}
			w, _ := Read(ctx, db, spec.Ref)
			if w.State != Leased || w.CompletedAt != nil {
				t.Fatal("enqueue fabricated completion", w)
			}
			if cancelled {
				if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
					c, err := PrepareCancel(ctx, tx, spec.Ref, 1)
					if err != nil {
						return err
					}
					return c.Apply(ctx)
				}); err != nil {
					t.Fatal(err)
				}
			}
			if err := h.deliver(ctx); err != nil {
				t.Fatal(err)
			}
			if err := h.deliver(ctx); err != nil {
				t.Fatal(err)
			}
			count := new(executorBusinessCounter)
			rec := record.NewRecordWithData(businessKey, count)
			err = db.Get(ctx, rec)
			w, _ = Read(ctx, db, spec.Ref)
			if cancelled {
				if !record.IsNotFound(err) || w.State != Cancelled {
					t.Fatal("cancelled delivery wrote business state", count, w, err)
				}
			} else if err != nil || count.Applied != 1 || w.State != Completed {
				t.Fatal("duplicate delivery repeated or lost transition", count, w, err)
			}
		})
	}
}

func TestExecutorPreservesUnconfirmedEnqueueAndOwningErrors(t *testing.T) {
	ctx := context.Background()
	db := sneatcoretesting.NewInMemoryTestDB()
	spec := sample().Spec
	schedule(t, db, spec, 0)
	claim := claimWork(t, db, spec.Ref, 1, spec.RunAt)
	want := errors.New("delivery not acknowledged")
	h := &executorHarness{queueFailure: want}
	executor, err := RegisterExecutor(h.register, spec.Ref.Producer, spec.Handler, "commitius", func(context.Context, ClaimRef) error { return want })
	if err != nil {
		t.Fatal(err)
	}
	if err := executor.Enqueue(ctx, claim); !errors.Is(err, want) {
		t.Fatal("enqueue uncertainty hidden", err)
	}
	w, _ := Read(ctx, db, spec.Ref)
	if w.State != Leased || w.Attempts != 1 || w.CompletedAt != nil || w.Claim.Token != claim.Token {
		t.Fatal("enqueue failure destroyed recoverable lease", w)
	}
	if err := h.deliver(ctx); !errors.Is(err, want) {
		t.Fatal("owning worker error hidden from existing host", err)
	}
	d := Discovery{Producer: spec.Ref.Producer, Handler: spec.Handler, State: Leased, Now: w.Claim.Until, Limit: 10}
	b, err := Discover(ctx, db, d)
	if err != nil || len(b.Ready) != 1 || b.Ready[0].Ref != claim.Ref {
		t.Fatal("unconfirmed queue work lost recovery entry", b, err)
	}
}

func TestExecutorRejectsInvalidConfigurationAndPayloads(t *testing.T) {
	runCalls := 0
	run := func(context.Context, ClaimRef) error { runCalls++; return nil }
	registerCalls := 0
	register := func(string, any) delaying.Delayer { registerCalls++; return nil }
	for _, cfg := range []struct {
		producer       coretypes.ExtID
		handler, queue string
	}{{"", "handler", "queue"}, {"commitius", "bad/handler", "queue"}, {"commitius", "handler", ""}} {
		if e, err := RegisterExecutor(register, cfg.producer, cfg.handler, cfg.queue, run); !errors.Is(err, ErrInvalid) || e != nil {
			t.Fatal("invalid registration accepted", err)
		}
	}
	if registerCalls != 0 {
		t.Fatal("invalid config registered a worker")
	}
	if _, err := RegisterExecutor(nil, "commitius", "handler", "queue", run); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := RegisterExecutor(register, "commitius", "handler", "queue", nil); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	if e, err := RegisterExecutor(register, "commitius", "handler", "queue", run); !errors.Is(err, ErrInvalid) || e != nil {
		t.Fatal("nil registered delayer accepted", err)
	}
	h := new(executorHarness)
	e, err := RegisterExecutor(h.register, "commitius", "handler", "queue", run)
	if err != nil {
		t.Fatal(err)
	}
	claim := ClaimRef{Ref: sample().Spec.Ref, Generation: 1, Token: "claim-token"}
	for _, change := range []func(*ClaimRef){func(c *ClaimRef) { c.Ref.Producer = "debtus" }, func(c *ClaimRef) { c.Ref.SpaceID = "" }, func(c *ClaimRef) { c.Generation = 0 }, func(c *ClaimRef) { c.Generation = MaxGeneration + 1 }, func(c *ClaimRef) { c.Token = "" }} {
		bad := claim
		change(&bad)
		if err := e.Enqueue(context.Background(), bad); !errors.Is(err, ErrInvalid) {
			t.Fatal("unsafe claim enqueued", err)
		}
	}
	if h.queued != nil {
		t.Fatal("invalid claim touched executor")
	}
	if err := h.worker(context.Background(), "", "subject", "deadline", 1, "token"); !errors.Is(err, ErrInvalid) || runCalls != 0 {
		t.Fatal("invalid delivered reference reached business code", err)
	}
	if err := h.worker(context.Background(), "personal", "subject", "deadline", 0, "token"); !errors.Is(err, ErrInvalid) || runCalls != 0 {
		t.Fatal("invalid generation reached business code", err)
	}
	var empty *Executor
	if err := empty.Enqueue(context.Background(), claim); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	e.delayer = nil
	if err := e.Enqueue(context.Background(), claim); !errors.Is(err, ErrInvalid) {
		t.Fatal("zero delayer accepted", err)
	}
}
