---
format: https://specscore.md/decision-specification
status: Draft
---

# Decision: One generic slug-claim package, namespaced

**Status:** Draft
**Date:** 2026-08-23
**Owner:** alex
**Tags:** slugs,dalgo,spaceus,bookius
**Source Idea:** —
**Supersedes:** —
**Superseded By:** —

## Context

NoticeBoard.cc needs a public URL per centre — `noticeboard.cc/venue/<slug>` —
which needs a globally unique, human-readable slug claimed against a space. Asked
how to build it, the founder proposed a generic shared package rather than a
local fix:

> The implementation should be a generic shared package but allow registering a
> unique slug for each extension separately. Something like
> `/ext/<ext-id>/space-slugs/<space-slug>` → `.spaceID`.

Looking for prior art turned up **two existing hand-rolled implementations of the
same mechanism**, which settles the question of whether a shared package is
warranted:

- **`sneat-core-modules/spaceus`** — `getUniqueSpaceID` derives an ID from the
  title's slug and claims it with a Get-then-take retry loop against the spaces
  collection. It also *strips hyphens*, so "St Mary's Village Hall" becomes
  `stmarysvillagehall`, and resolves a collision by appending `_` and a random
  character.
- **`sneat-co/bookius`** — a dedicated collection `bookiusBookingTypeSlugs`, whose
  own doc comment reads *"the atomic (spaceID, slug) uniqueness index for booking
  types: a composite-key document, never queried, only inserted-or-conflicts
  inside the same transaction as the type itself"*, with a sentinel
  `ErrBookingTypeSlugTaken` surfaced as HTTP 409.

Both are correct. Both are the same idea written twice. A third was about to be
written for space slugs.

## Decision

**One generic slug-claim package in `sneat-go-core`, keyed by an opaque
namespace.**

```
/slugs/<namespace>/claims/<slug>  →  { targetID, targetKind, claimedAt, claimedBy }
```

- The **document ID is the lock.** A claim is taken by an insert that fails if the
  document exists. This is the only shape that is atomic without depending on a
  read.
- The **namespace is an opaque string**, not a fixed "space" concept, so the same
  package serves every uniqueness scope in the fleet:

  | Namespace | Uniqueness |
  |---|---|
  | `communitycentrum:space` | global — the public `/venue/<slug>` case |
  | `bookius:bookingType:<spaceID>` | per space — the existing index |
  | `communitycentrum:room:<spaceID>` | per venue — rooms below the venue URL |

- The **claim and the record are written in one transaction.** A claim without its
  record, or a record without its claim, is a defect the package must make
  impossible for callers to produce.
- **The slug is stored on the record too.** The claim document answers slug →
  targetID for routing; the record answers targetID → slug for rendering canonical
  URLs without a query.
- **Slugs are normalised before claiming** — lowercased, Unicode-normalised,
  hyphens collapsed — so two claims cannot render identically.
- **A reserved-word list** ships with the package and is extensible per namespace.
- **Releasing a slug leaves a tombstone**, not a free slot.

The package lives in **`sneat-go-core`**, not in `spaceus`.

## Rationale

Namespacing by an opaque string rather than by extension is the one change from
the founder's sketch, and it is what lets the package absorb the prior art instead
of sitting beside it. Bookius' scope is *(space, slug)*; rooms will want *(venue,
slug)*; space slugs are global. An extension-only key expresses the third and not
the first two, so Bookius would have kept its own index and the duplication this
decision exists to remove would have survived it. `communitycentrum:space` is
simply the case where the namespace happens to name only an extension.

Extension scoping remains right for the space case specifically, because it
**mirrors the URL namespace exactly**: `noticeboard.cc/venue/…` is already scoped
by domain plus path segment, so one product's slugs cannot collide with another's.

`sneat-go-core` rather than `spaceus` because booking-type slugs have nothing to
do with spaces. Putting a generic claim mechanic behind the space module would
make Bookius import `spaceus` in order to claim a slug for a booking type, which
inverts the dependency for no reason.

The three rules that look like polish are the ones that fail in production:

**Tombstones.** A slug that has been printed on a village-hall poster must never
be handed to a different trust. Releasing into the free pool means last year's
poster silently starts sending people to someone else's hall — a failure that
cannot be detected from inside the system and cannot be retrofitted once posters
exist. Renaming therefore leaves a tombstone that redirects and blocks re-claiming.

**Reserved words.** `/venue/new`, `/venue/admin`, `/venue/api`, `/venue/robots.txt`.
Without a reserved list the first centre called "New" breaks routing, and the fix
after the fact requires taking someone's URL away.

**Normalisation.** `St-Marys` and `st-marys` are different document IDs that render
identically. In a private namespace that is a collision bug; in a public one it is
a phishing vector.

## Declined Alternatives

### Extension-keyed only, per the original sketch

`/ext/<ext-id>/space-slugs/<slug>`. Structurally correct — document-ID-as-lock,
per-extension namespacing, claim in the same transaction — and it is what this
decision keeps. Declined only in its **fixed `space-slugs` leaf**, which cannot
express Bookius' per-space scope or rooms' per-venue scope, so the duplication
would have remained.

### A flat collection with a composite ID

`/slugs/<namespace>:<slug>`. One collection to secure, one rule to write.
Declined because listing every slug in a namespace then needs a separate indexed
field, whereas the nested form is an ordinary subcollection listing — and
enumerating a namespace is exactly what an admin surface and any migration need.

### Leaving each consumer to roll its own

The status quo, and it demonstrably works twice. Declined under the standing rule
that a gap in our own tooling is fixed upstream and consumed, not reimplemented
per consumer. Two correct implementations already differ in their error type,
their collision behaviour and their normalisation; a third would differ again, and
none of them can be fixed once for everyone.

### Putting it in `spaceus`

Closest to where the immediate need arose. Declined because it would force every
extension claiming a slug for a non-space thing to depend on the space module.

## Consequences at Decision Time

**Expected to follow:**

- Bookius' `bookiusBookingTypeSlugs` index, `ErrBookingTypeSlugTaken` and its
  transactional claim collapse into calls to the package, behind its existing HTTP
  409 contract. Its data needs migrating into the new collection.
- `spaceus` gains a **public slug** claimed through the package. See below for what
  deliberately does *not* change.
- The package's atomicity has to be provable in tests. The in-memory dalgo adapter
  serialises every read-write transaction under one global lock, so a contention
  test written against it passes without proving anything — being addressed
  separately in `dal-go/dalgo`.

**Costs accepted:**

- A migration for Bookius' existing claims, which must run before its old index is
  removed, not after.
- Tombstones make the namespace monotonically grow. That is the intended trade:
  storage is cheap, a mis-resolving printed URL is not.
- One more shared package whose bugs are everyone's bugs.

**Deliberately NOT changed:** `spaceus`' `getUniqueSpaceID`. It *generates* an ID
with automatic collision suffixes; the package *claims* a caller-supplied slug and
fails on conflict. These are different contracts, and every Sneat space ID in
existence comes from the former — changing it is a fleet-wide blast radius nobody
asked for. `spaceus` gains a public slug alongside its ID; the ID keeps its
behaviour.

**Deliberately left open:** how long a tombstone lives, and what a second
"St Mary's Village Hall" is offered.

## Observed Consequences

None observed yet.

## Affected Features

- `slug-claims` — the package this decision creates.

---
*This document follows the https://specscore.md/decision-specification*
