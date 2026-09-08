package com.factoryos.production.repository.pagination;

import java.util.List;
import java.util.Objects;

import org.springframework.data.domain.Sort;

/**
 * Parses and validates gRPC request parameters for cursor-based pagination.
 *
 * <p>
 * Takes raw request parameters ({@code pageSize}, {@code pageToken}) and
 * produces:
 * <ul>
 * <li>A validated page size (default 20, max 100)</li>
 * <li>A decoded {@link Cursor} (or {@code null} for first page)</li>
 * <li>A Spring Data {@link Sort} for the query</li>
 * </ul>
 *
 * <p>
 * This class is immutable and thread-safe.
 */
public record CursorPageRequest(int pageSize, Cursor cursor, List<SortCriteria> sortCriteriaList) {

    private static final int DEFAULT_PAGE_SIZE = 20;
    private static final int MAX_PAGE_SIZE = 100;

    /**
     * Compact constructor with validation.
     *
     * @throws IllegalArgumentException if pageSize is out of range or
     *                                  sortCriteriaList is empty
     * @throws NullPointerException     if sortCriteriaList is null
     */
    public CursorPageRequest {
        Objects.requireNonNull(sortCriteriaList, "sortCriteriaList must not be null");
        if (sortCriteriaList.isEmpty()) {
            throw new IllegalArgumentException("sortCriteriaList must not be empty");
        }
        if (pageSize < 1 || pageSize > MAX_PAGE_SIZE) {
            throw new IllegalArgumentException(
                    "pageSize must be between 1 and %d, got %d".formatted(MAX_PAGE_SIZE, pageSize));
        }
    }

    /**
     * Creates a CursorPageRequest from raw gRPC request parameters.
     *
     * @param rawPageSize      the raw page size (0 = use default, negative or >100
     *                         = error)
     * @param pageToken        the page token (null or empty = first page)
     * @param sortCriteriaList the sort criteria for the query
     * @return a validated CursorPageRequest
     * @throws InvalidCursorException   if pageToken is malformed
     * @throws IllegalArgumentException if rawPageSize is negative or >100
     */
    public static CursorPageRequest of(int rawPageSize, String pageToken,
            List<SortCriteria> sortCriteriaList) {
        // 0 (not specified) → default, negative or > max → error
        int pageSize = (rawPageSize == 0) ? DEFAULT_PAGE_SIZE : rawPageSize;
        // Extract sort keys for cursor decoding
        List<SortKey<?>> sortKeys = sortCriteriaList.stream()
                .<SortKey<?>>map(SortCriteria::key)
                .toList();
        // Decode cursor: null/empty → null (first page), invalid → throws
        // InvalidCursorException
        Cursor cursor = Cursor.decode(pageToken, sortKeys);
        return new CursorPageRequest(pageSize, cursor, sortCriteriaList);
    }

    /**
     * Returns the query limit (pageSize + 1) to detect if there is a next page.
     *
     * @return the query limit
     */
    public int queryLimit() {
        return pageSize + 1;
    }

    /**
     * Builds a Spring Data {@link Sort} object from the sort criteria.
     *
     * <p>
     * Supports mixed directions:
     * {@code Sort.by(DESC, "createdAt").and(Sort.by(ASC, "name")).and(Sort.by(DESC, "id"))}
     *
     * @return the Sort object
     */
    public Sort toSort() {
        Sort result = null;
        for (SortCriteria criteria : sortCriteriaList) {
            Sort s = Sort.by(criteria.direction(), criteria.key().fieldName());
            result = (result == null) ? s : result.and(s);
        }
        return result;
    }
}
