package com.factoryos.production.rest.dto;

import java.util.List;
import java.util.function.Function;

import com.factoryos.production.repository.pagination.CursorPage;
import com.factoryos.production.repository.pagination.SortCriteria;

/**
 * REST response envelope for cursor-based (keyset) pagination.
 *
 * <p>JSON shape matches {@code PAGINATION_DESIGN.md}:
 * <pre>
 * {
 *   "data": [ ... ],
 *   "pagination": { "limit": 20, "nextCursor": "v1.eyJ..." }
 * }
 * </pre>
 *
 * <p>{@code nextCursor} is empty when there are no more results.
 *
 * @param <T> the item type
 */
public record CursorPageResponse<T>(List<T> data, Pagination pagination) {

    /**
     * Pagination metadata for cursor-based responses.
     *
     * @param limit      items per page used
     * @param nextCursor pass as {@code cursor} in the next request; empty = last page
     */
    public record Pagination(int limit, String nextCursor) {
    }

    /**
     * Creates a response directly from raw query results (limit+1 pattern).
     *
     * <p>Preferred in controllers — skips the intermediate {@link CursorPage}.
     * Internally applies the limit+1 trimming and cursor generation.
     *
     * @param rawItems         the raw query results (may be pageSize + 1 rows)
     * @param pageSize         the requested page size
     * @param sortCriteriaList the sort criteria for cursor encoding
     * @param valueExtractor   a function that extracts sort key values from an entity
     * @return a response matching the REST pagination design spec
     * @param <T> the item type
     */
    public static <T> CursorPageResponse<T> of(List<T> rawItems, int pageSize,
            List<SortCriteria> sortCriteriaList,
            Function<T, List<Object>> valueExtractor) {
        return from(CursorPage.of(rawItems, pageSize, sortCriteriaList, valueExtractor), pageSize);
    }

    /**
     * Wraps a {@link CursorPage} into the REST response envelope.
     *
     * @param page     the cursor page result
     * @param pageSize the page size that was used for this request
     * @return a response matching the REST pagination design spec
     * @param <T> the item type
     */
    public static <T> CursorPageResponse<T> from(CursorPage<T> page, int pageSize) {
        return new CursorPageResponse<>(page.items(), new Pagination(pageSize, page.nextPageToken()));
    }
}
