// Package slugs claims human-readable, globally-scoped-per-namespace slugs
// exactly once, without a caller ever having to write the concurrency logic
// itself.
//
// The same mechanism was hand-written twice in this fleet before this package
// existed — bookius' bookingType slugs and spaceus' space IDs — differently
// each time, and a third was about to be written for NoticeBoard's public
// venue URLs. See spec/decisions/0001-slug-claims.md for why a shared,
// namespaced package replaces all three rather than sitting beside them, and
// spec/features/slug-claims/README.md for the requirements this package is
// implementing.
//
// # The document ID is the lock
//
// A claim lives at /slugs/<namespace>/claims/<slug>, and the slug IS the
// document's ID. That is what makes Claim atomic without a preceding read:
// two callers racing to claim the same slug are really racing to insert the
// same document, and a storage backend that rejects a duplicate insert
// settles the race for us. This package never reads before it writes to
// decide whether a slug is free — a read-then-write is only safe under a
// backend that itself detects the conflict, and the whole point of this
// package is to be correct without assuming that.
//
// # Namespaces are opaque
//
// A namespace is any caller-supplied string. It is not a fixed concept like
// "space" or "extension" because the fleet needs three different uniqueness
// scopes from one mechanism: global (a community centre's public slug),
// per-parent (a booking type's slug, unique within its space), and per-parent
// again but one level down (a room's slug, unique within its venue). All
// three are just "some string" to this package; the scoping rule lives
// entirely in what string the caller passes as the namespace.
//
// # Claim joins the caller's transaction
//
// Claim, Release and Rename all take a caller-supplied dal.ReadwriteTransaction
// rather than opening one of their own. That is deliberate: a claim written
// outside the transaction that creates the record it names could commit while
// the record's write fails (or vice versa), leaving an orphan claim or an
// unfindable record. There is no API in this package whose natural use writes
// the claim separately from the record it names.
//
// # Release leaves a tombstone
//
// Nothing in this package deletes a claim document. A slug that has been
// printed on a poster and then handed to a different target sends real people
// to the wrong place, silently, forever — and that failure cannot be
// retrofitted once posters exist. So Release and Rename both leave the
// document in place and mark it as released, which (a) means a released slug
// can never be claimed again by someone else, and (b) lets Resolve report a
// successor slug for a caller to redirect to.
package slugs
