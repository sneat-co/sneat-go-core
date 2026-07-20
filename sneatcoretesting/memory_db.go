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

// SetupMemoryDB creates a strict in-memory database and installs it as the
// facade DB for the lifetime of t. The previous getter is restored at cleanup.
func SetupMemoryDB(t *testing.T) dal.DB {
	t.Helper()
	db := NewMemoryDB()
	previous := facade.GetSneatDB
	facade.GetSneatDB = func(context.Context) (dal.DB, error) { return db, nil }
	t.Cleanup(func() { facade.GetSneatDB = previous })
	return db
}
