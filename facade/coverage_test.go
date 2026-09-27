package facade

import (
	"context"
	"errors"
	"testing"

	"github.com/dal-go/dalgo/dal"
	"github.com/strongo/analytics"
)

type dummyAnalyticsMsg struct {
	analytics.Message
}

func (dummyAnalyticsMsg) Event() string {
	return "test_event"
}

func (dummyAnalyticsMsg) Category() string {
	return "test_cat"
}

func TestDocumentSnapshot(t *testing.T) {
	ds := NewDocumentSnapshot(true, func(p interface{}) error {
		return nil
	})
	if !ds.Exists() {
		t.Fatal("expected exists to be true")
	}
	if err := ds.DataTo(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserAnalytics(t *testing.T) {
	called := false
	ua := NewUserAnalytics(func(msg analytics.Message) {
		called = true
	})
	ua.Send(dummyAnalyticsMsg{})
	if !called {
		t.Fatal("expected Send to be called")
	}
}

func TestUserContext_EmptyIDPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on empty user ID")
		}
	}()
	NewUserContext("")
}

func TestContextWithUser_Analytics(t *testing.T) {
	ctx := context.Background()
	userCtx := NewUserContext("u1")

	// No analytics
	ctxNoUA := NewContextWithUser(ctx, userCtx)
	na := ctxNoUA.Analytics()
	na.Send(dummyAnalyticsMsg{})

	// With analytics
	called := false
	ua := NewUserAnalytics(func(msg analytics.Message) { called = true })
	ctxWithUA := NewContextWithUserAndAnalytics(ctx, userCtx, ua)
	ctxWithUA.Analytics().Send(dummyAnalyticsMsg{})
	if !called {
		t.Fatal("expected ua to be called")
	}
}

func TestRunReadwriteTransaction_GetSneatDBFails(t *testing.T) {
	ctx := WithSneatDBProvider(context.Background(), func(context.Context) (dal.DB, error) {
		return nil, errors.New("db provider failed")
	})
	err := RunReadwriteTransaction(ctx, func(ctx context.Context, tx dal.ReadwriteTransaction) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error when GetSneatDB fails, got nil")
	}
}

func TestGetSneatDB_Panics(t *testing.T) {
	ctx := context.Background()

	t.Run("WithSneatDB_nil_db", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		WithSneatDB(ctx, nil)
	})

	t.Run("WithSneatDBProvider_nil_ctx", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		WithSneatDBProvider(nil, func(context.Context) (dal.DB, error) { return nil, nil })
	})

	t.Run("WithSneatDBProvider_nil_provider", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		WithSneatDBProvider(ctx, nil)
	})

	t.Run("SetDefaultSneatDBProvider_nil", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		SetDefaultSneatDBProvider(nil)
	})

	t.Run("UpdateDefaultSneatDBProvider_nil_update", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		UpdateDefaultSneatDBProvider(nil)
	})

	t.Run("UpdateDefaultSneatDBProvider_update_returns_nil", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		UpdateDefaultSneatDBProvider(func(SneatDBProvider) SneatDBProvider {
			return nil
		})
	})
}
