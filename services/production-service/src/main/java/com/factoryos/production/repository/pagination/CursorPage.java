package com.factoryos.production.repository.pagination;

import java.util.List;
import java.util.Objects;
import java.util.function.Function;

/**
 * Wraps cursor-based pagination results using the limit+1 pattern.
 *
 * <p>
 * The limit+1 pattern detects if there is a next page without an expensive
 * {@code SELECT COUNT(*)}:
 * <ol>
 * <li>Query fetches {@code pageSize + 1} rows</li>
 * <li>If result has more than {@code pageSize} rows → there IS a next page</li>
 * <li>Trim to {@code pageSize} rows and generate cursor from last item</li>
 * </ol>
 *
 * @param <T> the entity type
 */
public record CursorPage<T>(List<T> items, String nextPageToken) {

    /**
     * Creates a CursorPage from raw query results.
     *
     * @param rawItems         the raw query results (may be pageSize + 1 rows)
     * @param pageSize         the requested page size
     * @param sortCriteriaList the sort criteria for cursor encoding
     * @param valueExtractor   a function that extracts sort key values from an
     *                         entity
     * @return a CursorPage with items and optional nextPageToken
     * @throws NullPointerException if any parameter is null
     */
    public static <T> CursorPage<T> of(List<T> rawItems, int pageSize,
            List<SortCriteria> sortCriteriaList,
            Function<T, List<Object>> valueExtractor) {
        Objects.requireNonNull(rawItems, "rawItems must not be null");
        Objects.requireNonNull(sortCriteriaList, "sortCriteriaList must not be null");
        Objects.requireNonNull(valueExtractor, "valueExtractor must not be null");

        // Last page (or empty): return all items, empty token
        if (rawItems.size() <= pageSize) {
            return new CursorPage<>(List.copyOf(rawItems), "");
        }

        // Has next page: trim the extra item
        List<T> pageItems = rawItems.subList(0, pageSize);
        T lastItem = pageItems.get(pageItems.size() - 1);

        // Extract sort key values from the last item using the provided function
        List<Object> cursorValues = valueExtractor.apply(lastItem);

        // Extract sort keys for cursor encoding (directions not needed)
        List<SortKey<?>> sortKeys = sortCriteriaList.stream()
                .<SortKey<?>>map(SortCriteria::key)
                .toList();

        // Encode as cursor token
        Cursor cursor = Cursor.of(sortKeys, cursorValues);

        return new CursorPage<>(List.copyOf(pageItems), cursor.encode());
    }
}
