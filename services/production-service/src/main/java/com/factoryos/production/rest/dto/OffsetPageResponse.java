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
 *   "meta": { "page": 1, "limit": 20, "total": 150 }
 * }
 * </pre>
 *
 * @param <T> the item type
 */
public record OffsetPageResponse<T>(List<T> data, Meta meta) {

    /**
     * Pagination metadata for page-based responses.
     *
     * @param page  current page number (1-indexed on the wire)
     * @param limit items per page
     * @param total total matching records
     */
    public record Meta(int page, int limit, long total) {
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
     * <p>The outbound half of the index-base adapter: the internal 0-indexed
     * page becomes the 1-indexed wire page ({@code ?page=1} is the first page).
     *
     * @param page the offset page result (0-indexed)
     * @return a response matching the REST pagination design spec
     * @param <T> the item type
     */
    public static <T> OffsetPageResponse<T> from(OffsetPage<T> page) {
        return new OffsetPageResponse<>(page.items(),
                new Meta(page.page() + 1, page.pageSize(), page.total()));
    }
}
