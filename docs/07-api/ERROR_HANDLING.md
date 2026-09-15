# Error Handling Design Standard

> This document defines the error taxonomy and transport mapping for all FactoryOS APIs.
>
> **Status:** Accepted — see [ADR-0007](../05-adr/0007-global-error-handling-standard.md).
>
> **Language-neutral.** It specifies the contract and the rules — not the code that
> implements them. Any service, in any language, satisfies this standard by honouring
> the taxonomy and the wire formats below.
>
> **Contract:** [`api/contracts/openapi/common/v1/schemas/errors.yaml`](../../api/contracts/openapi/common/v1/schemas/errors.yaml) — the REST wire format.

---

## 1. The Core Rule

> **The domain layer classifies an error by its *category*. A transport adapter maps category → transport code.**

An error is classified once, semantically, by the layer that understands it. It is then *projected* onto HTTP and gRPC by adapters. Nothing in the domain layer knows which protocol delivered the request.

```
Domain layer         classifies the failure → category
      ↓
Domain error         carries: code, category, severity, retryable, detail
      ↓
Transport adapter    projects category → HTTP status / gRPC status code
```

Four consequences:

1. **A domain error carries no transport status.** One failure must project to *both* an HTTP 409 and a gRPC `FAILED_PRECONDITION`. A single status field cannot express that, so it would force one adapter to ignore it and guess.
2. **A new entity adds no adapter logic.** It declares codes referencing an existing category and inherits the entire mapping.
3. **Both adapters derive from the category and nothing else**, so the two projections can never disagree about what kind of error occurred.
4. **The mapping must be exhaustive over categories** — adding one is a deliberate act that forces both adapters to be updated, not a silently-defaulted case.

### Why not use transport codes as the shared vocabulary?

gRPC has only 17 codes, and they are too coarse to be a *contract*: `INVALID_ARGUMENT` covers a malformed identifier, a bad cursor, and a page-size violation, and clients need to tell those apart. Transport codes are the right **projection**, not the right taxonomy. The same reasoning applies to HTTP statuses.

---

## 2. Error Categories

The category is the single input to both transport projections.

| Category | Meaning | gRPC | HTTP | Retryable | Log |
|---|---|---|---|---|---|
| `VALIDATION` | Malformed or missing request data | `INVALID_ARGUMENT` (3) | 400 | no | INFO |
| `INVALID_CURSOR` | Bad, expired, or mismatched pagination cursor | `INVALID_ARGUMENT` (3) | 400 | no | INFO |
| `AUTHENTICATION` | No credentials, or credentials rejected | `UNAUTHENTICATED` (16) | 401 | no | WARNING |
| `AUTHORIZATION` | Authenticated, but not permitted | `PERMISSION_DENIED` (7) | 403 | no | WARNING |
| `NOT_FOUND` | Entity does not exist **in this service** | `NOT_FOUND` (5) | 404 | no | INFO |
| `CONFLICT` | Illegal transition, duplicate, or lost optimistic lock | `FAILED_PRECONDITION` (9) / `ABORTED` (10) | 409 | no / **yes** | WARNING |
| `RATE_LIMIT` | Caller exceeded a quota | `RESOURCE_EXHAUSTED` (8) | 429 | yes | WARNING |
| `DEPENDENCY` | A downstream service or datastore is unavailable | `UNAVAILABLE` (14) | 503 | yes | ERROR |
| `INTERNAL` | Bug, corrupt data, or unhandled failure | `INTERNAL` (13) | 500 | no | CRITICAL |

`INVALID_CURSOR` is separate from `VALIDATION` because pagination cursors are the one error class with a dedicated specification and an already-defined message set (see [PAGINATION_DESIGN.md](PAGINATION_DESIGN.md) §5.5).

**Categories are a closed set.** Adding one means editing every adapter — that is the point. Prefer adding a new *code* within an existing category.

---

## 3. Two Mappings That Must Not Be Generalised

Both are why the mapping is a per-category decision and not a blanket rule. Getting either wrong produces a client-visible defect, not a cosmetic one.

### 3.1 `NOT_FOUND` → gRPC `NOT_FOUND` (5), never `UNAVAILABLE` (14)

The general rule "errors about something other than the incoming request are dependency failures → 14" is correct for `DEPENDENCY` and **wrong** for `NOT_FOUND`, which is about an entity *this service owns*.

Returning `UNAVAILABLE` for a missing work order tells the client *"the server is down, retry later"*. A client with retry logic will then hammer a request that can never succeed. `NOT_FOUND` therefore needs its own branch rather than falling through to the dependency default.

### 3.2 `CONFLICT` splits — `ABORTED` (10) only for lost optimistic locks

`ABORTED` carries a specific meaning in gRPC: *"the operation was aborted, typically due to a concurrency issue such as a sequencer check failure — the client should retry the whole transaction."*

| Conflict kind | gRPC | Why |
|---|---|---|
| Lost optimistic lock (concurrent write won) | `ABORTED` (10) | Genuinely retryable — the client should re-read and retry the whole transaction |
| Illegal state transition | `FAILED_PRECONDITION` (9) | Permanent — retrying will fail forever |
| Duplicate entity | `FAILED_PRECONDITION` (9) | Permanent unless the input changes |

Conflating these causes retry storms against requests that cannot succeed. Because the distinction is not expressible in the category alone, it must be derivable from the **code** — which is why code naming carries meaning (see §6).

### 3.3 A downstream `NOT_FOUND` is not this service's `NOT_FOUND`

When a call to another service returns `NOT_FOUND`, that is **not** a 404 for your caller. The caller asked for something *you* own; your dependency is telling you your own data is inconsistent.

| Situation | Category | Returned to caller | HTTP |
|---|---|---|---|
| Entity absent in **this** service | `NOT_FOUND` | gRPC `NOT_FOUND` (5) | 404 |
| Entity absent in a **downstream** service | `DEPENDENCY` | gRPC `FAILED_PRECONDITION` (9) | 500 |
| Downstream service unreachable | `DEPENDENCY` | gRPC `UNAVAILABLE` (14) | 503 |

---

## 4. REST Wire Format — RFC 9457 Problem Details

REST errors return **`Content-Type: application/problem+json`**.

```http
HTTP/1.1 409 Conflict
Content-Type: application/problem+json

{
  "type": "https://factoryos.dev/errors/work-order-invalid-transition",
  "title": "Invalid state transition",
  "status": 409,
  "detail": "Work order cannot go from 'closed' to 'released'.",
  "instance": "/api/v1/work-orders/0192f3a1-7c4e-7b2a-9f01-3d5e8a1b2c34",
  "code": "WORK_ORDER_INVALID_TRANSITION",
  "retryable": false,
  "timestamp": "2026-09-15T08:30:00Z",
  "errors": [
    {
      "field": "state",
      "current": "closed",
      "allowed": ["released"]
    }
  ]
}
```

### Field reference

| Field | Required | Source | Notes |
|---|---|---|---|
| `type` | **yes** | Derived from `code` | `https://factoryos.dev/errors/<kebab-case-code>`. Never stored independently, so it cannot drift from `code`. |
| `title` | **yes** | The code's definition | Short summary, constant for every instance of the code. Not the occurrence message. |
| `status` | **yes** | The category | HTTP status, repeated in the body per RFC 9457. |
| `code` | **yes** | The code's definition | The machine-readable contract. **Clients branch on this, not on `status`.** |
| `detail` | no | The occurrence | Human-readable, specific to this failure. |
| `instance` | no | The request | Request path that produced the problem. |
| `retryable` | no | The code's definition | Stated explicitly, not inferred from `status` — see §3.2. |
| `traceId` | no | Generated | Correlates a client report with a server stack trace. **Required on 5xx.** |
| `timestamp` | no | Generated | ISO-8601, UTC. |
| `errors[]` | no | The occurrence | Field-level detail; see below. |

`title` and `detail` are deliberately separate: `title` identifies *the kind* of error and is stable per code; `detail` describes *this occurrence*. A single `message` field collapses the two and leaves clients unable to distinguish them.

### `errors[]` members

| Field | Required | Notes |
|---|---|---|
| `field` | **yes** | The offending field. Use a shared constant per field name so client and server cannot disagree on spelling. |
| `code` | no | A sub-code for the field, when useful. |
| `message` | no | Field-specific explanation. |
| `current` | no | The value observed. Present for state and conflict errors. |
| `allowed` | no | Valid alternatives, when enumerable. |

### Rules

- **Never echo an internal error message on a 5xx.** Unhandled exceptions carry variable names, table and column names, or absolute file paths. Return a constant `detail` plus a `traceId`; put the detail in the log.
- **`code` is a string, not a closed enum in the schema.** Enumerating every code in the OpenAPI schema *and* in each service guarantees the two drift. The taxonomy is defined once (§6); the schema documents the shape.
- **Errors are declared once, canonically**, in `api/contracts/openapi/common/v1/schemas/errors.yaml`. Domain specs reference it rather than re-declaring a shape.
- **All `/api/v1` endpoints use `application/problem+json`** for errors. There is no exemption for any path.

---

## 5. gRPC Wire Format

Errors attach a typed `google.rpc.Status` to the trailing metadata, so any gRPC client — in any language — can decode the detail without parsing strings.

| Detail payload | Carries |
|---|---|
| `ErrorInfo` | `reason` = the error `code`; `domain` = `factoryos`; metadata `category`, `retryable` |
| `BadRequest` | `fieldViolations[]`, derived from `errors[]` |

`ErrorInfo.reason` holds the **same** `code` string the REST response uses, so one vocabulary spans both protocols. A client that switches on `code` works identically over HTTP and gRPC.

### Rules

- **The detail is protobuf, not a formatted string.** A string message cannot be parsed reliably across languages and breaks whenever the format changes.
- **Use the standard `google.rpc` detail types**, not a bespoke message, so generic gRPC tooling and proxies can render the error without knowing FactoryOS.
- **A client must be able to classify an error without reading `detail`** — from the status code plus `ErrorInfo.reason`.

---

## 6. The Code Taxonomy

A **code** is the machine-readable identifier of a specific failure. It is the finest granularity of the contract.

### Naming

`<DOMAIN>_<REASON>` — uppercase, underscore-separated: `WORK_ORDER_NOT_FOUND`, `MATERIAL_LOT_ALREADY_EXISTS`.

The name should read as the sentence *"<what> <went wrong>"*. Prefix domain-specific codes with their owning domain; shared codes carry no domain prefix.

### Every code declares

| Attribute | Purpose |
|---|---|
| `code` | The identifier itself. |
| `category` | Drives both transport projections (§2). |
| `title` | Short human-readable summary, constant per code. |
| `severity` | The log level this failure warrants (§7). |
| `retryable` | Whether the caller may retry the identical request. |

### Naming convention by suffix

Suffixes carry meaning, because some mappings are derived from them (§3.2):

| Suffix | Category | Effect |
|---|---|---|
| `_NOT_FOUND` | `NOT_FOUND` | 404 / gRPC 5 |
| `_ALREADY_EXISTS` | `CONFLICT` | 409 / gRPC 9 |
| `_INVALID_TRANSITION` | `CONFLICT` | 409 / gRPC 9 — permanent |
| `_CONCURRENT_MODIFICATION` | `CONFLICT` | 409 / gRPC **10** — retryable |
| `_REQUIRED`, `INVALID_*`, `MALFORMED_*` | `VALIDATION` | 400 / gRPC 3 |
| `*_UNAVAILABLE` | `DEPENDENCY` | 503 / gRPC 14 |

> **`_CONCURRENT_MODIFICATION` is load-bearing.** It is the only signal separating a retryable conflict from a permanent one. Renaming a code with this suffix changes its gRPC status code, so every such code must be asserted explicitly in tests.

### Shared codes

These apply to every service and must be resolvable without a domain context:

| Code | Category | Meaning |
|---|---|---|
| `INTERNAL_ERROR` | `INTERNAL` | Unhandled failure |
| `DEPENDENCY_UNAVAILABLE` | `DEPENDENCY` | A downstream service or datastore is unreachable |
| `RESOURCE_EXHAUSTED` | `RATE_LIMIT` | A quota was exceeded |
| `UNAUTHENTICATED` | `AUTHENTICATION` | No credentials, or rejected |
| `PERMISSION_DENIED` | `AUTHORIZATION` | Not permitted |
| `MALFORMED_REQUEST` | `VALIDATION` | The request could not be parsed |
| `MISSING_REQUIRED_FIELD` | `VALIDATION` | A required field is absent |
| `INVALID_CURSOR` | `INVALID_CURSOR` | Cursor malformed or undecodable |
| `UNSUPPORTED_CURSOR_VERSION` | `INVALID_CURSOR` | Cursor version not supported |
| `CURSOR_SORT_KEY_MISMATCH` | `INVALID_CURSOR` | Cursor does not match the query's sort keys |
| `PAGE_SIZE_OUT_OF_RANGE` | `VALIDATION` | Page size outside the permitted range |

Domain codes are owned by the domain that defines them and must never be reused across domains.

### Adding a code

1. **Reuse an existing category.** Adding a category means editing every adapter — a deliberate act, not a convenience.
2. **Name it per the convention above**, and pick the suffix deliberately — it may drive the mapping.
3. **Define** its title, category, severity, and retryable flag in one place, in the owning domain's taxonomy.
4. **Do not add mapping logic.** The category is the only input to the adapters; if a new code needs a new mapping, the design is wrong.
5. **Assert it.** Every code must have a test asserting its category and both transport projections. Codes whose suffix affects the mapping must be asserted explicitly.
6. **Document it** in this table set if it is shared, or in the owning domain's design doc if it is domain-specific.

---

## 7. Logging

Log level follows the code's declared **severity**, not a choice made at each throw site:

| Severity | Level | Use |
|---|---|---|
| `INFO` | INFO | Expected client behaviour — a bad cursor, a 404. |
| `WARNING` | WARN | Worth attention — a conflict, a denied permission. |
| `ERROR` | ERROR + stack trace | Service-level failure — a downstream outage. |
| `CRITICAL` | ERROR + stack trace | Availability or data-integrity impact. Alert. |

This is why severity belongs to the code rather than the call site: a bad cursor and a broken database connection are both "request failed", but only one should page an on-call engineer.

**Every 5xx logs a stack trace and returns a `traceId`.** The client receives the id; the log receives the detail.

---

## 8. Adoption

The taxonomy is language-neutral, so each language implements the same three things:

| Concern | Requirement |
|---|---|
| **Taxonomy** | Codes with their category, title, severity, and retryable flag — shared codes centrally, domain codes per domain |
| **Domain error base** | A base error carrying code + detail, importing no transport type |
| **Transport adapters** | One per protocol, mapping category → status, written once and reused by every handler |

**Wire-format compatibility is already ensured.** `google.rpc.Status` with `ErrorInfo` and `BadRequest` details is language-neutral and standard, so a Go client decodes exactly what a Java service emitted, and vice versa.

**The order matters:** the taxonomy and this standard define the contract; implementations follow and are validated against it. A service adopting the standard should be able to pass the taxonomy assertions in §6.5 without changing them.

---

## 9. References

- [ADR-0007: Global Error Handling Standard](../05-adr/0007-global-error-handling-standard.md) — the decision and its rationale
- [`api/contracts/openapi/common/v1/schemas/errors.yaml`](../../api/contracts/openapi/common/v1/schemas/errors.yaml) — the REST error schema
- [PAGINATION_DESIGN.md](PAGINATION_DESIGN.md) — cursor validation errors
- [RFC 9457 — Problem Details for HTTP APIs](https://www.rfc-editor.org/rfc/rfc9457)
- [gRPC Status Codes](https://grpc.io/docs/guides/status-codes/)
