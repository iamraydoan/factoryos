# Pagination Design Standard

> This document defines the pagination conventions for all FactoryOS APIs.

---

## Pagination Types

| Type | Use Case | Request | Response |
|---|---|---|---|
| **Page-based** | UI tables with page controls | `page`, `limit` | `page`, `limit`, `total`, `totalPages` |
| **Cursor-based** | Infinite scroll, mobile, large datasets | `limit`, `cursor` | `limit`, `nextCursor` |

---

## REST APIs

REST supports both pagination styles. Choose based on use case:

### Page-based (for UI tables)

**Request:**
```
GET /api/v1/work-orders?page=0&limit=20&state=released
```

**Response:**
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

### Cursor-based (for infinite scroll, mobile, large datasets)

**Request:**
```
GET /api/v1/work-orders?limit=20&cursor=v1.eyJ2IjoxLCJrZXlzIjpbImlkIl0sInZhbHMiOlsiNTUwZTg0MDAiXX0
```

**Response:**
```json
{
  "data": [ ... ],
  "pagination": {
    "limit": 20,
    "nextCursor": "v1.eyJ2IjoxLCJrZXlzIjpbImlkIl0sInZhbHMiOlsiNTUwZTg0MDAiXX0"
  }
}
```

> `nextCursor` is empty/absent when there are no more results.

### Field Reference

| Request | Type | Default | Description |
|---|---|---|---|
| `page` | int | 0 | Page number, 0-indexed (page-based only) |
| `limit` | int | 20 | Items per page (max 100) |
| `cursor` | string | — | Opaque cursor from previous response (cursor-based only) |

| Response (`pagination`) | Type | Description |
|---|---|---|
| `page` | int | Current page number (page-based only) |
| `limit` | int | Items per page used |
| `total` | int | Total matching records (page-based only) |
| `totalPages` | int | Total page count (page-based only) |
| `nextCursor` | string | Pass as `cursor` in next request; empty = last page (cursor-based only) |

---

## gRPC APIs

gRPC uses cursor-based pagination only (per Google API Design Guide).

### Request

```protobuf
message ListXxxRequest {
  // ... filters
  int32  pageSize  = 10;    // Items per page (default 20, max 100)
  string pageToken = 11;    // Opaque cursor (empty = first page)
}
```

### Response

```protobuf
message ListXxxResponse {
  repeated Xxx items = 1;
  string nextPageToken = 2;   // Empty = last page
}
```

### Rules

- `pageToken` is **opaque** — clients must not parse or construct it.
- `nextPageToken` is **empty** when there are no more results.
- `pageSize` defaults to **20**, max **100**. Server clamps out-of-range values.
- Default sort: **`id ASC`** unless specified otherwise.

---

## Cursor Design (Keyset Pagination)

The cursor encodes the **last item's sort key values**, not an offset.

### Why not offset?

```sql
-- Bad: scans and discards rows (slow on large tables)
SELECT * FROM work_orders ORDER BY id ASC LIMIT 20 OFFSET 10000;

-- Good: uses index directly
SELECT * FROM work_orders
WHERE id > 'a1b2c3d4-e5f6-7890-abcd-ef1234567890'
ORDER BY id ASC
LIMIT 20;
```

### Cursor Token Format

The cursor is a **versioned, opaque, URL-safe Base64** token.

**Format:** `v1.<base64url(json)>`

The version prefix (`v1.`) enables forward-compatible format changes.
Clients must not parse the token — it is opaque.

### Cursor JSON Structure

```json
{
  "keys": ["id"],
  "vals": ["a1b2c3d4-e5f6-7890-abcd-ef1234567890"]
}
```

| Field | Type | Description |
|---|---|---|
| `keys` | string[] | Sort key field names (must match query's sort keys) |
| `vals` | string[] | String-encoded values for each key (parallel to `keys`) |

> The version is carried by the **token prefix** (`v1.`), not inside the JSON.
> This allows the server to reject incompatible formats before parsing.

### Composite Key Cursors

For queries that sort by multiple fields, the cursor contains all sort key values:

**Sort:** `ORDER BY created_at DESC, id DESC`

**Cursor JSON:**
```json
{
  "keys": ["createdAt", "id"],
  "vals": ["2026-09-04T10:00:00Z", "a1b2c3d4-e5f6-7890-abcd-ef1234567890"]
}
```

**Generated WHERE clause:**
```sql
WHERE created_at < '2026-09-04T10:00:00Z'
   OR (created_at = '2026-09-04T10:00:00Z' AND id < 'a1b2c3d4-...')
ORDER BY created_at DESC, id DESC
LIMIT 21
```C

The last key (typically `id`) acts as a **deterministic tie-breaker**.

### Sort Key Types

| Type | Example Field | JSON Value |
|---|---|---|
| UUID | `id` | `"550e8400-e29b-41d4-a716-446655440000"` |
| Timestamp | `createdAt` | `"2026-09-04T10:00:00Z"` (ISO-8601) |
| String | `name` | `"Widget-A"` |
| Integer | `priority` | `"42"` |

All values are string-encoded in the cursor JSON, regardless of their native type.

### Validation

The server validates cursors on each request:

| Check | Error |
|---|---|
| Missing `v1.` prefix | `Invalid cursor format: missing version prefix` |
| Malformed Base64 | `Invalid cursor: malformed Base64` |
| Invalid JSON | `Invalid cursor: malformed JSON` |
| Wrong version | `Unsupported cursor version: ...` |
| Keys don't match query | `Cursor sort keys mismatch: expected [...], got [...]` |
| Values can't be parsed | `Invalid cursor value for key '...': ...` |

### limit + 1 Pattern

The server fetches `pageSize + 1` rows. If `> pageSize` rows return, the extra row is trimmed and a `nextCursor` is generated from the last visible row. This avoids `COUNT` queries.

```
Request: pageSize=20
Query:   LIMIT 21
Result:  21 rows → trim to 20, generate nextCursor from row #20
Result:  15 rows → return all 15, nextCursor is empty (last page)
```

---

## Naming Convention

| REST (JSON) | gRPC (proto) | Java | Go | Description |
|---|---|---|---|---|
| `page` | — | `page` | `Page` | Page number (REST page-based only) |
| `limit` | `page_size` | `pageSize` | `PageSize` | Items per page |
| `cursor` | `page_token` | `pageToken` | `PageToken` | Input cursor |
| `nextCursor` | `next_page_token` | `nextPageToken` | `NextPageToken` | Output cursor |
| `total` | — | `total` | `Total` | Total records (REST page-based only) |
| `totalPages` | — | `totalPages` | `TotalPages` | Total pages (REST page-based only) |

> **Rule:** REST JSON uses `camelCase`. gRPC proto uses `snake_case`. Java uses `camelCase`. Go uses `PascalCase` (exported).

---

## Java Implementation

Package: `com.factoryos.production.repository.pagination`

| Class | Purpose |
|---|---|
| `SortKey<V>` | Type-safe sort key definition (any entity, any field) |
| `Cursor` | Versioned token encode/decode/validation |
| `KeysetCondition` | Generates WHERE predicates from cursor values |
| `CursorPageRequest` | Cursor-based request: pageSize + pageToken → cursor |
| `CursorPage<T>` | Cursor-based result: items + nextPageToken |
| `OffsetPageRequest` | Page-based request: page + pageSize → Pageable |
| `OffsetPage<T>` | Page-based result: items + page + total + totalPages |

### Cursor-based Example

```java
List<SortKey<?>> sortKeys = List.of(SortKey.ofTimestamp("createdAt"), SortKey.id());
CursorPageRequest req = CursorPageRequest.of(rawPageSize, pageToken, sortKeys, Sort.Direction.DESC);

Specification<WorkOrder> spec = WorkOrderSpecs.withFiltersAndCursor(
    workCenterId, state, req.cursor(), req.direction());

List<WorkOrder> raw = repo.findAll(spec, req.toPageable()).getContent();
CursorPage<WorkOrder> page = CursorPage.of(raw, req.pageSize(), sortKeys, wo ->
    List.of(wo.getCreatedAt(), wo.getId()));
```

### Page-based Example

```java
List<SortKey<?>> sortKeys = List.of(SortKey.ofTimestamp("createdAt"), SortKey.id());
OffsetPageRequest req = OffsetPageRequest.of(rawPage, rawPageSize, sortKeys, Sort.Direction.DESC);

Specification<WorkOrder> spec = WorkOrderSpecs.withFilters(workCenterId, state);
org.springframework.data.domain.Page<WorkOrder> result = repo.findAll(spec, req.toPageable());
OffsetPage<WorkOrder> page = OffsetPage.from(result);
```
