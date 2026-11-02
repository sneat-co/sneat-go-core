package scheduledwork

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
	"github.com/sneat-co/sneat-go-core/coretypes"
)

// Discovery selects one registered producer/handler at a time. State must be
// Pending or Leased; expired leases have their own query so they cannot crowd
// out fresh due work. These values come from server registration, never a user.
type Discovery struct {
	Producer coretypes.ExtID
	Handler  string
	State    State
	Now      time.Time
	Limit    int
}

// Candidate is a discovery hint, not execution authority. PrepareClaim must
// reload and fence this identity/generation in a fresh transaction.
type Candidate struct {
	Ref        Ref
	Generation int64
	EligibleAt time.Time
}

// Rejection is an operational diagnostic only. It grants no right to execute,
// mutate or delete the row; malformed/foreign keys require separate recovery.
type Rejection struct {
	Key  *record.Key
	Code string
}

type Batch struct {
	Ready    []Candidate
	Rejected []Rejection
	// Full means the database batch was filled, not that there is no more work.
	Full bool
}

// Page keeps continuation separate from execution hints. Persist Next only
// after handling this batch; reuse exactly the same Discovery, including Now
// and Limit, until EOF. Pending and expired-lease cycles are independent.
type Page struct {
	Batch
	Next string
}

func (d Discovery) query(after string) (dal.StructuredQuery, error) {
	if !identifier(string(d.Producer)) || !identifier(d.Handler) || !instant(d.Now) || d.Limit < 1 || d.Limit > 100 {
		return nil, ErrInvalid
	}
	field := "nextAttemptAt"
	switch d.State {
	case Pending:
	case Leased:
		field = "claim.until"
	default:
		return nil, ErrInvalid
	}
	builder := dal.From(dal.NewCollectionGroupRef(Collection, "")).NewQuery().
		WhereField("spec.ref.producer", dal.Equal, string(d.Producer)).
		WhereField("spec.handler", dal.Equal, d.Handler).
		WhereField("state", dal.Equal, string(d.State)).
		WhereField(field, dal.LessOrEqual, d.Now.UTC()).
		OrderBy(dal.AscendingField(field), dal.Ascending(dal.DocumentID())).
		Limit(d.Limit)
	if after != "" {
		builder = builder.StartAfter(dal.Cursor(after))
	}
	return builder.SelectIntoRecord(func() record.Record {
		fields := make(map[string]any)
		return record.NewRecordWithIncompleteKey(Collection, reflect.String, &fields)
	}), nil
}

func (d Discovery) candidate(rec record.Record) (Candidate, string) {
	if rec == nil || rec.Key() == nil || rec.Key().Validate() != nil {
		return Candidate{}, "invalid-key"
	}
	if id, ok := rec.Key().ID.(string); !ok || id == "" {
		return Candidate{}, "invalid-key"
	}
	fields, ok := rec.Data().(*map[string]any)
	if !ok || fields == nil {
		return Candidate{}, "invalid-shape"
	}
	w, err := decodeWork(*fields)
	if err != nil {
		return Candidate{}, "invalid-work"
	}
	key, _ := w.Spec.Ref.Key() // decodeWork validated the reference.
	if !sameStringKey(key, rec.Key()) {
		return Candidate{}, "wrong-owner-key"
	}
	if w.Spec.Ref.Producer != d.Producer || w.Spec.Handler != d.Handler || w.State != d.State {
		return Candidate{}, "wrong-registration-or-state"
	}
	at := w.NextAttemptAt
	if d.State == Leased {
		at = w.Claim.Until
	}
	if at.After(d.Now) {
		return Candidate{}, "not-due"
	}
	return Candidate{Ref: w.Spec.Ref, Generation: w.Generation, EligibleAt: at}, ""
}

// Firestore collection-group readers reconstruct ancestor IDs as plain
// strings; owning factories may use named string types (SpaceID/ExtID).
// Compare every collection and string-valued segment at exact depth instead
// of Go interface identity, or path serialization that can hide bad structure.
func sameStringKey(expected, actual *record.Key) bool {
	for expected != nil && actual != nil {
		id := reflect.ValueOf(actual.ID)
		if (actual.IDKind != reflect.Invalid && actual.IDKind != reflect.String) || id.Kind() != reflect.String || actual.Collection() != expected.Collection() || id.String() != reflect.ValueOf(expected.ID).String() {
			return false
		}
		expected, actual = expected.Parent(), actual.Parent()
	}
	return expected == nil && actual == nil
}

// Discover reads at most Limit records using database predicates and an ordered
// collection-group query. It never enumerates commitments, all jobs or Spaces.
// No action is returned on query/stream/close failure. Rejected rows are never
// dispatched; handling poison-prefix recovery is required before enabling a
// tick in production. This function alone is not that recovery policy.
func Discover(ctx context.Context, executor dal.QueryExecutor, d Discovery) (batch Batch, err error) {
	page, err := discover(ctx, executor, d, "", false)
	return page.Batch, err
}

// DiscoverPage uses an adapter-qualified opaque cursor, including the last
// rejected row, to continue past a poison prefix without OFFSET or rewriting
// foreign metadata. Cursor failures discard all hints. This is a live traversal,
// not a database snapshot: restart from the beginning with a fresh Now after EOF
// so changed/newly due work behind the boundary is reconsidered. Callers must
// bound pages per tick and save cycle progress; this API never scans in a loop.
func DiscoverPage(ctx context.Context, executor dal.QueryExecutor, d Discovery, after string) (Page, error) {
	if len(after) > 64<<10 {
		return Page{}, ErrInvalid
	}
	return discover(ctx, executor, d, after, true)
}

func discover(ctx context.Context, executor dal.QueryExecutor, d Discovery, after string, paged bool) (page Page, err error) {
	query, err := d.query(after)
	if err != nil {
		return Page{}, err
	}
	if executor == nil {
		return Page{}, ErrInvalid
	}
	reader, err := executor.ExecuteQueryToRecordsReader(ctx, query)
	if err != nil {
		return Page{}, fmt.Errorf("discover scheduled work: %w", err)
	}
	if reader == nil {
		return Page{}, ErrInvalid
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil && err == nil {
			page = Page{}
			err = fmt.Errorf("close scheduled work discovery: %w", closeErr)
		}
	}()
	for i := 0; i < d.Limit; i++ {
		rec, readErr := reader.Next()
		if errors.Is(readErr, dal.ErrNoMoreRecords) {
			return page, nil
		}
		if readErr != nil {
			return Page{}, fmt.Errorf("read scheduled work discovery: %w", readErr)
		}
		candidate, code := d.candidate(rec)
		if code != "" {
			var key *record.Key
			if rec != nil {
				key = rec.Key()
			}
			page.Rejected = append(page.Rejected, Rejection{Key: key, Code: code})
		} else {
			page.Ready = append(page.Ready, candidate)
		}
	}
	page.Full = true
	if paged {
		page.Next, err = reader.Cursor()
		if err != nil {
			return Page{}, fmt.Errorf("scheduled work continuation: %w", err)
		}
		if page.Next == "" || page.Next == after || len(page.Next) > 64<<10 {
			return Page{}, ErrInvalid
		}
	}
	return page, nil
}
