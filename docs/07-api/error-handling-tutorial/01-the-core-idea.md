[Index](README.md) · [← Part 0 — Before you start](00-before-you-start.md)

# Part 1 — The one idea

Everything else in this tutorial follows from a single rule. From
[ERROR_HANDLING.md §1](../ERROR_HANDLING.md):

> **The domain layer classifies an error by its *category*. A transport adapter maps category → transport code.**

An error is classified **once**, semantically, by the layer that understands it. It is then
*projected* onto HTTP and gRPC by adapters. Nothing in the domain layer knows which protocol
delivered the request.

```
Domain layer         classifies the failure → category
      ↓
Domain error         carries: code, category, severity, retryable, detail
      ↓
Transport adapter    projects category → HTTP status / gRPC status code
```

### What "projection" means

If you are new to this idea, here it is in plain terms:

- The **domain** says *what kind of failure happened* — "this is a validation failure".
- The **adapter** says *how that kind of failure is spelled in this protocol* — "validation
  failures are HTTP 400" and "validation failures are gRPC `INVALID_ARGUMENT`".

Same failure, two spellings. The domain does not choose the spelling.

### Why the domain error carries no HTTP status

This is the load-bearing part, so it is worth slowing down for.

Suppose the domain error carried `httpStatus = 409`. Now consider a work order that cannot
move from `closed` back to `released`. That single failure must be:

- **HTTP 409 Conflict**, for a REST client
- **gRPC `FAILED_PRECONDITION` (9)**, for a gRPC client

One `httpStatus` field cannot express both. The gRPC adapter would have to *ignore* the
field and guess — which means the two adapters could disagree about what kind of error
occurred. So the domain error carries the semantic **category** (`CONFLICT`), and each
adapter derives its own spelling from it. That derivation is a `switch` inside each adapter
(Part 4 for HTTP, Part 5 for gRPC) — never a field on the category itself.

> Status is **derivable** from the category. The category is **not** derivable from the
> status. Encode what only the domain knows.

### Before and after, with real code

Here is a real failure in this repository. `OffsetPageRequest` validates the page size:

```java
// services/production-service/.../repository/pagination/OffsetPageRequest.java
if (pageSize < 1 || pageSize > MAX_PAGE_SIZE) {
    throw new IllegalArgumentException(
            "pageSize must be between 1 and %d, got %d".formatted(MAX_PAGE_SIZE, pageSize));
}
```

**Before** — send `GET /api/v1/work-orders?limit=500`. Nothing classifies that
`IllegalArgumentException`, so it escapes the controller as an unhandled exception. Here is
the **real captured response** from a running service:

```
$ curl -s -w "\nHTTP %{http_code}  %{content_type}\n" \
    "http://localhost:3001/api/v1/work-orders?limit=500"

{"timestamp":"2026-09-16T10:16:11.926Z","status":500,"error":"Internal Server Error","path":"/api/v1/work-orders"}
HTTP 500  application/json
```

The client is told **the server is broken**. It cannot tell what it did wrong, and it has no
machine-readable code to branch on. It will retry.

This is not an isolated case. Every parameter mistake behaves identically:

| Request | Real status |
|---|---|
| `?limit=500` | `500 Internal Server Error` |
| `?page=-1` | `500 Internal Server Error` |
| `?cursor=notavalidcursor` | `500 Internal Server Error` |
| `?cursor=v2.abc` | `500 Internal Server Error` |

Four different user errors, four identical server errors, no codes.

**After** you have implemented this tutorial, the same four requests produce four distinct
client errors. These are **real captured responses** from a service running this tutorial's
code — the `limit=500` case:

```http
HTTP/1.1 400 Bad Request
Content-Type: application/problem+json
```

```json
{
  "detail": "pageSize must be between 1 and 100, got 500",
  "instance": "/api/v1/work-orders",
  "status": 400,
  "title": "Page size out of range",
  "type": "https://factoryos.dev/errors/page-size-out-of-range",
  "code": "PAGE_SIZE_OUT_OF_RANGE",
  "retryable": false,
  "traceId": "7ae80c66-86ed-4ff9-b52b-114a2ae9e553",
  "timestamp": "2026-09-17T01:38:19.643097754Z",
  "errors": [
    {
      "field": "limit",
      "message": "must be between 1 and 100",
      "current": "500",
      "allowed": ["1..100"]
    }
  ]
}
```

And the other three, now **distinguishable**:

| Request | Code | Title |
|---|---|---|
| `?limit=500` | `PAGE_SIZE_OUT_OF_RANGE` | Page size out of range |
| `?page=-1` | `PAGE_SIZE_OUT_OF_RANGE` | Page size out of range |
| `?cursor=notavalidcursor` | `UNSUPPORTED_CURSOR_VERSION` | Unsupported cursor version |
| `?cursor=v2.abc` | `UNSUPPORTED_CURSOR_VERSION` | Unsupported cursor version |

Two further cursor cases get their own codes, with field-level detail:

| Request | Code |
|---|---|
| Cursor whose keys don't match the query's sort keys | `CURSOR_SORT_KEY_MISMATCH` |
| Cursor whose value can't be parsed as its key type | `INVALID_CURSOR` |

The mismatch case carries the diagnosis inline:

```json
"errors": [
  {
    "field": "cursor",
    "message": "cursor sort keys do not match the query",
    "current": "state,id",
    "allowed": ["createdAt", "id"]
  }
]
```

Now the client is told **it** sent something invalid (`400`), knows exactly what
(`PAGE_SIZE_OUT_OF_RANGE`), and knows retrying is pointless (`retryable: false`).

That is the whole point of the design. Everything that follows is machinery to make this
happen automatically, for every error, in every service.

### Check yourself

Run the Java test suite and confirm you have a green baseline before changing anything:

```bash
cd services && mvn -o -B test
```

Expected: `BUILD SUCCESS`, with every test green. If any test fails, fix it before moving on —
everything that follows assumes a green baseline.

> **If you see `Connection refused` to `localhost:5432`:** the
> `ProductionServiceApplicationTests` class is missing `@ActiveProfiles("test")`, so the
> context is falling back to `application.yml` and dialling a real PostgreSQL. That is an
> environment issue, not damage you caused. The fix is in commit `372a385`; if your checkout
> predates it, add the annotation. Part 6b explains the profile properly — and why
> `localhost` is the wrong hostname inside a devcontainer.

---


---

[← Part 0 — Before you start](00-before-you-start.md) · [Index](README.md) · [Part 2 — Maven ground floor →](02-maven-ground-floor.md)
