# Collection Query Conventions

> How to filter, sort, search, and pick fields on FactoryOS REST collection endpoints.
>
> The OpenAPI spec in [`api/contracts/`](../../api/contracts/) is authoritative per
> endpoint. Companion standards: [PAGINATION_DESIGN.md](PAGINATION_DESIGN.md) and
> [ERROR_HANDLING.md](ERROR_HANDLING.md).

---

## Quick examples

```
# Filter
GET /api/v1/orders?status=active&customerId=abc-123
GET /api/v1/products?price[gte]=10&price[lte]=100
GET /api/v1/products?category=electronics,clothing
GET /api/v1/telemetry/latest?time[gte]=2026-08-10T00:00:00Z&time[lt]=2026-08-11T00:00:00Z

# Sort
GET /api/v1/products?sort=-createdAt
GET /api/v1/products?sort=-featured,price,-createdAt

# Search
GET /api/v1/products?q=wireless+headphones

# Fields
GET /api/v1/users?fields=id,name,email

# Paging (first page is 1)
GET /api/v1/work-orders?page=1&limit=20&state=released
```

---

## Filtering

| Form | Meaning | Example |
|---|---|---|
| `?field=value` | Equals | `?status=active` |
| `?field=a,b,c` | Matches any (comma-separated) | `?category=electronics,clothing` |
| `?field[ne]=x` | Not equal | `?status[ne]=closed` |
| `?field[gt]=x` / `?field[gte]=x` | Greater than / or equal | `?price[gte]=10` |
| `?field[lt]=x` / `?field[lte]=x` | Less than / or equal | `?price[lte]=100` |
| `?field[like]=frag` | Contains (strings, case-insensitive) | `?metricName[like]=torque` |

Same operators work for numbers and dates: `gte` includes the bound, `gt` excludes it.

Rules:

- Each endpoint lists which fields accept which operators — check its spec.
- Combine filters with `&` (all must match).
- Filters apply before paging: `total` counts the filtered rows.

---

## Sorting

```
GET /api/v1/products?sort=-createdAt
GET /api/v1/products?sort=-featured,price,-createdAt
```

- `-` prefix = descending, otherwise ascending.
- Comma-separated, applied left to right.
- Each endpoint lists its sortable fields and its default sort.

---

## Search

```
GET /api/v1/products?q=wireless+headphones
```

- `q` is a plain text search (case-insensitive substring, max 200 chars).
- Only endpoints with a text field offer `q` (e.g. alerts search `message`).
- `q` narrows results and combines with filters. For exact matches use filters (`?email=alice`).

---

## Fields

```
GET /api/v1/users?fields=id,name,email
GET /api/v1/orders?fields=id,total,status&include=customer.name
```

- `fields` returns only the listed properties; the rest are omitted.
- Each endpoint lists its selectable fields.
- `include` expands related objects — reserved for future endpoints; none accept it yet.

Naming: query params and properties are `camelCase` (`assetId`, `createdAt`).

---

## Paging with queries

- First page is `?page=1`. Page-based responses return `{ data, meta }` with `page`, `limit`, `total`.
- When following cursor pages, keep the same filters and `sort`.

---

## History endpoint

`GET /telemetry/history` is an aggregate query, not a full collection:

- Required: `assetId`, `metricName`, `from` (inclusive), `to` (exclusive). Optional: `bucket` (default `1h`).
- `sort` accepts `bucketTime` only; `fields` selects bucket properties.
- Very large ranges are rejected (`BUCKET_COUNT_OUT_OF_RANGE`, 400) — widen `bucket` and retry.

---

## Errors

Bad query input returns 400 with a `code`:

| Code | When |
|---|---|
| `INVALID_FILTER_FIELD` | Unknown filter field |
| `INVALID_FILTER_OPERATOR` | Operator not allowed on that field |
| `INVALID_SORT_FIELD` | Unknown sort field |
| `INVALID_FIELD_SELECTION` | Unknown `fields` entry |
| `BUCKET_COUNT_OUT_OF_RANGE` | History range yields too many buckets |
| `PAGE_NUMBER_OUT_OF_RANGE` | `page` below 1 |
| `PAGE_SIZE_OUT_OF_RANGE` | `limit` outside 1–100 |

---

## Endpoint author checklist

1. `$ref` the shared `Page`/`Limit`/`Cursor`/`Sort`/`Q`/`Fields` components (add `Q` only with a text field).
2. Declare explicit filter params for the allowlist (bracket names for operators).
3. Document in the operation description: filterable fields + operators, sortable fields + default, searchable fields (if `q`), selectable fields (if `fields`).
4. Return `{ data, meta }`, and state the error codes for violations.
