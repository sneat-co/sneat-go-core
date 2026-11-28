// Package sneatmemdb provides Sneat's in-memory dal.DB configured with Sneat's
// production backend semantics (Firestore: optimistic concurrency, snapshot
// reads, bounded auto-retry, strict read-before-write transaction ordering, and
// microsecond timestamp precision).
//
// This package is the Sneat fleet's single door to dalgo's in-memory adapter
// (github.com/dal-go/dalgo/adapters/dalgo2memory). It has no testing or facade
// dependency, so it can be used both at runtime (dev sandboxes, local dev
// daemons) and by tests (typically through sneatcoretesting).
package sneatmemdb

import (
	"github.com/dal-go/dalgo/adapters/dalgo2memory"
	"github.com/dal-go/dalgo/branching"
	"github.com/dal-go/dalgo/dal"
)

// Option configures an in-memory Sneat database created by New.
type Option = dalgo2memory.Option

// Profile selects the backend semantics emulated by an in-memory database.
type Profile = dalgo2memory.Profile

// CollectionDef describes a collection registered in an in-memory database
// schema via WithCollection and passed to WithSchema.
type CollectionDef = dalgo2memory.CollectionDef

// CollectionOption configures a single collection definition created by
// WithCollection.
type CollectionOption = dalgo2memory.CollectionOption

var (
	// ErrReadAfterWriteInTransaction is returned when a transaction attempts a
	// read after staging a write, matching Firestore's transaction ordering.
	ErrReadAfterWriteInTransaction = dalgo2memory.ErrReadAfterWriteInTransaction

	// ErrTransactionConflict is returned when an optimistic read-write
	// transaction conflicts with a concurrent commit.
	ErrTransactionConflict = dalgo2memory.ErrTransactionConflict

	// ErrNestedTransaction is returned when a transaction starts another
	// transaction using its own transaction context.
	ErrNestedTransaction = dalgo2memory.ErrNestedTransaction
)

// SneatProfile returns the in-memory database profile matching Sneat's
// production backend: Firestore (transaction semantics and microsecond time
// precision).
func SneatProfile() Profile {
	return dalgo2memory.FirestoreProfile()
}

// WithSneatProfile is SneatProfile as an Option: it resets a database built
// with another profile to Sneat's Firestore-compatible backend bundle.
func WithSneatProfile() Option {
	return dalgo2memory.WithFirestoreProfile()
}

// New creates a new in-memory Sneat database configured with SneatProfile
// (Firestore transaction semantics and microsecond timestamp precision), plus
// any additional options.
func New(options ...Option) dal.DB {
	return dalgo2memory.New(SneatProfile(), options...)
}

// WithSchema configures collection schema enforcement and registered
// collection definitions on an in-memory Sneat database.
func WithSchema(allowUndefinedCollections bool, collections ...CollectionDef) Option {
	return dalgo2memory.WithSchema(allowUndefinedCollections, collections...)
}

// WithCollection defines a typed collection for use with WithSchema.
func WithCollection[T any](name string, newRecord func() *T, opts ...CollectionOption) CollectionDef {
	return dalgo2memory.WithCollection[T](name, newRecord, opts...)
}

// NewBranchingProvider returns the branching provider for serialized in-memory
// Sneat databases.
func NewBranchingProvider() branching.Provider {
	return dalgo2memory.NewBranchingProvider()
}

// IsTransactionConflict reports whether err represents an optimistic
// transaction conflict.
func IsTransactionConflict(err error) bool {
	return dalgo2memory.IsTransactionConflict(err)
}
