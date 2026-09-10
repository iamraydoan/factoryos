package com.factoryos.production.repository.pagination;

import java.util.List;
import java.util.Objects;

import org.springframework.data.domain.Page;

/**
 * Wraps offset-based pagination results for UI tables with page numbers.
 *
 * <p>Use this for "Page 1 of 50" style pagination.
 * For infinite scroll, use {@link CursorPage} instead.
 *
 * @param <T> the entity type
 */
public record OffsetPage<T>(List<T> items, int page, int pageSize, long total, int totalPages) {

    /**
     * Compact constructor with validation.
     *
     * @throws IllegalArgumentException if page < 0, pageSize < 1, or total < 0
     * @throws NullPointerException     if items is null
     */
    public OffsetPage {
        Objects.requireNonNull(items, "items must not be null");
        if (page < 0) {
            throw new IllegalArgumentException("page must be >= 0, got %d".formatted(page));
        }
        if (pageSize < 1) {
            throw new IllegalArgumentException("pageSize must be >= 1, got %d".formatted(pageSize));
        }
        if (total < 0) {
            throw new IllegalArgumentException("total must be >= 0, got %d".formatted(total));
        }
    }

    /**
     * Creates an OffsetPage from Spring Data's Page result.
     *
     * @param springPage the Spring Data Page
     * @return an OffsetPage wrapper
     * @throws NullPointerException if springPage is null
     */
    public static <T> OffsetPage<T> from(Page<T> springPage) {
        Objects.requireNonNull(springPage, "springPage must not be null");
        return new OffsetPage<>(
                springPage.getContent(),
                springPage.getNumber(),
                springPage.getSize(),
                springPage.getTotalElements(),
                springPage.getTotalPages()
        );
    }

    /**
     * Returns true if there is a next page.
     *
     * @return true if not on the last page
     */
    public boolean hasNext() {
        return page < totalPages - 1;
    }

    /**
     * Returns true if there is a previous page.
     *
     * @return true if not on the first page
     */
    public boolean hasPrevious() {
        return page > 0;
    }
}
