package facade

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/dalgo/mocks/mock_dal"
	"github.com/sneat-co/sneat-go-core/consts4dal"
	"go.uber.org/mock/gomock"
)

func TestRunReadwriteTransaction(t *testing.T) {

	t.Run("panics", func(t *testing.T) {
		defer mustPanicWithErrNotInitialized(t)
		ctx := context.Background()
		err := RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			return nil
		})
		if err != nil {
			t.Errorf("RunReadwriteTransaction() returned unexpected error: %v", err)
		}
	})

	contextWithMockDB := func(t *testing.T) context.Context {
		ctrl := gomock.NewController(t)
		mockDB := mock_dal.NewMockDB(ctrl)
		mockDB.EXPECT().RunReadwriteTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, f dal.RWTxWorker, options ...dal.TransactionOption) error {
			err := f(ctx, mock_dal.NewMockReadwriteTransaction(ctrl))
			if deadline, ok := ctx.Deadline(); ok && time.Until(deadline) <= 0 {
				return context.DeadlineExceeded
			}
			return err
		})
		return WithSneatDB(context.Background(), mockDB)
	}

	t.Run("no_error", func(t *testing.T) {
		t.Parallel()
		ctx := contextWithMockDB(t)
		err := RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			return nil
		})
		if err != nil {
			t.Errorf("RunReadwriteTransaction() returned unexpected error: %v", err)
		}
	})

	t.Run("returns_expected_error", func(t *testing.T) {
		t.Parallel()
		ctx := contextWithMockDB(t)
		expectedErr := errors.New("expected error")
		err := RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			return expectedErr
		})
		if !errors.Is(err, expectedErr) {
			t.Errorf("RunReadwriteTransaction() returned unexpected error: %v", err)
		}
	})

	t.Run("deadline", func(t *testing.T) {
		ctx := contextWithMockDB(t)
		const deadline = time.Millisecond
		withDefaultDeadLineCalled := 0
		previousWithDefaultDeadLine := consts4dal.WithDefaultDeadLine
		t.Cleanup(func() { consts4dal.WithDefaultDeadLine = previousWithDefaultDeadLine })
		consts4dal.WithDefaultDeadLine = func(ctx context.Context) (context.Context, context.CancelFunc) {
			withDefaultDeadLineCalled++
			return context.WithDeadline(ctx, time.Now().Add(deadline))
		}

		err := RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			time.Sleep(deadline + time.Millisecond)
			return nil
		})
		if err == nil {
			t.Error("RunReadwriteTransaction() did not return expected error")
		} else if !errors.Is(err, context.DeadlineExceeded) {
			t.Logf("RunReadwriteTransaction() returned unexpected error: %v", err)
		}
		if withDefaultDeadLineCalled != 1 {
			t.Errorf("WithDefaultDeadLine() called %d times, expected 1", withDefaultDeadLineCalled)
		}
	})
}
