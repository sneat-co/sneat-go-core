package scheduledwork

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
	"github.com/sneat-co/sneat-go-core/coretypes"
)

const cycleCollection = "scheduledWorkCycles"

var errCycleBusy = errors.New("scheduled work scan already leased")

// Keep future invariant failures testable before handing candidates to storage.
var validatePreparedCycle = (*scanCycle).validate

// A service-wide query checkpoint, deliberately NOT a business job. This
// explicit global Core key never invokes the optional-Space fallback factory.
// Clients must be denied this entire subtree before a dispatcher is enabled.
func cycleKey(producer coretypes.ExtID, handler string, state State) *record.Key {
	digest := sha256.Sum256([]byte(string(producer) + "\x00" + handler + "\x00" + string(state)))
	return record.NewKeyWithParentAndID(record.NewKeyWithID("ext", "core"), cycleCollection, "cycle_"+hex.EncodeToString(digest[:]))
}

type scanCycle struct {
	SchemaVersion int             `json:"schemaVersion" firestore:"schemaVersion"`
	Producer      coretypes.ExtID `json:"producer" firestore:"producer"`
	Handler       string          `json:"handler" firestore:"handler"`
	State         State           `json:"state" firestore:"state"`
	PageSize      int             `json:"pageSize" firestore:"pageSize"`
	Epoch         int64           `json:"epoch" firestore:"epoch"`
	Cutoff        time.Time       `json:"cutoff" firestore:"cutoff"`
	Cursor        string          `json:"cursor" firestore:"cursor"`
	Lease         *Claim          `json:"lease,omitempty" firestore:"lease,omitempty"`
}

func (c *scanCycle) validate(d *Dispatcher, state State) error {
	if c.SchemaVersion != 1 || c.Producer != d.executor.producer || c.Handler != d.executor.handler || c.State != state || c.PageSize != d.options.PageSize || c.Epoch < 1 || c.Epoch > MaxGeneration || len(c.Cursor) > 64<<10 {
		return ErrInvalid
	}
	if c.Cutoff.IsZero() {
		if c.Cursor != "" || c.Lease != nil {
			return ErrInvalid
		}
		return nil
	}
	if !instant(c.Cutoff) {
		return ErrInvalid
	}
	if c.Lease != nil && (!identifier(c.Lease.Token) || !instant(c.Lease.At) || !instant(c.Lease.Until) || c.Lease.At.Before(c.Cutoff) || !c.Lease.Until.After(c.Lease.At) || c.Lease.Until.Sub(c.Lease.At) > 10*time.Minute) {
		return ErrInvalid
	}
	return nil
}

func (d *Dispatcher) readCycle(ctx context.Context, tx dal.Getter, state State) (record.Record, *scanCycle, error) {
	key := cycleKey(d.executor.producer, d.executor.handler, state)
	fields := make(map[string]any)
	if err := tx.Get(ctx, record.NewRecordWithData(key, &fields)); err != nil {
		return nil, nil, err
	}
	encoded, err := json.Marshal(fields)
	if err != nil {
		return nil, nil, err
	}
	top, err := objectFields(encoded, []string{"lease"}, "schemaVersion", "producer", "handler", "state", "pageSize", "epoch", "cutoff", "cursor", "lease")
	if err != nil {
		return nil, nil, err
	}
	// Empty strings/zero cutoff have meaning: missing fields must not be silently
	// interpreted as a deliberate restart or advance.
	for _, key := range []string{"schemaVersion", "producer", "handler", "state", "pageSize", "epoch", "cutoff", "cursor"} {
		if _, ok := top[key]; !ok {
			return nil, nil, ErrInvalid
		}
	}
	if raw := top["lease"]; len(raw) > 0 && !bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		if _, err := objectFields(raw, nil, "token", "at", "until"); err != nil {
			return nil, nil, err
		}
	}
	c := new(scanCycle)
	if err := record.MapToData(c, fields); err != nil {
		return nil, nil, err
	}
	if err := c.validate(d, state); err != nil {
		return nil, nil, err
	}
	return record.NewRecordWithData(key, c).SetError(nil), c, nil
}

func (d *Dispatcher) now() time.Time { return d.options.Now().UTC().Truncate(time.Microsecond) }

func (d *Dispatcher) acquire(ctx context.Context, state State) (*scanCycle, error) {
	var acquired *scanCycle
	err := d.db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		rec, c, err := d.readCycle(ctx, tx, state)
		missing := record.IsNotFound(err)
		if err != nil && !missing {
			return err
		}
		now := d.now()
		if !instant(now) {
			return ErrInvalid
		}
		if missing {
			c = &scanCycle{SchemaVersion: 1, Producer: d.executor.producer, Handler: d.executor.handler, State: state, PageSize: d.options.PageSize}
			rec = record.NewRecordWithData(cycleKey(c.Producer, c.Handler, state), c)
		}
		if now.Before(c.Cutoff) || (c.Lease != nil && now.Before(c.Lease.At)) {
			return ErrInvalid
		}
		if c.Lease != nil && now.Before(c.Lease.Until) {
			return errCycleBusy
		}
		if c.Cutoff.IsZero() {
			if c.Epoch == MaxGeneration {
				return ErrInvalid
			}
			c.Epoch++
			c.Cutoff = now
		}
		c.Lease = &Claim{Token: rand.Text(), At: now, Until: now.Add(d.options.ScanLease)}
		if err := validatePreparedCycle(c, d, state); err != nil {
			return err
		}
		if missing {
			err = tx.Insert(ctx, rec)
		} else {
			err = tx.Set(ctx, rec)
		}
		if err == nil {
			acquired = c
		}
		return err
	})
	return acquired, err
}

func (d *Dispatcher) current(c, expected *scanCycle) error {
	now := d.now()
	if !instant(now) {
		return ErrInvalid
	}
	if c.Epoch != expected.Epoch || !c.Cutoff.Equal(expected.Cutoff) || c.Cursor != expected.Cursor || c.Lease == nil || expected.Lease == nil || c.Lease.Token != expected.Lease.Token || now.Before(c.Lease.At) || !now.Before(c.Lease.Until) {
		return ErrLeaseLost
	}
	return nil
}

// Advance and release are compare-and-swap writes. Failed/unknown commits are
// retried only by rereading durable state, never by resetting the cycle.
func (d *Dispatcher) checkpoint(ctx context.Context, expected *scanCycle, next string, release bool) error {
	if len(next) > 64<<10 || (!release && next != "" && next == expected.Cursor) {
		return ErrInvalid
	}
	err := d.db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		rec, c, err := d.readCycle(ctx, tx, expected.State)
		if err != nil {
			return err
		}
		if err := d.current(c, expected); err != nil {
			return err
		}
		if release {
			c.Lease = nil
		} else {
			c.Cursor = next
			if next == "" {
				c.Cutoff = time.Time{}
				c.Lease = nil
			}
		}
		if err := validatePreparedCycle(c, d, expected.State); err != nil {
			return err
		}
		return tx.Set(ctx, rec)
	})
	if err == nil && !release {
		expected.Cursor = next
	}
	return err
}

func (d *Dispatcher) claim(ctx context.Context, cycle *scanCycle, hint Candidate) (ClaimRef, error) {
	var claim ClaimRef
	err := d.db.RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		_, current, err := d.readCycle(ctx, tx, cycle.State)
		if err != nil {
			return err
		}
		if err := d.current(current, cycle); err != nil {
			return err
		}
		_, w, err := readWork(ctx, tx, hint.Ref)
		if err != nil {
			return err
		}
		if w.Generation != hint.Generation || w.Spec.Handler != d.executor.handler || w.Spec.Ref.Producer != d.executor.producer || w.State != cycle.State {
			return ErrConflict
		}
		eligible := w.NextAttemptAt
		if w.State == Leased {
			eligible = w.Claim.Until
		}
		if eligible.After(cycle.Cutoff) {
			return ErrConflict
		}
		change, value, err := PrepareClaim(ctx, tx, hint.Ref, hint.Generation, d.now(), d.options.JobLease)
		if err != nil {
			return err
		}
		// All reads precede this final clock check and the claim write. Reading the
		// cycle in this same transaction fences concurrent takeover at commit.
		if err := d.current(current, cycle); err != nil {
			return err
		}
		if err := change.Apply(ctx); err != nil {
			return err
		}
		claim = value
		return nil
	})
	return claim, err
}
