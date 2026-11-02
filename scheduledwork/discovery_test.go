package scheduledwork

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
	"github.com/sneat-co/sneat-go-core/sneatcoretesting"
)

func TestDiscoveryFiltersBeforeLimitAndRecoversExpiredLeaseSeparately(t *testing.T) {
	ctx := context.Background()
	db := sneatcoretesting.NewInMemoryTestDB()
	now := sample().Spec.RunAt
	put := func(w Work) {
		t.Helper()
		key, err := w.Spec.Ref.Key()
		if err != nil {
			t.Fatal(err)
		}
		if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			return tx.Insert(ctx, record.NewRecordWithData(key, &w))
		}); err != nil {
			t.Fatal(err)
		}
	}
	// An early prefix of ineligible records must not fill the database limit.
	for i := 0; i < 120; i++ {
		w := sample()
		w.Spec.Ref.SubjectID = fmt.Sprintf("future-%03d", i)
		w.NextAttemptAt = now.Add(time.Hour)
		put(w)
		w = sample()
		w.Spec.Ref.SubjectID = fmt.Sprintf("terminal-%03d", i)
		w.State = Cancelled
		put(w)
	}
	for _, id := range []string{"due-a", "due-b", "due-c"} {
		w := sample()
		w.Spec.Ref.SubjectID = id
		put(w)
	}
	expired := sample()
	expired.Spec.Ref.SubjectID = "expired"
	expired.State = Leased
	expired.Attempts = 1
	expired.Claim = &Claim{Token: "old-token", At: now, Until: now.Add(time.Minute)}
	put(expired)
	live := expired
	live.Spec.Ref.SubjectID = "live"
	live.Claim = &Claim{Token: "live-token", At: now, Until: now.Add(time.Hour)}
	put(live)
	other := sample()
	other.Spec.Ref.Producer = "debtus"
	put(other)
	d := Discovery{Producer: "commitius", Handler: sample().Spec.Handler, State: Pending, Now: now.Add(time.Minute), Limit: 2}
	b, err := Discover(ctx, db, d)
	if err != nil || len(b.Ready) != 2 || len(b.Rejected) != 0 || !b.Full {
		t.Fatal("due work hidden by ineligible prefix", b, err)
	}
	// Both candidates have the same timestamp; full document identity breaks ties.
	k0, _ := b.Ready[0].Ref.Key()
	k1, _ := b.Ready[1].Ref.Key()
	if k0.String() >= k1.String() {
		t.Fatal("unstable tie ordering", k0, k1)
	}
	for _, c := range b.Ready {
		if c.Ref.SubjectID[:4] != "due-" {
			t.Fatal("ineligible row returned", c)
		}
	}
	d.State = Leased
	b, err = Discover(ctx, db, d)
	if err != nil || len(b.Ready) != 1 || b.Ready[0].Ref != expired.Spec.Ref || b.Full {
		t.Fatal("expired/live lease selection wrong", b, err)
	}
	// Discovery grants no stale authority: cancellation after selection wins.
	d.State = Pending
	b, _ = Discover(ctx, db, d)
	picked := b.Ready[0]
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		c, err := PrepareCancel(ctx, tx, picked.Ref, picked.Generation)
		if err != nil {
			return err
		}
		return c.Apply(ctx)
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, _, err := PrepareClaim(ctx, tx, picked.Ref, picked.Generation, d.Now, time.Minute)
		return err
	}); !errors.Is(err, ErrConflict) {
		t.Fatal("discovery bypassed fresh claim", err)
	}
}

type discoveryExecutor struct {
	dal.QueryExecutor
	reader  dal.RecordsReader
	failure error
	query   dal.StructuredQuery
}

func (e *discoveryExecutor) ExecuteQueryToRecordsReader(_ context.Context, q dal.Query) (dal.RecordsReader, error) {
	e.query = q.(dal.StructuredQuery)
	return e.reader, e.failure
}

type discoveryReader struct {
	dal.RecordsReader
	rows                  []record.Record
	calls                 int
	failure, closeFailure error
	closed                bool
}

func (r *discoveryReader) Next() (record.Record, error) {
	r.calls++
	if len(r.rows) == 0 {
		if r.failure != nil {
			return nil, r.failure
		}
		return nil, dal.ErrNoMoreRecords
	}
	row := r.rows[0]
	r.rows = r.rows[1:]
	return row, nil
}
func (r *discoveryReader) Close() error { r.closed = true; return r.closeFailure }
func discoveryRow(t *testing.T, w Work) record.Record {
	t.Helper()
	fields, err := record.DataToMap(w)
	if err != nil {
		t.Fatal(err)
	}
	key, _ := w.Spec.Ref.Key()
	return record.NewRecordWithData(key, &fields)
}

func TestDiscoveryQueryAndUntrustedResults(t *testing.T) {
	w := sample()
	d := Discovery{Producer: w.Spec.Ref.Producer, Handler: w.Spec.Handler, State: Pending, Now: w.Spec.RunAt, Limit: 10}
	good := discoveryRow(t, w)
	malformed := discoveryRow(t, w)
	(*malformed.Data().(*map[string]any))["effectKey"] = nil
	foreign := record.NewRecordWithData(record.NewKeyWithID(Collection, "foreign"), good.Data())
	mismatch := w
	mismatch.Spec.Handler = "other"
	future := w
	future.NextAttemptAt = d.Now.Add(time.Hour)
	badShape := record.NewRecordWithData(good.Key(), &w)
	reader := &discoveryReader{rows: []record.Record{nil, record.NewRecordWithIncompleteKey(Collection, reflect.String, nil), badShape, malformed, foreign, discoveryRow(t, mismatch), discoveryRow(t, future), good}}
	ex := &discoveryExecutor{reader: reader}
	b, err := Discover(context.Background(), ex, d)
	if err != nil || len(b.Ready) != 1 || len(b.Rejected) != 7 || !reader.closed {
		t.Fatal("untrusted results were dispatched", b, err)
	}
	var codes []string
	for _, r := range b.Rejected {
		codes = append(codes, r.Code)
	}
	if !reflect.DeepEqual(codes, []string{"invalid-key", "invalid-key", "invalid-shape", "invalid-work", "wrong-owner-key", "wrong-registration-or-state", "not-due"}) {
		t.Fatal(codes)
	}
	q := ex.query
	if q.Limit() != 10 || q.Offset() != 0 || q.From().Base().Name() != Collection || len(q.OrderBy()) != 2 || q.OrderBy()[0].Expression().String() != "nextAttemptAt" || !reflect.DeepEqual(q.OrderBy()[1].Expression(), dal.DocumentID()) {
		t.Fatal("unbounded or unstable query", q)
	}
	expected := dal.NewGroupCondition(dal.And, dal.WhereField("spec.ref.producer", dal.Equal, string(d.Producer)), dal.WhereField("spec.handler", dal.Equal, d.Handler), dal.WhereField("state", dal.Equal, string(Pending)), dal.WhereField("nextAttemptAt", dal.LessOrEqual, d.Now))
	if !reflect.DeepEqual(q.Where(), expected) {
		t.Fatal("eligibility must be in database predicates", q.Where())
	}
	// A malicious/buggy reader that ignores Limit cannot induce more reads.
	reader = &discoveryReader{rows: []record.Record{good, good, good}}
	d.Limit = 1
	b, err = Discover(context.Background(), &discoveryExecutor{reader: reader}, d)
	if err != nil || len(b.Ready) != 1 || reader.calls != 1 || !b.Full {
		t.Fatal("unbounded read", b, reader.calls, err)
	}
}

func TestDiscoveryFailsClosedOnStorageAndInvalidParameters(t *testing.T) {
	w := sample()
	d := Discovery{Producer: w.Spec.Ref.Producer, Handler: w.Spec.Handler, State: Pending, Now: w.Spec.RunAt, Limit: 10}
	ctx := context.Background()
	want := errors.New("storage unavailable")
	for _, change := range []func(*Discovery){func(d *Discovery) { d.Producer = "" }, func(d *Discovery) { d.Handler = "" }, func(d *Discovery) { d.Now = time.Time{} }, func(d *Discovery) { d.Limit = 0 }, func(d *Discovery) { d.Limit = 101 }, func(d *Discovery) { d.State = Completed }} {
		bad := d
		change(&bad)
		if _, err := Discover(ctx, nil, bad); !errors.Is(err, ErrInvalid) {
			t.Fatal("invalid discovery accepted", err)
		}
	}
	if _, err := Discover(ctx, nil, d); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	if _, err := Discover(ctx, &discoveryExecutor{failure: want}, d); !errors.Is(err, want) {
		t.Fatal(err)
	}
	if _, err := Discover(ctx, &discoveryExecutor{}, d); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	for _, onClose := range []bool{false, true} {
		r := &discoveryReader{rows: []record.Record{discoveryRow(t, w)}}
		if onClose {
			r.closeFailure = want
		} else {
			r.failure = want
		}
		b, err := Discover(ctx, &discoveryExecutor{reader: r}, d)
		if !errors.Is(err, want) || len(b.Ready) != 0 || !r.closed {
			t.Fatal("partial results escaped failed stream", b, err)
		}
	}
}

func TestDiscoveryAcceptsFirestoreStringAncestorIDsAndRejectsOtherPaths(t *testing.T) {
	w := sample()
	d := Discovery{Producer: w.Spec.Ref.Producer, Handler: w.Spec.Handler, State: Pending, Now: w.Spec.RunAt, Limit: 10}
	expected, _ := w.Spec.Ref.Key()
	fields, _ := record.DataToMap(w)
	key := func(root, space, module, collection string, id any) *record.Key {
		return record.NewKeyWithParentAndID(record.NewKeyWithParentAndID(record.NewKeyWithID(root, space), "ext", module), collection, id)
	}
	// Actual collection-group snapshot shape: no named SpaceID/ExtID types.
	good := key("spaces", string(w.Spec.Ref.SpaceID), "core", Collection, expected.ID)
	if c, code := d.candidate(record.NewRecordWithData(good, &fields)); code != "" || c.Ref != w.Spec.Ref {
		t.Fatal("real Firestore ancestor types rejected", c, code)
	}
	for _, bad := range []*record.Key{
		key("spaces", "other-space", "core", Collection, expected.ID),
		key("spaces", string(w.Spec.Ref.SpaceID), "other-module", Collection, expected.ID),
		key("other-roots", string(w.Spec.Ref.SpaceID), "core", Collection, expected.ID),
		key("spaces", string(w.Spec.Ref.SpaceID), "core", "other-collection", expected.ID),
		key("spaces", string(w.Spec.Ref.SpaceID), "core", Collection, "other-leaf"),
		record.NewKeyWithParentAndID(good, Collection, expected.ID),
		record.NewKeyWithID(Collection, expected.ID),
	} {
		if _, code := d.candidate(record.NewRecordWithData(bad, &fields)); code != "wrong-owner-key" {
			t.Fatal("foreign ancestor accepted", bad, code)
		}
	}
	if sameStringKey(expected, key("spaces", string(w.Spec.Ref.SpaceID), "core", Collection, int64(123))) {
		t.Fatal("numeric ID normalized to string")
	}
}
