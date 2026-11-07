package scheduledwork

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
)

func TestCycleTakeoverFencesAdvanceReleaseAndClaims(t *testing.T) {
	f := newDispatchFixture(t)
	ctx := context.Background()
	w := sample()
	schedule(t, f.db, w.Spec, 0)
	first, err := f.d.acquire(ctx, Pending)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.d.acquire(ctx, Pending); !errors.Is(err, errCycleBusy) {
		t.Fatal("live lease stolen", err)
	}
	if err = f.d.checkpoint(ctx, first, "after-rejected", false); err != nil {
		t.Fatal(err)
	}
	if err = f.d.checkpoint(ctx, first, "", true); err != nil {
		t.Fatal(err)
	}
	f.now = f.now.Add(time.Second)
	second, err := f.d.acquire(ctx, Pending)
	if err != nil {
		t.Fatal(err)
	}
	if second.Lease.Token == first.Lease.Token || second.Epoch != first.Epoch || second.Cursor != "after-rejected" || !second.Cutoff.Equal(first.Cutoff) {
		t.Fatal("released cycle did not resume", second)
	}
	for _, release := range []bool{false, true} {
		if err = f.d.checkpoint(ctx, first, "wrong", release); !errors.Is(err, ErrLeaseLost) {
			t.Fatal("stale writer advanced", err)
		}
	}
	f.now = second.Lease.Until
	third, err := f.d.acquire(ctx, Pending)
	if err != nil {
		t.Fatal(err)
	}
	if third.Epoch != second.Epoch || third.Lease.Token == second.Lease.Token || third.Cursor != second.Cursor {
		t.Fatal("takeover reset cycle")
	}
	if _, err = f.d.claim(ctx, second, Candidate{Ref: w.Spec.Ref, Generation: 1}); !errors.Is(err, ErrLeaseLost) {
		t.Fatal("old scan claimed work", err)
	}
	if err = f.d.checkpoint(ctx, third, "", false); err != nil {
		t.Fatal(err)
	}
	if _, err = f.d.claim(ctx, third, Candidate{Ref: w.Spec.Ref, Generation: 1}); !errors.Is(err, ErrLeaseLost) {
		t.Fatal("EOF scan claimed work", err)
	}
	fourth, err := f.d.acquire(ctx, Pending)
	if err != nil {
		t.Fatal(err)
	}
	if fourth.Epoch != third.Epoch+1 || fourth.Cursor != "" || !fourth.Cutoff.Equal(f.now) {
		t.Fatal("EOF did not begin fresh epoch")
	}
	if err = f.d.checkpoint(ctx, fourth, strings.Repeat("x", (64<<10)+1), false); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
}

func TestCycleConcurrentAcquisitionHasOneOwner(t *testing.T) {
	f := newDispatchFixture(t)
	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); <-start; _, err := f.d.acquire(context.Background(), Pending); results <- err }()
	}
	close(start)
	wg.Wait()
	close(results)
	successes, busy := 0, 0
	for err := range results {
		if err == nil {
			successes++
		} else if errors.Is(err, errCycleBusy) {
			busy++
		} else {
			t.Fatal(err)
		}
	}
	if successes != 1 || busy != 1 {
		t.Fatal(successes, busy)
	}
	report, err := f.d.Tick(context.Background())
	if err != nil || report.Busy != 1 || report.Pages != 1 {
		t.Fatal(report, err)
	}
}

func TestCycleStrictDecodeRejectsCorruptOrMismatchedCheckpoints(t *testing.T) {
	for _, mutate := range []func(map[string]any){
		func(m map[string]any) { m["unexpected"] = "secret" }, func(m map[string]any) { m["Cursor"] = "alias" },
		func(m map[string]any) { delete(m, "cursor") }, func(m map[string]any) { m["cursor"] = nil },
		func(m map[string]any) { m["schemaVersion"] = 2 }, func(m map[string]any) { m["producer"] = "different" },
		func(m map[string]any) { m["handler"] = "different" }, func(m map[string]any) { m["state"] = string(Leased) },
		func(m map[string]any) { m["pageSize"] = 2 }, func(m map[string]any) { m["epoch"] = int64(0) },
		func(m map[string]any) { m["epoch"] = MaxGeneration + 1 }, func(m map[string]any) { m["epoch"] = "string" },
		func(m map[string]any) { m["cursor"] = strings.Repeat("a", (64<<10)+1) },
		func(m map[string]any) { m["cutoff"] = time.Time{} }, func(m map[string]any) { m["cutoff"] = sample().Spec.RunAt.Add(time.Nanosecond) },
		func(m map[string]any) { m["lease"] = map[string]any{"unknown": true} },
		func(m map[string]any) { m["lease"] = map[string]any{"token": "t", "at": nil, "until": nil} },
		func(m map[string]any) {
			m["lease"] = map[string]any{"token": "bad/token", "at": sample().Spec.RunAt, "until": sample().Spec.RunAt.Add(time.Minute)}
		},
	} {
		f := newDispatchFixture(t)
		cycle, err := f.d.acquire(context.Background(), Pending)
		if err != nil {
			t.Fatal(err)
		}
		fields, err := record.DataToMap(cycle)
		if err != nil {
			t.Fatal(err)
		}
		mutate(fields)
		key := cycleKey(cycle.Producer, cycle.Handler, Pending)
		if err = f.db.DB.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			return tx.Set(ctx, record.NewRecordWithData(key, &fields))
		}); err != nil {
			t.Fatal(err)
		}
		if _, err = f.d.acquire(context.Background(), Pending); err == nil {
			t.Fatal("corrupt checkpoint admitted", fields)
		}
	}
}

func TestScanCycleValidateRejectsNonInstantCutoff(t *testing.T) {
	f := newDispatchFixture(t)
	cycle := &scanCycle{
		SchemaVersion: 1,
		Producer:      f.d.executor.producer,
		Handler:       f.d.executor.handler,
		State:         Pending,
		PageSize:      f.d.options.PageSize,
		Epoch:         1,
		Cutoff:        time.Unix(1, 1).UTC(),
	}
	if err := cycle.validate(f.d, Pending); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
}

type cycleFaultTx struct {
	dal.ReadwriteTransaction
	get    func(record.Record) error
	set    func(record.Record) error
	insert func(record.Record) error
}

func (tx cycleFaultTx) Get(ctx context.Context, r record.Record) error {
	if tx.get != nil {
		if err := tx.get(r); err != nil {
			return err
		}
	}
	return tx.ReadwriteTransaction.Get(ctx, r)
}
func (tx cycleFaultTx) Set(ctx context.Context, r record.Record) error {
	if tx.set != nil {
		if err := tx.set(r); err != nil {
			return err
		}
	}
	return tx.ReadwriteTransaction.Set(ctx, r)
}
func (tx cycleFaultTx) Insert(ctx context.Context, r record.Record, opts ...dal.InsertOption) error {
	if tx.insert != nil {
		if err := tx.insert(r); err != nil {
			return err
		}
	}
	return tx.ReadwriteTransaction.Insert(ctx, r, opts...)
}

func TestCycleAndClaimStorageFailuresNeverAdvanceOrEraseUncertainty(t *testing.T) {
	for _, kind := range []string{"get-cycle", "insert-cycle", "claim-read", "claim-write", "checkpoint-write"} {
		t.Run(kind, func(t *testing.T) {
			f := newDispatchFixture(t)
			w := sample()
			schedule(t, f.db, w.Spec, 0)
			f.db.query = func(_ context.Context, q dal.StructuredQuery) (dal.RecordsReader, error) {
				if q.OrderBy()[0].Expression().String() == "nextAttemptAt" {
					return dispatchPage([]record.Record{discoveryRow(t, w)}, "after-work"), nil
				}
				return dispatchPage(nil, ""), nil
			}
			failure := errors.New("private-db-path")
			checkpointWrites := 0
			f.db.wrap = func(tx dal.ReadwriteTransaction) dal.ReadwriteTransaction {
				return cycleFaultTx{ReadwriteTransaction: tx, get: func(r record.Record) error {
					if kind == "get-cycle" && r.Key().Collection() == cycleCollection {
						return failure
					}
					if kind == "claim-read" && r.Key().Collection() == Collection {
						return failure
					}
					return nil
				}, insert: func(r record.Record) error {
					if kind == "insert-cycle" && r.Key().Collection() == cycleCollection {
						return failure
					}
					return nil
				}, set: func(r record.Record) error {
					if kind == "claim-write" && r.Key().Collection() == Collection {
						return failure
					}
					if kind == "checkpoint-write" && r.Key().Collection() == cycleCollection {
						checkpointWrites++
						if checkpointWrites == 1 {
							return failure
						}
					}
					return nil
				}}
			}
			report, err := f.d.Tick(context.Background())
			if !errors.Is(err, failure) || strings.Contains(err.Error(), "private") {
				t.Fatal(kind, report, err)
			}
			f.db.wrap = nil
			if kind == "claim-read" || kind == "claim-write" || kind == "checkpoint-write" {
				c := currentCycle(t, f, Pending)
				if c.Cursor != "" || c.Cutoff.IsZero() {
					t.Fatal("page advanced on failure", c)
				}
				stored, _ := Read(context.Background(), f.db, w.Spec.Ref)
				if kind == "checkpoint-write" {
					if stored.State != Leased || report.Enqueued != 1 {
						t.Fatal("committed claim lost", stored, report)
					}
				} else if stored.State != Pending || report.Claimed != 0 {
					t.Fatal("failed claim committed", stored, report)
				}
			}
		})
	}
}

func TestClaimUsesFreshClockAfterReads(t *testing.T) {
	f := newDispatchFixture(t)
	ctx := context.Background()
	w := sample()
	schedule(t, f.db, w.Spec, 0)
	cycle, err := f.d.acquire(ctx, Pending)
	if err != nil {
		t.Fatal(err)
	}
	f.db.wrap = func(tx dal.ReadwriteTransaction) dal.ReadwriteTransaction {
		return cycleFaultTx{ReadwriteTransaction: tx, get: func(r record.Record) error {
			if r.Key().Collection() == Collection {
				f.now = cycle.Lease.Until
			}
			return nil
		}}
	}
	if _, err = f.d.claim(ctx, cycle, Candidate{Ref: w.Spec.Ref, Generation: 1}); !errors.Is(err, ErrLeaseLost) {
		t.Fatal("expired scan lease committed", err)
	}
	f.db.wrap = nil
	work, _ := Read(ctx, f.db, w.Spec.Ref)
	if work.State != Pending {
		t.Fatal("stale claim escaped")
	}
	f.now = time.Time{}
	if _, err = f.d.acquire(ctx, Pending); !errors.Is(err, ErrInvalid) {
		t.Fatal("zero server clock admitted", err)
	}
}

func TestCycleRejectsClockRollbackExhaustionAndStorageCodecFailure(t *testing.T) {
	f := newDispatchFixture(t)
	ctx := context.Background()
	cycle, err := f.d.acquire(ctx, Pending)
	if err != nil {
		t.Fatal(err)
	}
	f.now = f.now.Add(-time.Second)
	if _, err = f.d.acquire(ctx, Pending); !errors.Is(err, ErrInvalid) {
		t.Fatal("clock rollback accepted", err)
	}
	f.now = time.Time{}
	if err = f.d.current(cycle, cycle); !errors.Is(err, ErrInvalid) {
		t.Fatal("invalid callback clock accepted", err)
	}
	if _, _, err = f.d.readCycle(ctx, workGetter{fields: map[string]any{"unsupported": make(chan int)}}, Pending); err == nil {
		t.Fatal("codec failure admitted")
	}
	f.now = sample().Spec.RunAt
	cycle.Cutoff = time.Time{}
	cycle.Lease = nil
	cycle.Cursor = ""
	cycle.Epoch = MaxGeneration
	if err = f.db.DB.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Set(ctx, record.NewRecordWithData(cycleKey(cycle.Producer, cycle.Handler, Pending), cycle))
	}); err != nil {
		t.Fatal(err)
	}
	if _, err = f.d.acquire(ctx, Pending); !errors.Is(err, ErrInvalid) {
		t.Fatal("exhausted epoch wrapped", err)
	}
}

func TestCycleCandidateValidationAndReadFailuresStopBeforeWrites(t *testing.T) {
	f := newDispatchFixture(t)
	ctx := context.Background()
	failure := errors.New("future cycle invariant")
	original := validatePreparedCycle
	defer func() { validatePreparedCycle = original }()
	validatePreparedCycle = func(*scanCycle, *Dispatcher, State) error { return failure }
	if _, err := f.d.acquire(ctx, Pending); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	validatePreparedCycle = original
	cycle, err := f.d.acquire(ctx, Pending)
	if err != nil {
		t.Fatal(err)
	}
	validatePreparedCycle = func(*scanCycle, *Dispatcher, State) error { return failure }
	if err = f.d.checkpoint(ctx, cycle, "position", false); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	if c := currentCycle(t, f, Pending); c.Cursor != "" {
		t.Fatal("failed candidate committed")
	}
	validatePreparedCycle = original
	f.db.wrap = func(tx dal.ReadwriteTransaction) dal.ReadwriteTransaction {
		return cycleFaultTx{ReadwriteTransaction: tx, get: func(record.Record) error { return failure }}
	}
	if err = f.d.checkpoint(ctx, cycle, "position", false); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	if _, err = f.d.claim(ctx, cycle, Candidate{}); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	f.db.wrap = nil
	// A valid stored record can exhaust its attempt counter. It must stop the
	// page without committing a claim or silently skipping the obligation.
	w := sample()
	w.Attempts = MaxGeneration
	setDispatchWork(t, f, w)
	if _, err = f.d.claim(ctx, cycle, Candidate{Ref: w.Spec.Ref, Generation: 1}); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
}
