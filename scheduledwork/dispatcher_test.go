package scheduledwork

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
	"github.com/sneat-co/sneat-go-core/sneatcoretesting"
	"github.com/strongo/delaying"
)

type dispatchDB struct {
	dal.DB
	query func(context.Context, dal.StructuredQuery) (dal.RecordsReader, error)
	wrap  func(dal.ReadwriteTransaction) dal.ReadwriteTransaction
}

func (db *dispatchDB) ExecuteQueryToRecordsReader(ctx context.Context, q dal.Query) (dal.RecordsReader, error) {
	return db.query(ctx, q.(dal.StructuredQuery))
}
func (db *dispatchDB) RunReadwriteTransaction(ctx context.Context, f dal.RWTxWorker, opts ...dal.TransactionOption) error {
	return db.DB.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		if db.wrap != nil {
			tx = db.wrap(tx)
		}
		return f(ctx, tx)
	}, opts...)
}

type dispatchFixture struct {
	d      *Dispatcher
	db     *dispatchDB
	now    time.Time
	sends  []ClaimRef
	send   func(context.Context, ClaimRef) error
	worker delayedClaimWorker
}

func newDispatchFixture(t *testing.T) *dispatchFixture {
	t.Helper()
	f := &dispatchFixture{now: sample().Spec.RunAt, db: &dispatchDB{DB: sneatcoretesting.NewInMemoryTestDB()}}
	f.db.query = func(context.Context, dal.StructuredQuery) (dal.RecordsReader, error) {
		return dispatchPage(nil, ""), nil
	}
	ex, err := RegisterExecutor(func(name string, fn any) delaying.Delayer {
		f.worker = fn.(delayedClaimWorker)
		return delaying.NewDelayer(name, fn, func(ctx context.Context, _ delaying.Params, args ...any) error {
			claim := ClaimRef{Ref: sample().Spec.Ref, Generation: args[3].(int64), Token: args[4].(string)}
			claim.Ref.SpaceID = sample().Spec.Ref.SpaceID
			claim.Ref.SubjectID = args[1].(string)
			claim.Ref.Action = args[2].(string)
			f.sends = append(f.sends, claim)
			if f.send != nil {
				return f.send(ctx, claim)
			}
			return nil
		}, func(context.Context, delaying.Params, ...[]any) error { return errors.New("unexpected batch enqueue") })
	}, sample().Spec.Ref.Producer, sample().Spec.Handler, "scheduled", func(ctx context.Context, claim ClaimRef) error {
		return f.db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			execution, err := PrepareExecution(ctx, tx, claim, func() time.Time { return f.now })
			if err != nil {
				return err
			}
			return execution.Complete(ctx)
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	f.d, err = NewDispatcher(f.db, ex, DispatcherOptions{PageSize: 1, PagesPerState: 1, JobLease: time.Minute, ScanLease: time.Minute, StateTimeout: time.Second, EnqueueTimeout: 100 * time.Millisecond, Now: func() time.Time { return f.now }})
	if err != nil {
		t.Fatal(err)
	}
	return f
}
func dispatchPage(rows []record.Record, next string) dal.RecordsReader {
	return &discoveryPageReader{discoveryReader: &discoveryReader{rows: rows}, position: next}
}
func currentCycle(t *testing.T, f *dispatchFixture, state State) *scanCycle {
	t.Helper()
	_, c, err := f.d.readCycle(context.Background(), f.db, state)
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func setDispatchWork(t *testing.T, f *dispatchFixture, w Work) {
	t.Helper()
	key, _ := w.Spec.Ref.Key()
	if err := f.db.DB.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return tx.Set(ctx, record.NewRecordWithData(key, &w))
	}); err != nil {
		t.Fatal(err)
	}
}

func TestDispatcherResumesRejectedPagesAndRestartsOnlyAtEOF(t *testing.T) {
	f := newDispatchFixture(t)
	w := sample()
	schedule(t, f.db, w.Spec, 0)
	poison := discoveryRow(t, w)
	(*poison.Data().(*map[string]any))["unknownField"] = "private-data"
	step := 0
	firstCutoff := f.now
	f.db.query = func(_ context.Context, q dal.StructuredQuery) (dal.RecordsReader, error) {
		if q.Offset() != 0 || q.Limit() != 1 {
			t.Fatal("unbounded discovery")
		}
		if q.OrderBy()[0].Expression().String() == "[claim.until]" {
			return dispatchPage(nil, ""), nil
		}
		step++
		switch step {
		case 1:
			if q.StartAfter() != "" {
				t.Fatal("unexpected initial position")
			}
			return dispatchPage([]record.Record{poison}, "after-poison"), nil
		case 2:
			if q.StartAfter() != "after-poison" {
				t.Fatal("lost durable boundary")
			}
			if c := currentCycle(t, f, Pending); !c.Cutoff.Equal(firstCutoff) || c.Epoch != 1 {
				t.Fatal("reset live cycle", c)
			}
			return dispatchPage([]record.Record{discoveryRow(t, w)}, "after-valid"), nil
		case 3:
			if q.StartAfter() != "after-valid" {
				t.Fatal("lost second boundary")
			}
			return dispatchPage(nil, ""), nil
		case 4:
			if q.StartAfter() != "" {
				t.Fatal("new cycle did not restart")
			}
			if c := currentCycle(t, f, Pending); !c.Cutoff.Equal(f.now) || c.Epoch != 2 {
				t.Fatal("fresh cycle missing", c)
			}
			return dispatchPage(nil, ""), nil
		default:
			t.Fatal("extra page")
		}
		return nil, ErrInvalid
	}
	report, err := f.d.Tick(context.Background())
	if err != nil || report.Rejected != 1 || report.Claimed != 0 || report.Pages != 2 {
		t.Fatal(report, err)
	}
	c := currentCycle(t, f, Pending)
	if c.Lease != nil || c.Cursor != "after-poison" || !c.Cutoff.Equal(firstCutoff) {
		t.Fatal("budget erased continuation", c)
	}
	f.now = f.now.Add(time.Minute)
	report, err = f.d.Tick(context.Background())
	if err != nil || report.Claimed != 1 || report.Enqueued != 1 || len(f.sends) != 1 {
		t.Fatal(report, err)
	}
	stored, err := Read(context.Background(), f.db, w.Spec.Ref)
	if err != nil || stored.State != Leased {
		t.Fatal("enqueue completed business work", stored, err)
	}
	for i := 0; i < 2; i++ {
		if _, err = f.d.Tick(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if c = currentCycle(t, f, Pending); !c.Cutoff.IsZero() || c.Cursor != "" || c.Lease != nil || c.Epoch != 2 {
		t.Fatal("EOF not explicit", c)
	}
}

func TestDispatcherUnknownEnqueueRecoversThroughIndependentExpiredLeaseCycle(t *testing.T) {
	f := newDispatchFixture(t)
	w := sample()
	schedule(t, f.db, w.Spec, 0)
	f.d.options.PageSize = 2
	f.send = func(context.Context, ClaimRef) error { return errors.New("provider private-token unknown reply") }
	f.db.query = func(_ context.Context, q dal.StructuredQuery) (dal.RecordsReader, error) {
		current, err := Read(context.Background(), f.db, w.Spec.Ref)
		if err != nil {
			t.Fatal(err)
		}
		field := q.OrderBy()[0].Expression().String()
		if (field == "nextAttemptAt" && current.State == Pending) || (field == "[claim.until]" && current.State == Leased && !f.now.Before(current.Claim.Until)) {
			return dispatchPage([]record.Record{discoveryRow(t, *current)}, ""), nil
		}
		return dispatchPage(nil, ""), nil
	}
	report, err := f.d.Tick(context.Background())
	if err == nil || strings.Contains(err.Error(), "private") || report.Unconfirmed != 1 || report.Claimed != 1 {
		t.Fatal(report, err)
	}
	old := f.sends[0]
	work, _ := Read(context.Background(), f.db, w.Spec.Ref)
	if work.State != Leased || work.Claim.Token != old.Token {
		t.Fatal("uncertainty was erased")
	}
	f.now = f.now.Add(2 * time.Minute)
	f.send = nil
	report, err = f.d.Tick(context.Background())
	if err != nil || report.Enqueued != 1 || report.Claimed != 1 || len(f.sends) != 2 || old.Token == f.sends[1].Token {
		t.Fatal(report, err)
	}
	if err = f.worker(context.Background(), string(old.Ref.SpaceID), old.Ref.SubjectID, old.Ref.Action, old.Generation, old.Token); !errors.Is(err, ErrLeaseLost) {
		t.Fatal("stale delivery accepted", err)
	}
	latest := f.sends[1]
	if err = f.worker(context.Background(), string(latest.Ref.SpaceID), latest.Ref.SubjectID, latest.Ref.Action, latest.Generation, latest.Token); err != nil {
		t.Fatal(err)
	}
	work, _ = Read(context.Background(), f.db, w.Spec.Ref)
	if work.State != Completed {
		t.Fatal("real owning worker did not complete")
	}
}

func TestDispatcherRejectsChangedDiscoveryHints(t *testing.T) {
	for _, kind := range []string{"generation", "handler", "state", "future-pending", "deleted", "lease-renewed"} {
		t.Run(kind, func(t *testing.T) {
			f := newDispatchFixture(t)
			w := sample()
			schedule(t, f.db, w.Spec, 0)
			hint := discoveryRow(t, w)
			if kind == "lease-renewed" {
				w.State = Leased
				w.Attempts = 1
				w.Claim = &Claim{Token: "old", At: f.now.Add(-time.Minute), Until: f.now}
				// RunAt must precede the original claim.
				w.Spec.RunAt = w.Claim.At
				w.NextAttemptAt = w.Spec.RunAt
				setDispatchWork(t, f, w)
				hint = discoveryRow(t, w)
			}
			f.db.query = func(_ context.Context, q dal.StructuredQuery) (dal.RecordsReader, error) {
				leased := q.OrderBy()[0].Expression().String() == "[claim.until]"
				if leased != (kind == "lease-renewed") {
					return dispatchPage(nil, ""), nil
				}
				switch kind {
				case "generation":
					w.Generation++
				case "handler":
					w.Spec.Handler = "changed-handler"
				case "state":
					w.State = Cancelled
				case "future-pending":
					w.NextAttemptAt = f.now.Add(time.Minute)
				case "lease-renewed":
					w.Claim.Until = f.now.Add(time.Minute)
				case "deleted":
					key, _ := w.Spec.Ref.Key()
					if err := f.db.DB.RunReadwriteTransaction(context.Background(), func(ctx context.Context, tx dal.ReadwriteTransaction) error { return tx.Delete(ctx, key) }); err != nil {
						t.Fatal(err)
					}
				}
				if kind != "deleted" {
					setDispatchWork(t, f, w)
				}
				return dispatchPage([]record.Record{hint}, "boundary"), nil
			}
			report, err := f.d.Tick(context.Background())
			if err != nil || report.Stale != 1 || report.Claimed != 0 || len(f.sends) != 0 {
				t.Fatal(kind, report, err)
			}
		})
	}
}

func TestDispatcherStateTimeoutDoesNotStarveLeaseRecovery(t *testing.T) {
	f := newDispatchFixture(t)
	f.d.options.StateTimeout = 10 * time.Millisecond
	pending, leased := false, false
	f.db.query = func(ctx context.Context, q dal.StructuredQuery) (dal.RecordsReader, error) {
		if q.OrderBy()[0].Expression().String() == "nextAttemptAt" {
			pending = true
			if _, ok := ctx.Deadline(); !ok {
				t.Fatal("missing deadline")
			}
			<-ctx.Done()
			return nil, ctx.Err()
		}
		leased = true
		if ctx.Err() != nil {
			t.Fatal("pending timeout consumed recovery allowance")
		}
		return dispatchPage(nil, ""), nil
	}
	_, err := f.d.Tick(context.Background())
	if !errors.Is(err, context.DeadlineExceeded) || !pending || !leased {
		t.Fatal(err, pending, leased)
	}
	if c := currentCycle(t, f, Pending); c.Lease != nil || c.Cutoff.IsZero() {
		t.Fatal("timeout erased cycle", c)
	}
}

func TestDispatcherConfigurationIsBoundedAndReportsSafeErrors(t *testing.T) {
	f := newDispatchFixture(t)
	for _, mutate := range []func(*DispatcherOptions){
		func(o *DispatcherOptions) { o.PageSize = 0 }, func(o *DispatcherOptions) { o.PageSize = 101 },
		func(o *DispatcherOptions) { o.PagesPerState = 0 }, func(o *DispatcherOptions) { o.PagesPerState = 5 },
		func(o *DispatcherOptions) { o.JobLease = 0 }, func(o *DispatcherOptions) { o.JobLease = time.Hour + 1 },
		func(o *DispatcherOptions) { o.ScanLease = 11 * time.Minute }, func(o *DispatcherOptions) { o.ScanLease = time.Millisecond },
		func(o *DispatcherOptions) { o.StateTimeout = 0 }, func(o *DispatcherOptions) { o.StateTimeout = 3 * time.Minute },
		func(o *DispatcherOptions) { o.EnqueueTimeout = 31 * time.Second }, func(o *DispatcherOptions) { o.EnqueueTimeout = 2 * time.Second },
		func(o *DispatcherOptions) { o.EnqueueTimeout = time.Nanosecond }, func(o *DispatcherOptions) { o.Now = nil },
	} {
		o := f.d.options
		mutate(&o)
		if d, err := NewDispatcher(f.db, f.d.executor, o); d != nil || !errors.Is(err, ErrInvalid) {
			t.Fatal(o, d, err)
		}
	}
	for _, ex := range []*Executor{nil, {}, {producer: "bad/producer"}, {producer: "p", handler: "h", queue: "q"}} {
		if _, err := NewDispatcher(f.db, ex, f.d.options); !errors.Is(err, ErrInvalid) {
			t.Fatal("invalid executor accepted")
		}
	}
	if _, err := NewDispatcher(nil, f.d.executor, f.d.options); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	var nilDispatcher *Dispatcher
	if _, err := nilDispatcher.Tick(context.Background()); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := f.d.Tick(nil); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	failure := errors.New("private-cursor")
	f.db.query = func(context.Context, dal.StructuredQuery) (dal.RecordsReader, error) { return nil, failure }
	_, err := f.d.Tick(context.Background())
	var operational *DispatchError
	if !errors.As(err, &operational) || !errors.Is(err, failure) || err.Error() != "scheduled work dispatch incomplete" {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = f.d.Tick(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal("parent cancellation lost", err)
	}
}

func TestDispatcherCancellationBetweenBoundedUnits(t *testing.T) {
	f := newDispatchFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var report DispatchReport
	if err := f.d.runState(ctx, Pending, time.Second, &report); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	cycle, err := f.d.acquire(context.Background(), Pending)
	if err != nil {
		t.Fatal(err)
	}
	if err = f.d.pages(ctx, cycle, &report); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	ctx, cancel = context.WithCancel(context.Background())
	f.db.query = func(context.Context, dal.StructuredQuery) (dal.RecordsReader, error) {
		cancel()
		return dispatchPage([]record.Record{discoveryRow(t, sample())}, "position"), nil
	}
	if err = f.d.pages(ctx, cycle, &report); !errors.Is(err, context.Canceled) || report.Claimed != 0 {
		t.Fatal("cancellation after query admitted mutation", report, err)
	}
}
