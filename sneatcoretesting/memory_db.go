// Package sneatcoretesting contains test-only infrastructure shared by Sneat
// extensions and applications.
package sneatcoretesting

import (
	"context"
	"testing"

	"github.com/dal-go/dalgo/adapters/dalgo2memory"
	"github.com/dal-go/dalgo/dal"
	"github.com/sneat-co/sneat-go-core/facade"
)

// NewMemoryDB creates a strict in-memory database for Sneat tests.
//
// The strict transaction mode mirrors Firestore: a transaction cannot read
// after its first write. Tests should use this instead of constructing a
// dalgo2memory database directly.
func NewMemoryDB() dal.DB {
	return dalgo2memory.NewDB(dalgo2memory.WithNoReadsAfterWritesInTransaction())
}

// ContextWithMemoryDB returns a child context with a new strict in-memory
// database installed as its facade DB.
func ContextWithMemoryDB(parent context.Context) (context.Context, dal.DB) {
	db := NewMemoryDB()
	return facade.WithSneatDB(parent, db), db
}

// SetupMemoryDB creates a strict in-memory database and installs it as the
// facade DB in a new context. It does not mutate application-wide state, so
// tests using it can call t.Parallel safely.
func SetupMemoryDB(t *testing.T) (context.Context, dal.DB) {
	t.Helper()
	return ContextWithMemoryDB(context.Background())
}
