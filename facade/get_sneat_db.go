package facade

import (
	"context"
	"fmt"
	"sync"

	"github.com/dal-go/dalgo/dal"
)

// SneatDBProvider resolves the database to use for a context.
type SneatDBProvider func(ctx context.Context) (dal.DB, error)

type sneatDBContextKey struct{}

var (
	defaultSneatDBProviderMu sync.RWMutex
	defaultSneatDBProvider   SneatDBProvider = uninitializedSneatDBProvider
	contextDBKey                             = sneatDBContextKey{}
)

// GetSneatDB returns a context override when one is present, otherwise it uses
// the default provider configured by the application at startup.
func GetSneatDB(ctx context.Context) (dal.DB, error) {
	if provider, ok := ctx.Value(contextDBKey).(SneatDBProvider); ok {
		return provider(ctx)
	}

	defaultSneatDBProviderMu.RLock()
	provider := defaultSneatDBProvider
	defaultSneatDBProviderMu.RUnlock()
	return provider(ctx)
}

// WithSneatDB returns a child context that resolves db through GetSneatDB.
// This is useful for request-scoped database selection and parallel tests.
func WithSneatDB(ctx context.Context, db dal.DB) context.Context {
	if db == nil {
		panic("facade.WithSneatDB: nil DB")
	}
	return WithSneatDBProvider(ctx, func(context.Context) (dal.DB, error) {
		return db, nil
	})
}

// WithSneatDBProvider returns a child context that resolves its database using
// provider. Unlike changing the default provider, this is safe in parallel
// tests and allows testing provider failures.
func WithSneatDBProvider(ctx context.Context, provider SneatDBProvider) context.Context {
	if ctx == nil {
		panic("facade.WithSneatDBProvider: nil context")
	}
	if provider == nil {
		panic("facade.WithSneatDBProvider: nil provider")
	}
	return context.WithValue(ctx, contextDBKey, provider)
}

// SetDefaultSneatDBProvider replaces the application-wide fallback provider.
// Applications should call this only while wiring dependencies at startup;
// tests should use WithSneatDB instead.
func SetDefaultSneatDBProvider(provider SneatDBProvider) {
	if provider == nil {
		panic("facade.SetDefaultSneatDBProvider: nil provider")
	}
	defaultSneatDBProviderMu.Lock()
	defaultSneatDBProvider = provider
	defaultSneatDBProviderMu.Unlock()
}

// UpdateDefaultSneatDBProvider atomically decorates or replaces the current
// application-wide provider. It is intended for startup wiring such as adding
// a cache around the primary database provider.
func UpdateDefaultSneatDBProvider(update func(SneatDBProvider) SneatDBProvider) {
	if update == nil {
		panic("facade.UpdateDefaultSneatDBProvider: nil update")
	}
	defaultSneatDBProviderMu.Lock()
	defer defaultSneatDBProviderMu.Unlock()
	provider := update(defaultSneatDBProvider)
	if provider == nil {
		panic("facade.UpdateDefaultSneatDBProvider: update returned nil provider")
	}
	defaultSneatDBProvider = provider
}

func uninitializedSneatDBProvider(context.Context) (dal.DB, error) {
	err := fmt.Errorf("%w: facade.GetSneatDB(context.Context) (dal.DB, error)", ErrNotInitialized)
	panic(err)
}
