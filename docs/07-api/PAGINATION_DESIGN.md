# Pagination Design Standard

> This document defines the pagination conventions for all FactoryOS APIs.

---

## REST APIs

REST supports two pagination styles. Choose based on use case:

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
GET /api/v1/work-orders?limit=20&cursor=g6tjMjAyNi0wOC0yN1QxMDowMzowMFo
```

**Response:**
```json
{
  "data": [ ... ],
  "pagination": {
    "limit": 20,
    "nextCursor": "g6tjMjAyNi0wOC0yN1QxMDowMzowMFo"
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

### Cursor format

The cursor is a **base64url-encoded JSON object** containing the last item's sort key fields:

```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
```

### Sort key

- Default sort: `id ASC`
- Use composite cursor: `id > lastId`

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
