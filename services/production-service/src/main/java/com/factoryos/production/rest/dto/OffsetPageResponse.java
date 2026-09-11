package com.factoryos.production.rest.dto;

import java.util.List;

import org.springframework.data.domain.Page;

import com.factoryos.production.repository.pagination.OffsetPage;

/**
 * REST response envelope for offset-based (page-number) pagination.
 *
 * <p>JSON shape matches {@code PAGINATION_DESIGN.md}:
 * <pre>
 * {
 *   "data": [ ... ],
 *   "pagination": { "page": 0, "limit": 20, "total": 150, "totalPages": 8 }
 * }
 * </pre>
 *
 * @param <T> the item type
 */
public record OffsetPageResponse<T>(List<T> data, Pagination pagination) {

    /**
     * Pagination metadata for page-based responses.
     *
     * @param page       current page number (0-indexed)
     * @param limit      items per page
     * @param total      total matching records
     * @param totalPages total page count
     */
    public record Pagination(int page, int limit, long total, int totalPages) {
    }

    /**
     * Creates a response directly from a Spring Data {@link Page}.
     *
     * <p>Preferred in controllers — skips the intermediate {@link OffsetPage}.
     *
     * @param springPage the Spring Data Page result
     * @return a response matching the REST pagination design spec
     * @param <T> the item type
     */
    public static <T> OffsetPageResponse<T> of(Page<T> springPage) {
        return from(OffsetPage.from(springPage));
    }

    /**
     * Wraps an {@link OffsetPage} into the REST response envelope.
     *
     * @param page the offset page result
     * @return a response matching the REST pagination design spec
     * @param <T> the item type
     */
    public static <T> OffsetPageResponse<T> from(OffsetPage<T> page) {
        return new OffsetPageResponse<>(page.items(),
                new Pagination(page.page(), page.pageSize(), page.total(), page.totalPages()));
    }
}
