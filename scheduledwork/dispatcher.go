package scheduledwork

import (
	"context"
	"errors"
	"time"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/record"
)

// DispatcherOptions is trusted server configuration. PageSize is persisted as
// part of cursor identity; changing it requires an explicit reviewed reset.
type DispatcherOptions struct {
	PageSize       int
	PagesPerState  int
	JobLease       time.Duration
	ScanLease      time.Duration
	StateTimeout   time.Duration
	EnqueueTimeout time.Duration
	Now            func() time.Time
}

// Dispatcher executes ONE registered producer/handler. It does not register a
// cron, authenticate a host, backfill jobs or change any business/financial rule.
type Dispatcher struct {
	db       dal.DB
	executor *Executor
	options  DispatcherOptions
}

func NewDispatcher(db dal.DB, executor *Executor, options DispatcherOptions) (*Dispatcher, error) {
	validDuration := func(v time.Duration) bool { return v > 0 && v%time.Microsecond == 0 }
	if db == nil || executor == nil || executor.delayer == nil || !identifier(string(executor.producer)) || !identifier(executor.handler) || !identifier(executor.queue) || options.Now == nil || options.PageSize < 1 || options.PageSize > 100 || options.PagesPerState < 1 || options.PagesPerState > 4 || !validDuration(options.JobLease) || options.JobLease > time.Hour || !validDuration(options.ScanLease) || options.ScanLease > 10*time.Minute || !validDuration(options.StateTimeout) || options.StateTimeout > 2*time.Minute || !validDuration(options.EnqueueTimeout) || options.EnqueueTimeout > 30*time.Second || options.EnqueueTimeout > options.StateTimeout || options.ScanLease < options.StateTimeout+options.EnqueueTimeout {
		return nil, ErrInvalid
	}
	return &Dispatcher{db: db, executor: executor, options: options}, nil
}

// DispatchReport contains counts, never cursor contents, claim tokens, raw
// record keys or product data. Enqueued counts acknowledgments, NOT outcomes.
type DispatchReport struct {
	Pages, Candidates, Rejected, Claimed, Enqueued, Unconfirmed, Stale, Busy int
}

// DispatchError deliberately has a fixed safe message. Its cause remains
// available for errors.Is/As; do not serialize unwrapped storage/provider errors.
type DispatchError struct{ cause error }

func (e *DispatchError) Error() string { return "scheduled work dispatch incomplete" }
func (e *DispatchError) Unwrap() error { return e.cause }

// Tick bounds both pages AND elapsed time. Pending work and expired leases get
// separate time allowances; a slow pending query cannot consume lease recovery's
// allowance. Adapters/delayers must honor context deadlines. No goroutine is
// detached to forcefully interrupt an adapter that violates that contract.
func (d *Dispatcher) Tick(ctx context.Context) (DispatchReport, error) {
	var report DispatchReport
	if d == nil || ctx == nil {
		return report, ErrInvalid
	}
	if err := ctx.Err(); err != nil {
		return report, &DispatchError{cause: err}
	}
	cleanupBudget := min(d.options.EnqueueTimeout, 5*time.Second)
	ctx, cancel := context.WithTimeout(ctx, 2*(d.options.StateTimeout+cleanupBudget))
	defer cancel()
	var failures []error
	for _, state := range []State{Pending, Leased} {
		if err := d.runState(ctx, state, cleanupBudget, &report); err != nil {
			failures = append(failures, err)
		}
	}
	if len(failures) > 0 {
		return report, &DispatchError{cause: errors.Join(failures...)}
	}
	return report, nil
}

func (d *Dispatcher) runState(parent context.Context, state State, cleanupBudget time.Duration, report *DispatchReport) error {
	if err := parent.Err(); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(parent, d.options.StateTimeout)
	defer cancel()
	cycle, err := d.acquire(ctx, state)
	if errors.Is(err, errCycleBusy) {
		report.Busy++
		return nil
	}
	if err != nil {
		return err
	}
	err = d.pages(ctx, cycle, report)
	if cycle.Cursor == "" && cycle.Cutoff.IsZero() {
		return err
	}
	// A state timeout gets a bounded cleanup attempt without consuming the other
	// state's allowance. Parent cancellation is still propagated to cleanup.
	cleanup, cancelCleanup := context.WithTimeout(parent, cleanupBudget)
	defer cancelCleanup()
	releaseErr := d.checkpoint(cleanup, cycle, "", true)
	return errors.Join(err, releaseErr)
}

func (d *Dispatcher) pages(ctx context.Context, cycle *scanCycle, report *DispatchReport) error {
	var uncertain error
	for i := 0; i < d.options.PagesPerState; i++ {
		if err := ctx.Err(); err != nil {
			return errors.Join(uncertain, err)
		}
		page, err := DiscoverPage(ctx, d.db, Discovery{Producer: cycle.Producer, Handler: cycle.Handler, State: cycle.State, Now: cycle.Cutoff, Limit: cycle.PageSize}, cycle.Cursor)
		if err != nil {
			return errors.Join(uncertain, err)
		}
		report.Pages++
		report.Rejected += len(page.Rejected)
		report.Candidates += len(page.Ready)
		for _, hint := range page.Ready {
			if err := ctx.Err(); err != nil {
				return errors.Join(uncertain, err)
			}
			claim, err := d.claim(ctx, cycle, hint)
			if errors.Is(err, ErrConflict) || record.IsNotFound(err) {
				report.Stale++
				continue
			}
			// Losing the cycle lease must stop the page; a later tick rereads it.
			// A live competing job lease is also safe to rediscover, without advancing.
			if err != nil {
				return errors.Join(uncertain, err)
			}
			report.Claimed++
			enqueueCtx, cancel := context.WithTimeout(ctx, d.options.EnqueueTimeout)
			err = d.executor.Enqueue(enqueueCtx, claim)
			cancel()
			if err != nil {
				report.Unconfirmed++
				uncertain = errors.Join(uncertain, err)
			} else {
				report.Enqueued++
			}
		}
		if err := d.checkpoint(ctx, cycle, page.Next, false); err != nil {
			return errors.Join(uncertain, err)
		}
		if page.Next == "" {
			cycle.Cutoff = time.Time{}
			return uncertain
		}
	}
	return uncertain
}
