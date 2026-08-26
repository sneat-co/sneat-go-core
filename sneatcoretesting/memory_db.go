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

// NewInMemoryTestDB is the fleet's single place that names Firestore as the
// backend an in-memory test database emulates. dal-go/dalgo v0.74.0 made the
// emulated backend an explicit, compile-enforced parameter of
// dalgo2memory.New — dalgo itself is platform-independent, with adapters for
// Firestore, Datastore, MySQL, Postgres, SQLite and more, so no single
// vendor's transaction semantics can be a neutral default. The Sneat fleet's
// production backend IS Firestore, so its tests must exercise Firestore's
// transaction semantics: genuine contention between concurrent read-write
// transactions, snapshot reads, bounded auto-retry, and strict
// read-before-write ordering within a transaction. Call sites elsewhere in
// the fleet should use this helper instead of choosing a dalgo2memory profile
// ad hoc — that keeps the "our production backend is Firestore" decision made
// once, here, rather than repeated (and potentially gotten wrong) at every
// call site in every repo.
func NewInMemoryTestDB(options ...dalgo2memory.Option) dal.DB {
	return dalgo2memory.New(dalgo2memory.FirestoreProfile(), options...)
}

// NewMemoryDB creates a strict in-memory database for Sneat tests.
//
// The strict transaction mode mirrors Firestore: a transaction cannot read
// after its first write. Tests should use this instead of constructing a
// dalgo2memory database directly.
func NewMemoryDB() dal.DB {
	return NewInMemoryTestDB()
}

// NewStrictSchemaMemoryDB creates an empty strict-schema in-memory database
// for tests that must reject accesses to undeclared collections. It retains the
// Firestore-compatible transaction ordering enforced by NewMemoryDB.
func NewStrictSchemaMemoryDB() dal.DB {
	return NewInMemoryTestDB(dalgo2memory.WithSchema(false))
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
