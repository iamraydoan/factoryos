[Index](README.md) · [← Part 7 — Proving the taxonomy](07-proving-the-taxonomy.md)

# Part 8 — Rolling it out

### What you have built

Two shared modules and one service's adoption of them:

| Artifact | Contents |
|---|---|
| `libs/factoryos-common-error` | `ErrorCategory`, `ErrorSeverity`, `ErrorCode`, `CommonErrorCode`, `ErrorKey`, `DomainException` — no transport dependency |
| `libs/factoryos-common-error-adapters` | `ProblemDetailFactory`, `ProblemDetailExceptionHandler`, `TraceIdFilter`, `GrpcStatusFactory`, `GrpcErrorAdvice`, `ErrorHandlingAutoConfiguration` |
| `services/production-service` | `WorkOrderErrorCode`, classified cursor and page-size failures, both adapters active |

### The four defects from Part 0, rechecked

| Defect (ADR-0007 §1) | Status |
|---|---|
| A malformed path or query parameter escaped as an HTTP 500 | **Fixed** — `MALFORMED_REQUEST`, 400 |
| An out-of-range page size produced a bare 400 with no machine-readable code | **Fixed** — `PAGE_SIZE_OUT_OF_RANGE` with `errors[]` detail |
| A bad pagination cursor produced an HTTP 500 | **Fixed** — three distinct cursor codes, all 400 |
| Nothing carried a code a client can branch on | **Fixed** — `code` is identical over REST and gRPC |

### Adopting it in another service

The pattern is the same for each remaining service. Per service, three steps:

1. **Add both dependencies** to that service's `pom.xml` — versionless, since the parent
   manages them. The auto-configuration then registers the adapters with no further code.
2. **Declare that domain's codes.** A new enum in that service's package implementing
   `ErrorCode`, following the naming convention from `ERROR_HANDLING.md` §6:

   ```java
   public enum NcrErrorCode implements ErrorCode {
       NCR_NOT_FOUND(ErrorCategory.NOT_FOUND, "NCR not found", ErrorSeverity.INFO, false),
       NCR_ALREADY_DISPOSITIONED(ErrorCategory.CONFLICT, "NCR already dispositioned",
               ErrorSeverity.WARNING, false);
       // ...
   }
   ```

3. **Reclassify that service's failures** — replace raw `IllegalArgumentException` and
   message-only exceptions with `DomainException` carrying a code.

Current state of the other four services, so you know what to expect:

| Service | State |
|---|---|
| `quality-service` | `QualityServiceApplication.java` only — little to reclassify yet |
| `warehouse-service` | `WarehouseServiceApplication.java` only |
| `maintenance-service` | `MaintenanceServiceApplication.java` only |
| `planning-service` | no `src/main/java` tree yet |

So for three of them, step 3 is mostly about *not* repeating the old pattern as you add
endpoints. The infrastructure is waiting.

> **A service declares codes; it never writes projection logic.** `ERROR_HANDLING.md` §6: the
> category is the *default* input to every adapter, and a code-level transport override — when
> protocol semantics genuinely require one — is declared through the `GrpcStatusOverrideProvider`
> bean (Part 6a), not by branching in service code. A service never maps a status itself.

### Tick the EPIC

`docs/09-epics/EPIC-001-Platform-Foundation.md` has two error-handling tasks. One is now done.
Change:

```markdown
- [ ] Implement the shared error taxonomy and transport adapters for Java services per [ERROR_HANDLING.md](../ERROR_HANDLING.md).
```

to:

```markdown
- [x] Implement the shared error taxonomy and transport adapters for Java services per [ERROR_HANDLING.md](../ERROR_HANDLING.md).
```

**Leave the Go task unchecked.** It is deliberately out of scope here:

```markdown
- [ ] Adopt the error taxonomy in Go services per [ERROR_HANDLING.md](../ERROR_HANDLING.md) §8.
```

Go adoption is a separate piece of work. It requires no wire-format change — `google.rpc.Status`
with `ErrorInfo` details is already language-neutral, so a Go client decodes what this Java
service emits today.

### Add a CHANGELOG entry

`CLAUDE.md` requires it, and the repo's convention is entries under `[Unreleased]` grouped by
category. Add to the existing `### Added` section:

```markdown
- **Shared error taxonomy & transport adapters (`libs/factoryos-common-error`, `libs/factoryos-common-error-adapters`):** Category-based error taxonomy with `DomainException` and `CommonErrorCode`, plus one adapter per protocol — RFC 9457 Problem Details over HTTP and `google.rpc.Status` with `ErrorInfo`/`BadRequest` over gRPC. Zero transport dependency in the taxonomy module, enforced by its POM. Implements [ERROR_HANDLING.md](docs/07-api/ERROR_HANDLING.md) and [ADR-0007](docs/05-adr/0007-global-error-handling-standard.md).
- **Error taxonomy adoption (`services/production-service`):** `WorkOrderErrorCode`, cursor failures split into `INVALID_CURSOR` / `UNSUPPORTED_CURSOR_VERSION` / `CURSOR_SORT_KEY_MISMATCH` per [PAGINATION_DESIGN.md](docs/07-api/PAGINATION_DESIGN.md) §5.5, page-size and malformed-UUID failures classified as client errors. Fixes the four defects recorded in ADR-0007 §1 — malformed parameters and bad cursors previously escaped as HTTP 500.
```

And to `### Fixed`:

```markdown
- **`services/production-service`:** An out-of-range `limit`/`page`, a malformed pagination cursor, and a malformed `workCenterId` over gRPC now return classified client errors (400 / `INVALID_ARGUMENT`) instead of HTTP 500 / gRPC `INTERNAL`. Each carries a machine-readable `code` and, where applicable, `errors[]` field detail.
```

### Not covered here

Stated plainly so you are not left wondering:

- **Go services.** Separate task, no wire-format change needed (§8 above).
- **`AUTHENTICATION` / `AUTHORIZATION` wiring.** The categories, codes
  (`UNAUTHENTICATED`, `PERMISSION_DENIED`) and both projections exist and are tested, but no
  identity provider is wired — Zitadel is not yet stood up (`EPIC-001` infrastructure items
  are unchecked). The taxonomy is ready when it is.
- **Event-driven error handling.** The taxonomy covers request/response. Errors in Kafka
  consumers and the transactional outbox are a separate concern with different semantics —
  there is no caller to return a status to.

### Final check

```bash
cd services && mvn -o -B clean test
```

Expect **193 tests, 0 failures** across all eight modules. If that is green, the standard is
implemented, adopted in one service, and proven by tests.

### Check yourself

- EPIC-001's Java task is `[x]`; the Go task is still `[ ]`
- A CHANGELOG entry exists under `[Unreleased]` → `### Added`
- `mvn -o -B clean test` is green with **193 tests**
- You can describe the three steps to adopt this in `quality-service`
- You can explain why Go adoption needs no wire-format change

---

[← Part 7 — Proving the taxonomy](07-proving-the-taxonomy.md) · [Index](README.md)
