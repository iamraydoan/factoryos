# ADR-0007: Global Error Handling Standard

* **Status:** Accepted
* **Date:** 2026-09-14
* **Authors:** @raydoan

---

## 1. Context

`PROJECT_BIBLE.md` §4 states the standard:

> **Error Handling:** Standardized gRPC status codes & RFC-7807 Problem Details for REST APIs.

Neither half was implemented. The error contract that did exist was **orphaned**: `api/contracts/openapi/common/v1/schemas/errors.yaml` defined an `ErrorResponse {code, message, details[]}` shape that nothing referenced, while each domain re-declared its own copy and most path files declared no error responses at all.

Consequences observed in the codebase:

* A malformed path or query parameter escaped as an HTTP 500 with a default framework body.
* An out-of-range page size produced a bare 400 carrying no machine-readable code.
* A bad pagination cursor produced an HTTP 500, because nothing classified it as a client error.

Each is a *user error reported as a server error*, and none carries a code a client can branch on.

**Requirements:**

* One vocabulary shared by every service, in every language.
* Clients must distinguish error classes mechanically — a transport status alone cannot separate a malformed identifier from a bad cursor.
* Domain logic must stay transport-agnostic (`PROJECT_BIBLE.md` §3, Bounded Context Isolation).
* Must satisfy the Schema-First mandate (§1) — the contract precedes the code.
* Must be implementable incrementally; existing services must not break while adopting it.

**Design-first decision:** the standard is authored before the implementation, and the implementation is then validated against it. This is deliberate — the contract is what services agree on, so it must not be reverse-engineered from any one language's code.

---

## 2. Decision

### 2.1 Errors are classified by **category**, not by transport status

A code carries `code`, `category`, `severity`, and `retryable`. It does **not** carry an HTTP status or a gRPC code.

**The domain layer classifies an error by category; a transport adapter maps category → transport code.**

Rationale:

1. **A transport status on a domain error leaks the protocol into the domain.** One failure must produce *both* an HTTP 409 and a gRPC `FAILED_PRECONDITION`; a single status field cannot express that, so it would force one adapter to ignore it and guess. It also violates §3.
2. **Status is derivable; category is not.** The category is the semantic judgement only the domain can make; the status is a mechanical projection of it. Encode what only the domain knows.
3. **Category scales without touching adapters.** A new service declares `category = CONFLICT` and inherits the full mapping.
4. **A custom status annotation was rejected.** It requires an annotation processor, and the value cannot be validated where the error is constructed — the failure surfaces while handling a failure.

### 2.2 Two mappings are special-cased rather than generalised

The standard calls out two exceptions to the category mapping, because a blanket rule produces a client-visible defect:

* **`NOT_FOUND` → gRPC `NOT_FOUND` (5)**, never `UNAVAILABLE` (14). The general "non-request errors are dependency failures → 14" rule is wrong here: it tells a client *"server down, retry"* for an entity that can never appear, and clients with retry logic will hammer it.
* **`CONFLICT` → `ABORTED` (10) only for a lost optimistic lock**, else `FAILED_PRECONDITION` (9). `ABORTED` means *"retry the whole transaction"* — correct for a concurrency loss, wrong for a permanent state-machine rejection. Conflating them causes retry storms. The distinction is carried by the code's naming suffix, which is therefore load-bearing and directly tested.

A third rule resolves cross-service ambiguity: a downstream service's `NOT_FOUND` is **not** this service's 404. It is a `DEPENDENCY` failure, because it means our own data is inconsistent with a dependency.

### 2.3 REST uses RFC 9457 Problem Details plus a `code` extension member

`Content-Type: application/problem+json`, with the RFC's `type`, `title`, `status`, `detail`, `instance` members plus FactoryOS extensions: `code`, `errors[]`, `retryable`, `traceId`, `timestamp`.

`code` is the machine-readable contract clients branch on. `type` is **derived** from `code` (`https://factoryos.dev/errors/<kebab-case-code>`) rather than stored, so the URI cannot drift from the code it identifies.

`title` and `detail` are kept separate per the RFC: `title` identifies the *kind* of error and is constant per code; `detail` describes *this occurrence*.

The schema keeps the existing name `ErrorResponse` and becomes the single canonical error contract. The per-domain copies are deleted — the project is pre-1.0, nothing external consumed them, and retaining duplicates would leave the tree with competing answers to "what does an error look like?".

### 2.4 The taxonomy is language-neutral; implementations follow it

The taxonomy — categories, codes, and both transport mappings — is defined in `docs/07-api/ERROR_HANDLING.md` and the contract schema. Each language then implements three things: the taxonomy, a domain error base that imports no transport type, and one adapter per protocol.

This is stated as a decision because the alternative — letting each language define its own error vocabulary — is what produced the current state. The standard is the shared artifact; implementations are validated against it.

### 2.5 Adoption is incremental

Go services currently use inline per-method status construction with no shared vocabulary and no structured detail. That is functional but unclassified. Migration proceeds service by service, and **requires no wire-format change**: `google.rpc.Status` with `ErrorInfo` and `BadRequest` details is language-neutral and standard, so both languages already agree on the bytes.

---

## 3. Consequences

### Positive Impacts

* **One vocabulary across services and protocols.** The gRPC `ErrorInfo.reason` and the REST `code` are the same string, so a client that switches on `code` behaves identically over both.
* **One error contract in the tree.** Duplicated per-domain error shapes are gone; every spec references the canonical schema.
* **No adapter logic per new entity.** A new code references an existing category and inherits the mapping.
* **`PROJECT_BIBLE.md` §4 is satisfied** by a documented, testable standard rather than an aspiration.
* **Log levels derive from declared severity**, so expected client errors and genuine outages are no longer logged identically.
* **No per-endpoint wiring.** Handlers fail; the adapter translates. No handler needs its own error-handling branch.
* **Client errors stop being reported as server errors** — the specific defects that motivated this ADR.

### Negative Impacts & Trade-offs

* **The taxonomy must be defined in two places per language** — the shared codes and each domain's codes. *Mitigation:* the standard is the single documented source, the schema declares `code` as a plain string rather than a closed enum, and exhaustive tests assert every code's category and both projections.
* **A new category requires editing every adapter.** Deliberate: it forces the mapping to be considered rather than silently defaulted.
* **The REST schema is partly documentation.** The OpenAPI schema cannot express everything the runtime produces, so schema and behaviour can drift. *Mitigation:* a round-trip test asserts the emitted payload satisfies the schema's required members and the `code`→`type` derivation.
* **Code-name suffixes carry semantics.** A rename can silently change a transport code. *Mitigation:* the suffix rule is stated explicitly and the affected codes are asserted in tests.
* **Cross-language adoption is not simultaneous.** Until a service adopts the standard, it emits unclassified errors, so clients see mixed quality during the transition.

---

## 4. Alternatives Considered

* **Use gRPC status codes as the shared vocabulary.** Rejected: 17 codes are too coarse to be a contract. `INVALID_ARGUMENT` cannot distinguish a malformed identifier, a bad cursor, and a page-size violation — and clients need to. They are the right *transport projection*, not the right taxonomy.

* **Put an HTTP status on the domain error base.** Rejected: couples the domain to a protocol, makes the gRPC mapping meaningless, and violates `PROJECT_BIBLE.md` §3.

* **A custom status annotation on each error type.** Rejected: requires an annotation processor, and the value cannot be validated where the error is constructed.

* **Keep the simple `{code, message, details[]}` shape.** Rejected: leaves the `PROJECT_BIBLE.md` §4 mandate unmet, gives clients no `type` URI or category to switch on, and collapses "what kind of error" and "what happened this time" into one field.

* **Let each service define its own error vocabulary.** Rejected: this is the current state, and it produced inconsistent, unclassified errors across services.

* **Per-service error modules rather than a shared design.** Rejected: guarantees drift in exactly the vocabulary that must be identical across services.

* **Content negotiation** — emit the older shape unless the client requests `application/problem+json`. Rejected: doubles the serialization paths and the test surface for every error, to serve a client population that the pre-1.0 deletion already accommodates.

* **Generate the taxonomy from the schema.** Rejected: it would require `code` to be a closed enum in the schema, inverting the intended dependency and recreating the drift problem in the opposite direction.

---

## 5. References

* [ERROR_HANDLING.md](../07-api/ERROR_HANDLING.md) — the standard: categories, mappings, wire formats, taxonomy
* [`api/contracts/openapi/common/v1/schemas/errors.yaml`](../../api/contracts/openapi/common/v1/schemas/errors.yaml) — the REST error schema
* [PAGINATION_DESIGN.md](../07-api/PAGINATION_DESIGN.md) — cursor validation errors
* [PROJECT_BIBLE.md](../00-governance/PROJECT_BIBLE.md) §3, §4
* [EPIC-001: Platform Foundation](../09-epics/EPIC-001-Platform-Foundation.md) — "Platform Services & SDK"
* [RFC 9457 — Problem Details for HTTP APIs](https://www.rfc-editor.org/rfc/rfc9457)
