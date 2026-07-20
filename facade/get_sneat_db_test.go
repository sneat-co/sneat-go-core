package facade

import (
	"context"
	"errors"
	"testing"

	"github.com/dal-go/dalgo/adapters/dalgo2memory"
	"github.com/dal-go/dalgo/dal"
)

func TestGetDatabase(t *testing.T) {
	ctx := context.Background()
	defer mustPanicWithErrNotInitialized(t)
	_, _ = GetSneatDB(ctx)
	//if db != nil {
	//	t.Error("GetSneatDB() = ", db)
	//}
	//if !errors.Is(err, ErrNotInitialized) {
	//	t.Error("Expected to return ErrNotInitialized")
	//}
}

func TestWithSneatDB_IsolatesContexts(t *testing.T) {
	t.Parallel()
	db1 := dalgo2memory.NewDB()
	db2 := dalgo2memory.NewDB()
	ctx1 := WithSneatDB(context.Background(), db1)
	ctx2 := WithSneatDB(context.Background(), db2)

	got1, err := GetSneatDB(ctx1)
	if err != nil {
		t.Fatal(err)
	}
	got2, err := GetSneatDB(ctx2)
	if err != nil {
		t.Fatal(err)
	}
	if got1 != db1 {
		t.Error("first context did not resolve its database")
	}
	if got2 != db2 {
		t.Error("second context did not resolve its database")
	}
}

func TestWithSneatDBProvider_ReturnsProviderError(t *testing.T) {
	t.Parallel()
	wantErr := errors.New("provider failed")
	ctx := WithSneatDBProvider(context.Background(), func(context.Context) (dal.DB, error) {
		return nil, wantErr
	})

	_, err := GetSneatDB(ctx)
	if !errors.Is(err, wantErr) {
		t.Errorf("GetSneatDB() error = %v, want %v", err, wantErr)
	}
}

func TestDefaultSneatDBProviderConfiguration(t *testing.T) {
	defaultSneatDBProviderMu.RLock()
	previous := defaultSneatDBProvider
	defaultSneatDBProviderMu.RUnlock()
	t.Cleanup(func() { SetDefaultSneatDBProvider(previous) })

	db := dalgo2memory.NewDB()
	SetDefaultSneatDBProvider(func(context.Context) (dal.DB, error) {
		return db, nil
	})
	decorated := false
	UpdateDefaultSneatDBProvider(func(provider SneatDBProvider) SneatDBProvider {
		return func(ctx context.Context) (dal.DB, error) {
			decorated = true
			return provider(ctx)
		}
	})

	got, err := GetSneatDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got != db {
		t.Error("configured default provider did not resolve its database")
	}
	if !decorated {
		t.Error("updated default provider did not invoke its decorator")
	}
}

func mustPanicWithErrNotInitialized(t *testing.T) {
	if r := recover(); r == nil {
		t.Error("Expected to panic")
	} else if err := r.(error); !errors.Is(err, ErrNotInitialized) {
		t.Error("Expected to return ErrNotInitialized")
	}
}
