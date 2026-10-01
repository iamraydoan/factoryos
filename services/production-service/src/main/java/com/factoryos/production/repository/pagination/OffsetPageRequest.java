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
     * Creates an OffsetPageRequest from 0-indexed internal page numbers.
     *
     * <p>Internal callers (tests, non-REST code) use this. REST controllers must use
     * {@link #ofApiPage(int, int, List)} instead, which accepts the 1-indexed
     * wire page number.
     *
     * @param rawPage        the page number (0-indexed, negative = error)
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
     * Creates an OffsetPageRequest from the 1-indexed API page number.
     *
     * <p>This is the REST boundary adapter: the wire speaks 1-indexed pages
     * ({@code ?page=1} is the first page) while Spring Data stays 0-indexed
     * internally. Page numbers below 1 are a client error
     * ({@code PAGE_NUMBER_OUT_OF_RANGE}, 400).
     *
     * @param apiPage        the wire page number (1-indexed, below 1 = error)
     * @param rawPageSize    the page size (0 = default, negative or >100 = error)
     * @param sortCriteriaList the sort criteria
     * @return a validated OffsetPageRequest with the 0-indexed page
     */
    public static OffsetPageRequest ofApiPage(int apiPage, int rawPageSize,
                                              List<SortCriteria> sortCriteriaList) {
        if (apiPage < 1) {
            throw new IllegalArgumentException(
                    "page must be >= 1, got %d".formatted(apiPage));
        }
        int pageSize = (rawPageSize == 0) ? DEFAULT_PAGE_SIZE : rawPageSize;
        return new OffsetPageRequest(apiPage - 1, pageSize, sortCriteriaList);
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

    /**
     * Builds a Spring Data {@link Sort} object from the sort criteria.
     *
     * <p>Supports mixed directions:
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
