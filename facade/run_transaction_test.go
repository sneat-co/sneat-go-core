package facade

import (
	"context"
	"errors"
	"github.com/dal-go/dalgo/dal"
	"github.com/dal-go/mocks4dalgo/mock_dal"
	"github.com/sneat-co/sneat-go-core/consts4dal"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
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

	mockGetSneatDB := func() {
		GetSneatDB = func(_ context.Context) (dal.DB, error) {
			ctrl := gomock.NewController(t)
			mockDB := mock_dal.NewMockDB(ctrl)
			mockDB.EXPECT().RunReadwriteTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, f dal.RWTxWorker, options ...dal.TransactionOption) error {
				err := f(ctx, mock_dal.NewMockReadwriteTransaction(ctrl))
				if deadline, ok := ctx.Deadline(); ok && time.Until(deadline) <= 0 {
					return context.DeadlineExceeded
				}
				return err
			})
			return mockDB, nil
		}
	}

	t.Run("no_error", func(t *testing.T) {
		mockGetSneatDB()
		ctx := context.Background()
		err := RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			return nil
		})
		if err != nil {
			t.Errorf("RunReadwriteTransaction() returned unexpected error: %v", err)
		}
	})

	t.Run("returns_expected_error", func(t *testing.T) {
		mockGetSneatDB()
		ctx := context.Background()
		expectedErr := errors.New("expected error")
		err := RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
			return expectedErr
		})
		if !errors.Is(err, expectedErr) {
			t.Errorf("RunReadwriteTransaction() returned unexpected error: %v", err)
		}
	})

	t.Run("deadline", func(t *testing.T) {
		mockGetSneatDB()
		const deadline = time.Millisecond
		withDefaultDeadLineCalled := 0
		consts4dal.WithDefaultDeadLine = func(ctx context.Context) (context.Context, context.CancelFunc) {
			withDefaultDeadLineCalled++
			return context.WithDeadline(ctx, time.Now().Add(deadline))
		}

		ctx := context.Background()
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
