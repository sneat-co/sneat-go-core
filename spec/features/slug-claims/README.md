---
format: https://specscore.md/feature-specification
status: Draft
---

# Feature: Slug claims

> [SpecScore.**Studio**](https://specscore.studio): | [Explore](https://specscore.studio/app/github.com/sneat-co/sneat-go-core/spec/features/slug-claims?op=explore) | [Edit](https://specscore.studio/app/github.com/sneat-co/sneat-go-core/spec/features/slug-claims?op=edit) | [Ask question](https://specscore.studio/app/github.com/sneat-co/sneat-go-core/spec/features/slug-claims?op=ask) | [Request change](https://specscore.studio/app/github.com/sneat-co/sneat-go-core/spec/features/slug-claims?op=request-change) |
**Status:** Draft
**Source Ideas:** —

## Summary

A shared package for claiming human-readable slugs uniquely, whatever they name.
A claim is a document whose **ID is the slug**, under an opaque **namespace**:

```
/slugs/<namespace>/claims/<slug>  →  { targetID, targetKind, claimedAt, claimedBy }
```

The document ID is the lock, so claiming is atomic without depending on a read.
Callers claim inside **their own transaction**, alongside the record the slug
names, so a claim and its record are never written apart. Slugs are normalised
and checked against reserved words before they are claimed, and releasing one
leaves a **tombstone** rather than a free slot.

Per [Decision 0001](../../decisions/0001-slug-claims.md).

## Problem

The same mechanism has already been written twice in this fleet, differently each
time. `spaceus` claims space IDs with a Get-then-take retry loop that strips
hyphens and appends `_` plus a random character on collision. `bookius` keeps a
dedicated `bookiusBookingTypeSlugs` collection, claimed by insert inside the
caller's transaction, with its own sentinel error. Both work; neither can be fixed
for the other; a third was about to be written for NoticeBoard's public venue URLs.

The parts that get written twice are the easy parts. The parts that get skipped are
the ones that fail later, in public:

- A slug that has been **printed on a poster** and then released to a different
  owner sends real people to the wrong place, and nothing inside the system can
  detect it.
- An unreserved namespace lets the first thing called **"new"** or **"admin"**
  break routing, and the fix takes someone's URL away after they have used it.
- Unnormalised slugs let `St-Marys` and `st-marys` both be claimed while rendering
  identically — a collision bug in private, a phishing vector in public.

## Behavior

### The document ID is the lock

#### REQ: claim-is-atomic

Claiming a slug MUST be performed by an **insert that fails when the document
already exists**, at `/slugs/<namespace>/claims/<slug>`. The package MUST NOT rely
on a preceding read to establish uniqueness: a read-then-write is only safe under a
backend that detects the conflict, and the package must be correct without assuming
it. A failed claim MUST return an error identifiable by an exported predicate
(`IsSlugTaken` or equivalent), so callers can map it to their own contract — for
example Bookius' HTTP 409.

#### REQ: claim-joins-the-caller-transaction

The claim MUST be written inside a transaction **supplied by the caller**, so the
claim and the record the slug names commit together or not at all. The package MUST
NOT open its own transaction for a claim, and MUST NOT offer an API shape whose
natural use writes the claim separately from the record.

### Namespaces are opaque

#### REQ: namespace-is-opaque

The namespace MUST be an arbitrary caller-supplied string, not a fixed concept, so
one package serves a global scope (`communitycentrum:space`), a per-space scope
(`bookius:bookingType:<spaceID>`) and a per-parent scope
(`communitycentrum:room:<spaceID>`) identically. The package MUST validate that a
namespace is non-empty and safe as a document ID, and MUST be able to **enumerate
every claim in a namespace**, which admin surfaces and migrations both need.

### Slugs are normalised before they are claimed

#### REQ: normalisation

A slug MUST be normalised before it is claimed or resolved: Unicode-normalised,
lowercased, with runs of separators collapsed and leading/trailing separators
removed. Two inputs that normalise to the same string MUST resolve to the same
claim rather than to two claims. The normalised form MUST be what is stored and
what appears in URLs, and the package MUST expose the normaliser so callers can
show a user the slug they will actually get before they commit to it.

#### REQ: reserved-words

The package MUST carry a base list of reserved slugs that cannot be claimed in any
namespace — at least `new`, `edit`, `admin`, `api`, `static`, `assets`, `robots`,
`sitemap`, `favicon`, `well-known` — and MUST let a caller extend it per namespace.
An attempt to claim a reserved slug MUST fail with an error distinguishable from
*taken*, because the remedy differs: one is "choose another name", the other is
"someone already has this".

### Resolving

#### REQ: resolve

The package MUST resolve a `(namespace, slug)` to its `targetID` and `targetKind`
in a single read, with no query. Resolving a slug that has never been claimed and
resolving one that is **tombstoned** MUST be distinguishable, since the first is a
404 and the second is a redirect.

### Release leaves a tombstone

#### REQ: release-leaves-tombstone

Releasing a claim — on rename, or when the target is deleted — MUST leave a
**tombstone** at the same document, recording what it previously pointed to and
when it was released, rather than deleting the document. A tombstoned slug MUST NOT
be claimable by a different target. Renaming MUST record the **new** slug on the
tombstone so callers can issue a redirect. The package MUST provide rename as a
single operation that claims the new slug and tombstones the old one in the
caller's transaction, so a rename cannot half-happen.

## Acceptance Criteria

### AC: second-claim-is-refused (verifies REQ:claim-is-atomic)

**Scenario:** a slug is claimed once
**Given** an empty namespace
**When** two claims of the same normalised slug are attempted
**Then** the first succeeds and the second fails with an error for which the exported *taken* predicate returns true.

### AC: concurrent-claims-yield-one-winner (verifies REQ:claim-is-atomic)

**Scenario:** two registrations race for the same name
**Given** a backend that surfaces write conflicts
**When** two transactions concurrently claim the same slug in the same namespace
**Then** exactly one commits and the other fails as *taken* or as a retryable conflict, and there is no state in which both hold the claim.

### AC: claim-and-record-commit-together (verifies REQ:claim-joins-the-caller-transaction)

**Scenario:** no orphan claims
**Given** a caller transaction that claims a slug and then fails before commit
**When** the transaction is rolled back
**Then** no claim document exists — the slug is free.

### AC: namespaces-are-independent (verifies REQ:namespace-is-opaque)

**Scenario:** the same slug in two namespaces
**Given** namespaces `a:space` and `b:space`
**When** the slug `main-hall` is claimed in each
**Then** both succeed and each resolves to its own target.

### AC: namespace-can-be-enumerated (verifies REQ:namespace-is-opaque)

**Scenario:** an admin lists what is claimed
**Given** a namespace holding several claims and a tombstone
**When** the namespace is enumerated
**Then** every claim is returned, with tombstones distinguishable from live claims.

### AC: equivalent-inputs-normalise-together (verifies REQ:normalisation)

**Scenario:** case and separators do not create a second claim
**Given** the slug `St--Marys Village Hall ` claimed in a namespace
**When** `st-marys-village-hall` is claimed in the same namespace
**Then** it is refused as taken, and both inputs resolve to the same target.

### AC: reserved-slug-is-refused-distinctly (verifies REQ:reserved-words)

**Scenario:** a centre called "New"
**Given** an empty namespace
**When** the slug `new` is claimed
**Then** it is refused with an error distinguishable from *taken*; and a namespace-specific extra reserved word behaves the same way.

### AC: resolve-returns-target-in-one-read (verifies REQ:resolve)

**Scenario:** routing a public URL
**Given** a claimed slug
**When** it is resolved
**Then** its `targetID` and `targetKind` are returned from a single document read with no query.

### AC: unclaimed-and-tombstoned-are-distinguishable (verifies REQ:resolve)

**Scenario:** 404 versus redirect
**Given** one slug never claimed and one tombstoned by a rename
**When** each is resolved
**Then** the first reports *not found* and the second reports a tombstone carrying the new slug.

### AC: renamed-slug-is-not-reusable (verifies REQ:release-leaves-tombstone)

**Scenario:** last year's poster keeps working, and cannot be hijacked
**Given** a target renamed from `st-marys-village-hall` to `st-marys-hall`
**When** a different target attempts to claim `st-marys-village-hall`
**Then** it is refused, and resolving the old slug reports a tombstone naming the new one.

### AC: rename-is-atomic (verifies REQ:release-leaves-tombstone)

**Scenario:** a rename cannot half-happen
**Given** a rename whose caller transaction fails before commit
**When** the transaction is rolled back
**Then** the old slug still resolves to the target and the new slug is unclaimed.

## Architecture & Dependencies

- **Home** — `sneat-go-core`, not `spaceus`: booking-type slugs have nothing to do
  with spaces, and a generic mechanic behind the space module would make Bookius
  depend on `spaceus` to claim a slug.
- **Storage** — dalgo, `/slugs/<namespace>/claims/<slug>`; the claim is written with
  an insert on a caller-supplied `dal.ReadwriteTransaction`.
- **Prior art being consolidated** — `bookius`' `bookiusBookingTypeSlugs` +
  `ErrBookingTypeSlugTaken` (a faithful instance of this pattern, to be migrated),
  and `spaceus`' `getUniqueSpaceID` (a *different* contract — ID generation with
  automatic collision suffixes — which is deliberately **not** changed; `spaceus`
  gains a public slug alongside its ID).
- **Testing** — `concurrent-claims-yield-one-winner`, `claim-and-record-commit-together`
  and `rename-is-atomic` are proved against `dalgo2memory`'s opt-in
  `WithOptimisticConcurrency()` mode (`dal-go/dalgo` v0.65.0), in which read-write
  transactions genuinely interleave, buffer their reads and writes locally, and
  fail at commit with `ErrTransactionConflict` if another transaction committed a
  conflicting write in the meantime — giving both real contention and real
  rollback, not a single global lock a test would pass against regardless.

## Not Doing

- **Generating slugs.** The package claims a slug the caller supplies and refuses a
  taken one; it does not invent alternatives or append suffixes.
- **Changing `spaceus`' space-ID generation** — different contract, fleet-wide blast
  radius, nobody asked.
- **Serving redirects.** The package reports a tombstone and its successor; issuing
  the 301 is the caller's routing concern.
- **Expiring tombstones automatically** — see Open Questions.
- **Confusable/homoglyph detection** beyond Unicode normalisation.

## Open Questions

- **How long does a tombstone live?** Forever is safest and grows the namespace
  monotonically; anything shorter needs a defensible number, and the risk is
  asymmetric — a released slug that a poster still points at sends people to the
  wrong hall.
- **What is a second "St Mary's Village Hall" offered?** The package refuses rather
  than inventing, so the answer belongs to callers — but if every caller writes its
  own suggestion logic, that is this decision's problem all over again.
- **Should a claim record who holds it** beyond `claimedBy`, and should the package
  enforce that only that holder may release it, or is that the caller's ACL?

---
*This document follows the https://specscore.md/feature-specification*
