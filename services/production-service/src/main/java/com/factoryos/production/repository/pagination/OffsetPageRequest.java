package com.factoryos.production.repository.pagination;

import java.util.List;
import java.util.Objects;

import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Sort;

/**
 * Parses and validates request parameters for offset-based pagination.
 *
 * <p>Use this for UI tables that show page numbers (Page 1, 2, 3...).
 * For infinite scroll, use {@link CursorPageRequest} instead.
 */
public record OffsetPageRequest(int page, int pageSize, List<SortCriteria> sortCriteriaList) {

    private static final int DEFAULT_PAGE_SIZE = 20;
    private static final int MAX_PAGE_SIZE = 100;

    /**
     * Compact constructor with validation.
     *
     * @throws IllegalArgumentException if page < 0, pageSize out of range, or sortCriteriaList empty
     * @throws NullPointerException     if sortCriteriaList is null
     */
    public OffsetPageRequest {
        Objects.requireNonNull(sortCriteriaList, "sortCriteriaList must not be null");
        if (sortCriteriaList.isEmpty()) {
            throw new IllegalArgumentException("sortCriteriaList must not be empty");
        }
        if (page < 0) {
            throw new IllegalArgumentException("page must be >= 0, got %d".formatted(page));
        }
        if (pageSize < 1 || pageSize > MAX_PAGE_SIZE) {
            throw new IllegalArgumentException(
                    "pageSize must be between 1 and %d, got %d".formatted(MAX_PAGE_SIZE, pageSize));
        }
    }

    /**
     * Creates an OffsetPageRequest from raw request parameters.
     *
     * @param rawPage        the page number (0-indexed, negative becomes 0)
     * @param rawPageSize    the page size (0 = default, negative or >100 = error)
     * @param sortCriteriaList the sort criteria
     * @return a validated OffsetPageRequest
     */
    public static OffsetPageRequest of(int rawPage, int rawPageSize,
                                       List<SortCriteria> sortCriteriaList) {
        int pageSize = (rawPageSize == 0) ? DEFAULT_PAGE_SIZE : rawPageSize;
        return new OffsetPageRequest(rawPage, pageSize, sortCriteriaList);
    }

    /**
     * Converts to Spring Data Pageable.
     *
     * @return a Pageable for JPA queries
     */
    public Pageable toPageable() {
        Sort sort = toSort();
        return PageRequest.of(page, pageSize, sort);
    }

    private Sort toSort() {
        Sort result = null;
        for (SortCriteria criteria : sortCriteriaList) {
            Sort s = Sort.by(criteria.direction(), criteria.key().fieldName());
            result = (result == null) ? s : result.and(s);
        }
        return result;
    }
}
