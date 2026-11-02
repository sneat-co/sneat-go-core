---
format: https://specscore.md/feature-specification
status: Draft
---

# Feature: Shared scheduled work

> [SpecScore.**Studio**](https://specscore.studio): | [Explore](https://specscore.studio/app/github.com/sneat-co/sneat-go-core/spec/features/scheduled-work?op=explore) | [Edit](https://specscore.studio/app/github.com/sneat-co/sneat-go-core/spec/features/scheduled-work?op=edit) | [Ask question](https://specscore.studio/app/github.com/sneat-co/sneat-go-core/spec/features/scheduled-work?op=ask) | [Request change](https://specscore.studio/app/github.com/sneat-co/sneat-go-core/spec/features/scheduled-work?op=request-change) |
**Status:** Draft
**Source Ideas:** —

## Summary

Space-bounded durable background work using shared delayed execution, initially consumed by ibetican.

## Problem

Products need durable scheduled execution without periodically reading all business entities. The existing `strongo/delaying` and `sneat-go-host/taskshost` runtime already supplies typed, authenticated delayed execution; retain it. Add the missing shared durable identity, due-time discovery and transactional claim protocol rather than another executor.

The governing reviewed consumer design is [ibetican shared scheduled work](https://github.com/sneat-co/commitius/blob/integration/ibetican-typed-goals/spec/features/ibetican-product-experience/shared-scheduled-work/README.md). Its first consumer is ibetican. Calendarius and Debtus are future compatibility cases only and remain unchanged.

## Behavior

### Connected lifecycle

A producer creates or changes its business record and scheduled action atomically → the application can close → a bounded indexed query finds eligible work → the existing delayed executor calls its registered handler → the handler asserts the current generation/claim and business authority in the transaction committing its transition or durable external intent → acknowledgment records actual execution outcome. Replacement/cancellation invalidates old deliveries. Crash recovery claims expired leases rather than losing them.

### Shared ownership and contract

`scheduledwork.Ref` binds owning Space, producer, subject and logical action. Its deterministic length-framed key lives at `spaces/{space}/ext/core/scheduledWork/{id}`. Reject missing/invalid Spaces before calling the shared module factory, which otherwise has a global fallback. Producer/action/handler names are bounded; subject identities cannot supply a path. A new business reminder can use a new subject identity without erasing its predecessor's domain history.

The model stores only routing and execution metadata: a supported handler key, opaque subject revision, UTC-comparable microsecond execution time, positive safe generation, next eligible attempt, attempt count, claim token/lease, optional external intent identity and explicit completion or quarantine state. It neither contains private message content nor makes a product's verification, payment or recipient decisions. Syntax validation is not handler registration or authorization.

Pending and expired-lease discovery require separate bounded database-filtered collection-group queries and deployed indexes. Future/terminal rows cannot consume a due batch. Validate the returned Core/Space key chain and stable identity before acting; collection name equality is insufficient. Retry backoff and quarantine must prevent one bad early item from starving later due work.

Schedule/cancel preparation reads during the producer transaction's read phase and writes in that same transaction. External enqueue/provider calls are outside retried callbacks. A transaction-bound claim assertion must fence the owning business mutation or external intent, not merely its later acknowledgment. Rescheduling/cancellation before that commit wins; after external intent commits, cancellation may be too late and must return an honest reconciliation disposition. Stable effect identity survives uncertainty and lease takeover. Generic exhaustion never erases an obligation: only owning policy authorizes a final stop; unresolved work remains recoverable.

Calendarius compatibility includes occurrence identity, explicit timezone/DST normalization and edits/cancellation. Debtus compatibility includes `DtNext` due semantics, reminder replacement/discard, clearing delivery history on replacement and stable provider acknowledgment/uncertainty. Do not import their implementation packages, duplicate their business state or migrate either consumer now.

### Current source boundary and continuation

The first source slice contains the model, validation and deterministic key factory only. Focused capped Go tests/vet pass with100% statement coverage for this new package. A test found that the shared record ID validator alone accepted a path-like subject; the owning reference now rejects path separators and control characters explicitly. This is not a working scheduler, a published provider or production acceptance.

Next implement the DAL transaction participant, indexed due/expired-lease reads, claim/intent/retry/completion operations and existing-executor adapter. Qualify the paused-worker cancellation/lease races, actual Firestore indexes/rules and an ibetican handler journey before provider publication and consumer enablement. Existing ibetican records require a reviewed reconciliation/cutover inventory; do not keep a hidden full-collection scan as a fallback.

## Acceptance Criteria

- **AC: scope** — Given missing/invalid Space or path-like identity, key creation fails without a global fallback; distinct valid producer/subject/action/Space identities do not alias.
- **AC: durable-schedule** — Given create/edit/cancel and an uncertain response, business state and generation-bound scheduling commit atomically or neither commits. Duplicate and obsolete deliveries cannot restore old meaning.
- **AC: due-index** — Given future/terminal prefixes and due/expired-lease work, bounded indexed database queries select the eligible batch without scanning business collections or the complete registry. Real storage/index qualification is required.
- **AC: claim-fence** — Given a paused worker racing cancellation, replacement or lease takeover, the same transaction fences both claim and owning business effect/intent; stale acknowledgments also fail. Prove both race orderings.
- **AC: honest-delivery** — Enqueue acknowledgment is not completion. External uncertainty retains the same effect identity and recovery; exhausted work is not a fulfilled obligation or a fabricated payment result.
- **AC: reusable-consumers** — The public boundary supports ibetican now and Calendarius/Debtus compatibility later without importing product implementations. This change does not migrate either existing product or change financial rules.
- **AC: connected-proof** — Provider release requires an actual owning consumer handler journey with the shared executor and current storage/rules/index evidence. A model-only test or unused API is insufficient.

## Open Questions

Transaction API ergonomics, concrete index definitions and first-consumer existing-record reconciliation remain implementation gates. The model slice does not resolve them or claim release readiness.

---
*This document follows the https://specscore.md/feature-specification*
