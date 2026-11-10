---
format: https://specscore.md/feature-specification
status: Draft
---

# Feature: External auth policy

> [SpecScore.**Studio**](https://specscore.studio): | [Explore](https://specscore.studio/app/github.com/sneat-co/sneat-go-core/spec/features/external-auth-policy?op=explore) | [Edit](https://specscore.studio/app/github.com/sneat-co/sneat-go-core/spec/features/external-auth-policy?op=edit) | [Ask question](https://specscore.studio/app/github.com/sneat-co/sneat-go-core/spec/features/external-auth-policy?op=ask) | [Request change](https://specscore.studio/app/github.com/sneat-co/sneat-go-core/spec/features/external-auth-policy?op=request-change) |

**Status:** Draft
**Source Ideas:** —
**Scope:** T09a bounded provider foundation (sneatauth / apicore).

Implementation is present locally; independent code review, CI and publication remain pending.

## Summary

Immutable reserved-UID policy and verifier wrapper for trusted external authentication composition.

## Problem

In multi-tenant or system-administered platforms, certain subject identities (UIDs) are reserved strictly for internal system processes and trusted composition. If an external authentication provider (such as Firebase Auth or an external OAuth provider) mints an external identity matching a reserved system UID (either accidentally or through post-bootstrap account creation), downstream services must never authenticate that credential as the reserved entity.

Authentication gates at the HTTP entry boundary must deterministically reject any presented token carrying a reserved UID, empty UID, or whitespace-only UID, regardless of client headers, requested scopes, or claims asserting administrative status.

## Architecture and Behavior

### 1. Stable Error Classifications

The package exposes three stable sentinel errors classified using `errors.Is`:
- `ErrInvalidExternalAuthConfig`: Indicates invalid policy or wrapper construction (e.g. nil/empty reserved set, blank entries, nil verifier, or uninitialized nil/zero policy).
- `ErrReservedExternalUID`: Indicates that an externally presented token matched an exact reserved UID in the policy.
- `ErrInvalidExternalUID`: Indicates that required external authentication yielded no token, or that a presented token carries an empty or whitespace-only UID.

### 2. Immutable Policy Construction (`NewReservedExternalUIDPolicy`)

- Trusted code constructs the policy with an explicit non-empty slice of reserved UIDs:
  ```go
  policy, err := sneatauth.NewReservedExternalUIDPolicy([]string{"sys-principal-1", "sys-principal-2"})
  ```
- Rejection of invalid inputs: Nil or empty slices, as well as any empty or whitespace-only string entries, immediately return `(nil, ErrInvalidExternalAuthConfig)`.
- Deduplication: Duplicate entries are deduplicated into private storage.
- Defensive copy: Input slices are defensively copied into an immutable internal map; subsequent caller mutation of the source slice cannot alter the policy.
- Exact nonblank identity matching: Nonblank IDs are preserved exactly as supplied. No case conversion, prefix trimming, or whitespace stripping is performed on policy entries or checked UIDs. Case variants, distinct prefixes, and whitespace-wrapped nonblank characters remain distinct.
- Fail closed: Any public check (`policy.CheckUID(uid)`) on a `nil` or zero-value policy fails closed with `ErrInvalidExternalAuthConfig`. It never acts as an allow-all policy.

### 3. Captured-Verifier Wrapper (`WrapExternalTokenVerifier`)

- Adapts an existing `ExternalTokenVerifier` function `func(r *http.Request, authRequired bool) (*Token, error)`:
  ```go
  wrapped, err := sneatauth.WrapExternalTokenVerifier(rawVerifier, policy)
  ```
- Construction validation: Fails closed returning `(nil, ErrInvalidExternalAuthConfig)` without panicking if either the verifier is nil or the policy is nil/zero.
- Invocation contract:
  - Forwards the identical `*http.Request` pointer and `authRequired` boolean flag to the captured verifier exactly once.
  - Error precedence: If the verifier returns an error, the wrapper returns `nil` token and that exact SAME error object, discarding any accompanying token.
  - Optional vs required authentication:
    - If `tok == nil` and `authRequired == false`, returns `nil, nil` (anonymous request preserved).
    - If `tok == nil` and `authRequired == true`, returns `nil, ErrInvalidExternalUID`.
  - Nonblank UID validation: Any non-nil token with an empty or whitespace-only UID is rejected with `ErrInvalidExternalUID` even when authentication is optional.
  - Reserved UID denial: A token matching a reserved UID is rejected with `ErrReservedExternalUID` regardless of `Issuer`, `IsAdmin`, `Original`, or other claims.
  - Preservation: On success, returns the identical `*Token` pointer with all fields preserved.
  - Per-invocation enforcement: Evaluates each token on every call; does not cache decisions or run only at startup.

### 4. HTTP Boundary Conformance (`apicore`)

- `VerifyRequest`:
  - On denial, returns an error classified as `facade.ErrUnauthorized` (via `%w: %v` wrapping).
  - Does not itself write the authentication-error response; existing origin/content validation behavior remains unchanged.
  - Leaves the returned context without an authenticated `sneatauth.Token`.
- `VerifyRequestAndCreateUserContext`:
  - On denial, returns a `nil` `ContextWithUser`, returns an error classified as `facade.ErrUnauthorized`, and writes an HTTP 401 response with JSON error details via `httpserver.HandleError`.
  - Rejects blank/whitespace UIDs before attempting to call `facade.NewUserContext`, preventing panics.
- Boundary error identity note:
  - While direct wrapper calls preserve verifier error identity and the policy exposes `ErrReservedExternalUID`, `apicore.VerifyRequest` formats the underlying error using `%v`, intentionally classifying the result as `facade.ErrUnauthorized` while dropping inner `errors.Is` identity. Tests independently verify both boundaries.

### 5. DAL Bypass Non-Interference

`facade.SpaceAccessAuthorizedByCaller` is a Data Access Layer (DAL) contract opt-out that permits trusted internal space operations to bypass membership checks by returning an empty User ID. It does not contain an authenticated `sneatauth.Token`, cannot be minted or configured by HTTP headers/DTO bodies, and cannot bypass external token verification.

## Downstream Activation Gate

> [!IMPORTANT]
> This increment (T09a) implements the bounded generic provider mechanism in `sneat-go-core`. It does **NOT** activate CommittedMe or complete full system protection (T09/P19).
>
> Full product protection requires the downstream host consumer to:
> 1. Supply the concrete product reserved UID set at initialization time.
> 2. Wrap all downstream token verification entry points.
> 3. Enforce reserved UID checks across all account mutation and creation paths.
> 4. Establish overlapping Firestore security rule denials.
> 5. Wire native authority composition and bootstrap prerequisites (P1–P4, P10, P19–P22).
>
> System principal capabilities remain disabled until these subsequent increments are reviewed and landed.


## Requirements and acceptance criteria

### REQ: immutable-policy

Trusted composition supplies a nonempty reserved-UID set. Reject blank entries and nil/zero policies, copy input storage, deduplicate exact entries, and preserve exact nonblank identities.

#### AC: policy-construction-and-matching

Given nil, empty, blank-containing, duplicate and ordinary reserved-ID inputs,
When the policy is constructed and queried, including after the caller mutates its input slice,
Then invalid configuration returns `ErrInvalidExternalAuthConfig`, original exact reserved IDs remain denied, and distinct case/prefix/nonblank whitespace variants remain distinct.
(verifies REQ: immutable-policy)

### REQ: checked-verifier

Capture a nonnil verifier and initialized policy; forward the same request and authentication flag exactly once. Return no token on a verifier or policy error. Preserve verifier error identity and ordinary token identity/fields; check every invocation.

#### AC: verifier-errors-and-identities

Given a verifier returning an error with or without a token, optional anonymity, missing required identity, blank nonnil identity, a reserved identity or an ordinary identity,
When the wrapper is invoked,
Then it applies the documented error precedence and sentinel classifications, returns no token on denial, preserves optional anonymity and ordinary token identity, and forwards the original request and flag once.
(verifies REQ: checked-verifier)

#### AC: repeated-and-concurrent-checks

Given an immutable policy and successive reserved/ordinary verified token results,
When calls repeat and policy reads run concurrently,
Then each invocation uses the current token, reserved identities remain denied, ordinary identities remain allowed, and policy reads introduce no data race.
(verifies REQ: checked-verifier)

### REQ: entry-conformance

The wrapper composes with both real apicore entry functions without changing their production error handling or facade construction. External request data cannot configure policy or mint system authority.

#### AC: http-denial-and-ordinary-users

Given the wrapper installed at the real token-getter boundary and reserved, malformed or ordinary verified token results,
When either apicore entry function verifies a request,
Then denials classify as `facade.ErrUnauthorized` without accepting a token; the handled entry returns nil user context and HTTP401, and ordinary/optional-anonymous requests retain their existing outcomes.
(verifies REQ: entry-conformance)

#### AC: claims-and-dal-bypass

Given a reserved outer UID accompanied by privileged-looking claims, headers or request data,
When the real verification boundary is called,
Then it still rejects that UID; no policy is configured from the request, and the unchanged caller-authorized DAL context creates no authenticated token or system principal.
(verifies REQ: entry-conformance)

## Verification and delivery

Unit checks are in `sneatauth/external_uid_policy_test.go`; real boundary conformance is in `apicore/request_validation_external_auth_test.go`. Existing production apicore/facade behavior is unchanged. Local focused tests and the affected-package coverage profile pass with100% statements. Local race execution is unavailable because this VM has no C compiler; the unchanged root-module CI explicitly runs `go test -race` and enforces100% statement coverage, and must pass before delivery. Independent Astra code approval and native WB delivery/publication are still required.

The downstream gate comes from the approved [System principal and platform prerequisites](https://github.com/sneat-co/commitius/blob/4b7edcef11c5385e0fca291b34bc79984c7f910f/spec/features/committedme-group-spaces/README.md#system-principal). No Firebase post-bootstrap scenario, deployed security-rule behavior, trusted native principal or whole-product protection is claimed by these provider conformance tests.

## Open Questions

No unresolved behavior for this bounded contract. Concrete product reserved IDs, downstream wiring, native/bootstrap authority and complete activation proof remain later reviewed work under the gate above.

---
*This document follows the https://specscore.md/feature-specification*
