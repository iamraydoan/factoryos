# Pagination Design Standard

> This document defines the pagination conventions for all FactoryOS APIs.
>
> **Language-neutral.** It specifies wire formats, token formats, and implementation
> requirements — not code. The contract is identical regardless of the stack a
> service is written in, so no language is shown as the reference.
>
> The authoritative contract for any single API is its schema in
> [`api/contracts/`](../../api/contracts/): the OpenAPI spec for REST endpoints and
> the `.proto` for gRPC services.

---

## 1. Pagination Types

| Type | Use Case | Request | Response |
|---|---|---|---|
| **Page-based** | UI tables with page controls | `page`, `limit` | `page`, `limit`, `total`, `totalPages` |
| **Cursor-based** | Infinite scroll, mobile, large datasets | `limit`, `cursor` | `limit`, `nextCursor` |

| Style | Requires `COUNT` | Stable under concurrent writes | Use when |
|---|---|---|---|
| Page-based | yes | no — rows shift between pages | The UI shows page numbers and a total |
| Cursor-based | no | yes — the cursor anchors to a row | The client scrolls, or the dataset is large |

**gRPC is always cursor-based** (per the Google API Design Guide); there is no page-based variant. REST supports both.

---

## 2. REST API

### 2.1 Page-based (UI tables)

**Request**

```
GET /api/v1/work-orders?page=0&limit=20&state=released
```

**Response** — `200 OK`, `application/json`:

```json
{
  "data": [ ... ],
  "pagination": {
    "page": 0,
    "limit": 20,
    "total": 150,
    "totalPages": 8
  }
}
```

### 2.2 Cursor-based (infinite scroll, mobile, large datasets)

**Request**

```
GET /api/v1/work-orders?limit=20&cursor=v1.eyJ2IjoxLCJrZXlzIjpbImlkIl0sInZhbHMiOlsiNTUwZTg0MDAiXX0
```

**Response** — `200 OK`, `application/json`:

```json
{
  "data": [ ... ],
  "pagination": {
    "limit": 20,
    "nextCursor": "v1.eyJ2IjoxLCJrZXlzIjpbImlkIl0sInZhbHMiOlsiNTUwZTg0MDAiXX0"
  }
}
```

> `nextCursor` is **empty or absent on the last page**. Clients stop when it is empty.

### 2.3 Field Reference

| Request | Type | Default | Description |
|---|---|---|---|
| `page` | int | 0 | Page number, 0-indexed (page-based only) |
| `limit` | int | 20 | Items per page (max 100) |
| `cursor` | string | — | Opaque cursor from the previous response (cursor-based only) |

| Response (`pagination`) | Type | Description |
|---|---|---|
| `page` | int | Current page number (page-based only) |
| `limit` | int | Items per page used |
| `total` | int | Total matching records (page-based only) |
| `totalPages` | int | Total page count (page-based only) |
| `nextCursor` | string | Pass as `cursor` in the next request; empty = last page (cursor-based only) |

### 2.4 Rules

- The response envelope is **always** `{ data, pagination }`. Both styles share it; the `pagination` members differ.
- `data` is an array even when empty — never `null`.
- Field names are `camelCase` (see §4).
- Invalid pagination input returns **400** with an error `code`; see §5.4 and [ERROR_HANDLING.md](ERROR_HANDLING.md).

---

## 3. gRPC API

gRPC uses cursor-based pagination only.

### 3.1 Request fields

Every `ListXxxRequest` carries the endpoint's own filters plus these two fields:

| Field | Type | Presence | Description |
|---|---|---|---|
| `page_size` | int32 | optional | Items per page. Unset/0 = default **20**; max **100** |
| `page_token` | string | optional | Opaque cursor from the previous response. Empty = first page |

### 3.2 Response fields

| Field | Type | Description |
|---|---|---|
| `items` | repeated `Xxx` | The page of results |
| `next_page_token` | string | Pass as `page_token` in the next request. **Empty = last page** |

Because the wire type is fixed (`repeated`), the query's rows are mapped onto the entity message — e.g. `repeated WorkOrder work_orders`. Only the token field name is fixed by this standard.

### 3.3 Rules

- `page_token` is **opaque** — clients must not parse, construct, or persist it beyond the next request.
- `next_page_token` is **empty** when there are no more results.
- `page_size` unset or `0` means the default **20**. A value outside `1..100` is **rejected** with `INVALID_ARGUMENT` — the server does not silently clamp, because a client asking for 5000 rows has a bug worth surfacing.
- Default sort is the endpoint's stable ordering. Where none is specified, use `id ASC`.

---

## 4. Naming Convention

The same concept has a different name in each protocol and language. This table is the contract:

| REST (JSON) | gRPC (proto) | Java | Go | Description |
|---|---|---|---|---|
| `page` | — | `page` | `Page` | Page number (REST page-based only) |
| `limit` | `page_size` | `pageSize` | `PageSize` | Items per page |
| `cursor` | `page_token` | `pageToken` | `PageToken` | Input cursor |
| `nextCursor` | `next_page_token` | `nextPageToken` | `NextPageToken` | Output cursor |
| `total` | — | `total` | `Total` | Total records (REST page-based only) |
| `totalPages` | — | `totalPages` | `TotalPages` | Total pages (REST page-based only) |

> **Rule:** REST JSON uses `camelCase`. gRPC proto uses `snake_case`. Java uses `camelCase`. Go uses `PascalCase` (exported).
>
> Other languages follow their own idiom (`snake_case` in Python, `camelCase` in TypeScript) — the **wire** names above are what matter.

---

## 5. Cursor Design (Keyset Pagination)

The cursor encodes the **last item's sort key values**, not an offset.

### 5.1 Why not offset?

An offset query must scan and discard every row before the offset. On a large table that cost grows with page depth, and the result is unstable: a row inserted or deleted mid-scroll shifts every subsequent page, so a client can silently skip or duplicate records.

A keyset query anchors to an indexed row value instead. It seeks directly into the index — cost is constant regardless of depth — and because it anchors to a row rather than a position, concurrent writes do not shift the page.

This is why cursor-based pagination is required for large or actively-written datasets, and why the composite index in §5.6 matters.

### 5.2 Cursor Token Format

The cursor is a **versioned, opaque, URL-safe Base64** token.

**Format:** `v1.<base64url(json)>`

- The version prefix (`v1.`) enables forward-compatible format changes: a server can reject an incompatible token **before** parsing it.
- Clients must treat the token as opaque. Its internal structure is not part of the API contract and may change.
- URL-safe Base64 (no padding) so the token needs no percent-encoding in a query string.

### 5.3 Cursor Payload

The decoded payload carries the sort key names and their string-encoded values:

```json
{
  "keys": ["id"],
  "vals": ["a1b2c3d4-e5f6-7890-abcd-ef1234567890"]
}
```

| Field | Type | Description |
|---|---|---|
| `keys` | string[] | Sort key field names. **Must match the query's sort keys** |
| `vals` | string[] | String-encoded values, parallel to `keys` |

> The version lives in the **token prefix**, not in the payload — so an incompatible
> version is rejected before any parsing happens.

### 5.4 Sort Key Types

| Type | Example Field | Encoded Value |
|---|---|---|
| UUID | `id` | `"550e8400-e29b-41d4-a716-446655440000"` |
| Timestamp | `createdAt` | `"2026-09-04T10:00:00Z"` (ISO-8601, UTC) |
| String | `name` | `"Widget-A"` |
| Integer | `priority` | `"42"` |

All values are **string-encoded in the payload**, regardless of their native type. Each sort key definition therefore needs a parse function that converts its encoded string back to the native type.

### 5.5 Validation

The server validates every cursor before use. Each failure maps to the error taxonomy in [ERROR_HANDLING.md](ERROR_HANDLING.md):

| Check | Error code | HTTP | gRPC |
|---|---|---|---|
| Missing `v1.` prefix | `UNSUPPORTED_CURSOR_VERSION` | 400 | `INVALID_ARGUMENT` |
| Wrong version | `UNSUPPORTED_CURSOR_VERSION` | 400 | `INVALID_ARGUMENT` |
| Malformed Base64 or JSON | `INVALID_CURSOR` | 400 | `INVALID_ARGUMENT` |
| Keys don't match the query's sort keys | `CURSOR_SORT_KEY_MISMATCH` | 400 | `INVALID_ARGUMENT` |
| A value can't be parsed as its key type | `INVALID_CURSOR` | 400 | `INVALID_ARGUMENT` |
| `limit`/`page_size` outside `1..100` | `PAGE_SIZE_OUT_OF_RANGE` | 400 | `INVALID_ARGUMENT` |

A rejected cursor is **always a client error (400)**, never a 500 — including a cursor that is well-formed but was issued for a different query.

### 5.6 Composite Keys

For queries sorted by multiple fields, the cursor carries every sort key value.

**Sort order:** `createdAt DESC, id DESC`

**Cursor payload:**

```json
{
  "keys": ["createdAt", "id"],
  "vals": ["2026-09-04T10:00:00Z", "a1b2c3d4-e5f6-7890-abcd-ef1234567890"]
}
```

**Predicate the server builds:**

| Clause | Purpose |
|---|---|
| `createdAt` strictly less than the cursor's `createdAt` | Rows after the cursor by the primary key |
| **OR** (`createdAt` equal **AND** `id` strictly less than the cursor's `id`) | Same timestamp, broken by the tie-breaker |
| `ORDER BY createdAt DESC, id DESC` | Must match the cursor's key order exactly |
| `LIMIT page_size + 1` | The limit+1 probe (§5.7) |

The **last key is the deterministic tie-breaker** and should be unique — almost always `id`. Without it, rows sharing a timestamp have no defined order and the cursor can skip or repeat them.

Compare using `strictly less` for `DESC` and `strictly greater` for `ASC` — the direction of every key in the predicate must match the direction of the sort.

### 5.7 The limit + 1 Pattern

The server fetches **`page_size + 1`** rows. The extra row is a probe: its presence proves another page exists, without a `COUNT` query.

| Rows returned | Interpretation | Action |
|---|---|---|
| `page_size + 1` | More results exist | Trim to `page_size`; emit `nextCursor` from the **last kept** row |
| `≤ page_size` | Last page | Return all rows; `nextCursor` is empty |

Two requirements follow:

- **Always fetch `page_size + 1`.** Fetching `page_size` truncates the last full page and never emits a cursor, so the client believes it has reached the end.
- **Generate the cursor from the last row you keep**, not the trimmed probe row. Using the probe row skips one record per page.

### 5.8 Index Requirements

The keyset predicate must be backed by a composite index **in the sort order** — for `ORDER BY createdAt DESC, id DESC`, an index on `(created_at, id)`. Without it the database cannot seek and falls back to scanning, which forfeits the entire benefit of keyset pagination.

---

## 6. Implementation Requirements

These obligations apply to any implementation, in any language. They are the parts that are easy to get subtly wrong.

### 6.1 Reusable, not per-entity

Pagination must be a **single reusable component per language**, not reimplemented per endpoint. Entity-specific code supplies only:

- the **filters** for the query,
- the **sort keys** (field name, direction, and a parser for decoding), and
- a way to **extract the sort key values** from a result row.

Everything else — token encoding/decoding, validation, the limit+1 trim, cursor generation — belongs to the shared component. Duplicating it per entity is how the two hint fields below drift.

### 6.2 Sort keys are an ordered list of (field, direction)

A query's sort must be expressible as a **list** of key/direction pairs, not a single direction applied to all keys. `createdAt DESC, id DESC` requires it, and so does any mixed order such as `state ASC, createdAt DESC`.

### 6.3 The value extractor must match the sort key order

The row-to-values function must emit values in **exactly** the order of the sort keys. The pairing is positional, so a mismatch produces a cursor that decodes to the wrong position — it does not raise an error, it silently returns wrong results. This is the single most likely bug in a new implementation.

### 6.4 Sort definitions and extractors must live together

Keep a query's sort order and its value extractor in one place — a module-level constant per query — so they cannot drift. Defining the sort in one file and the extractor in another is how §6.3 happens.

### 6.5 Filters and pagination are separate concerns

Filtering (which rows) and pagination (which page of them) must be composable independently, so a query can use either style with the same filters. Do not bake a filter into the pagination component.

### 6.6 Never trust a cursor

Treat every cursor as hostile input: validate the prefix, decode defensively, verify the key names match the query, and parse each value through its key's parser. A cursor is client-supplied state — a well-formed token from a *different* query is still a client error, not a server fault.

---

## 7. References

- [ERROR_HANDLING.md](ERROR_HANDLING.md) — error codes and the transport mapping for §5.5
- [api/contracts/](../../api/contracts/) — the OpenAPI and Protobuf contracts that bind each endpoint to this standard
- [Google API Design Guide — Pagination](https://cloud.google.com/apis/design/design_patterns#list_pagination)
