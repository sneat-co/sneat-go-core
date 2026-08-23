package slugs

import "errors"

// ErrSlugTaken, ErrSlugReserved and ErrSlugNotFound are the three sentinel
// errors this package returns. They exist as three distinct values — rather
// than one generic "claim failed" error with a reason code — because
// REQ:reserved-words is explicit that "taken" and "reserved" must be
// distinguishable: the remedy for one is "choose another name" and the
// remedy for the other is "someone already has this", and a caller (e.g.
// Bookius mapping to its own HTTP 409) needs to tell them apart without
// string-matching an error message.
var (
	// ErrSlugTaken means the (namespace, slug) pair is already claimed —
	// either by a live claim or by a tombstone, since a tombstoned slug is
	// deliberately not claimable by a different target
	// (REQ:release-leaves-tombstone). Test with IsSlugTaken.
	ErrSlugTaken = errors.New("slugs: slug is already claimed")

	// ErrSlugReserved means the normalised slug is on the reserved-word list
	// (the base list, or one extended per namespace via WithReservedWords)
	// and was never attempted against storage. Test with IsSlugReserved.
	ErrSlugReserved = errors.New("slugs: slug is reserved")

	// ErrSlugNotFound means no claim, live or tombstoned, exists at
	// (namespace, slug). Test with IsSlugNotFound.
	ErrSlugNotFound = errors.New("slugs: slug is not claimed")
)

// IsSlugTaken reports whether err (or anything it wraps) is ErrSlugTaken.
func IsSlugTaken(err error) bool { return errors.Is(err, ErrSlugTaken) }

// IsSlugReserved reports whether err (or anything it wraps) is
// ErrSlugReserved.
func IsSlugReserved(err error) bool { return errors.Is(err, ErrSlugReserved) }

// IsSlugNotFound reports whether err (or anything it wraps) is
// ErrSlugNotFound.
func IsSlugNotFound(err error) bool { return errors.Is(err, ErrSlugNotFound) }
