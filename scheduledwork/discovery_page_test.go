package scheduledwork

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/dal-go/record"
)

type discoveryPageReader struct {
	*discoveryReader
	position  string
	cursorErr error
}

func (r *discoveryPageReader) Cursor() (string, error) { return r.position, r.cursorErr }

func TestDiscoveryPageAdvancesPastRejectedPrefixWithoutMutatingIt(t *testing.T) {
	w := sample()
	d := Discovery{Producer: w.Spec.Ref.Producer, Handler: w.Spec.Handler, State: Pending, Now: w.Spec.RunAt, Limit: 2}
	malformed := discoveryRow(t, w)
	(*malformed.Data().(*map[string]any))["effectKey"] = nil
	foreign := record.NewRecordWithData(record.NewKeyWithID(Collection, "foreign"), malformed.Data())
	reader := &discoveryPageReader{discoveryReader: &discoveryReader{rows: []record.Record{malformed, foreign}}, position: "opaque-position-after-rejected-rows"}
	ex := &discoveryExecutor{reader: reader}
	page, err := DiscoverPage(context.Background(), ex, d, "")
	if err != nil || len(page.Ready) != 0 || len(page.Rejected) != 2 || !page.Full || page.Next != reader.position || reader.calls != 2 || !reader.closed {
		t.Fatal("rejected prefix did not yield bounded continuation", page, err)
	}
	if ex.query.Offset() != 0 || ex.query.Limit() != 2 || ex.query.StartAfter() != "" {
		t.Fatal("unexpected first page window")
	}
	reader = &discoveryPageReader{discoveryReader: &discoveryReader{rows: []record.Record{discoveryRow(t, w)}}}
	ex.reader = reader
	last, err := DiscoverPage(context.Background(), ex, d, page.Next)
	if err != nil || len(last.Ready) != 1 || len(last.Rejected) != 0 || last.Full || last.Next != "" || string(ex.query.StartAfter()) != page.Next {
		t.Fatal("valid work behind poison prefix was hidden", last, err)
	}
	if !reflect.DeepEqual(ex.query.OrderBy()[0].Expression().String(), "nextAttemptAt") || ex.query.Offset() != 0 {
		t.Fatal("continuation changed eligibility ordering")
	}
	if (*malformed.Data().(*map[string]any))["effectKey"] != nil || foreign.Key().Collection() != Collection {
		t.Fatal("read API repaired untrusted records")
	}
	// A separate lease cycle starts at its own beginning and eligibility field.
	d.State = Leased
	ex.reader = &discoveryPageReader{discoveryReader: &discoveryReader{}}
	if _, err := DiscoverPage(context.Background(), ex, d, ""); err != nil || ex.query.OrderBy()[0].Expression().String() != "[claim.until]" || ex.query.StartAfter() != "" {
		t.Fatal("lease discovery inherited the pending position", err)
	}
}

func TestDiscoveryPageFailsClosedOnMissingRepeatedOrFailedContinuation(t *testing.T) {
	w := sample()
	d := Discovery{Producer: w.Spec.Ref.Producer, Handler: w.Spec.Handler, State: Pending, Now: w.Spec.RunAt, Limit: 1}
	failure := errors.New("cursor unavailable")
	for _, tc := range []struct {
		name, after, next   string
		cursorErr, closeErr error
	}{
		{name: "missing"},
		{name: "same position", after: "position", next: "position"},
		{name: "unsupported adapter", cursorErr: failure},
		{name: "oversized", next: strings.Repeat("x", (64<<10)+1)},
		{name: "failed close", next: "position", closeErr: failure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := &discoveryPageReader{discoveryReader: &discoveryReader{rows: []record.Record{discoveryRow(t, w)}, closeFailure: tc.closeErr}, position: tc.next, cursorErr: tc.cursorErr}
			page, err := DiscoverPage(context.Background(), &discoveryExecutor{reader: r}, d, tc.after)
			if err == nil || len(page.Ready) != 0 || len(page.Rejected) != 0 || page.Next != "" || !r.closed {
				t.Fatal("failed continuation leaked hints/progress", page, err)
			}
		})
	}
	if _, err := DiscoverPage(context.Background(), nil, d, strings.Repeat("x", (64<<10)+1)); !errors.Is(err, ErrInvalid) {
		t.Fatal("oversized input admitted", err)
	}
}
