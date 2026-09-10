package com.factoryos.production.repository.pagination;

import static org.junit.jupiter.api.Assertions.*;

import java.util.List;

import org.junit.jupiter.api.Test;

class OffsetPageRequestTest {

    private static final List<SortCriteria> SORT_BY_ID_ASC = List.of(
            SortCriteria.asc(SortKey.id()));
    private static final List<SortCriteria> SORT_BY_CREATED_THEN_ID_DESC = List.of(
            SortCriteria.desc(SortKey.ofTimestamp("createdAt")),
            SortCriteria.desc(SortKey.id()));

    // ========================================================================
    // of(): Default Values
    // ========================================================================

    @Test
    void of_zeroPageSize_usesDefault() {
        OffsetPageRequest pr = OffsetPageRequest.of(0, 0, SORT_BY_ID_ASC);
        assertEquals(20, pr.pageSize());
    }

    @Test
    void of_negativePageSize_throwsIllegalArgument() {
        assertThrows(IllegalArgumentException.class,
            () -> OffsetPageRequest.of(0, -5, SORT_BY_ID_ASC));
    }

    @Test
    void of_pageSizeOver100_throwsIllegalArgument() {
        assertThrows(IllegalArgumentException.class,
            () -> OffsetPageRequest.of(0, 150, SORT_BY_ID_ASC));
    }

    @Test
    void of_validPageSize_usesProvided() {
        OffsetPageRequest pr = OffsetPageRequest.of(0, 50, SORT_BY_ID_ASC);
        assertEquals(50, pr.pageSize());
    }

    @Test
    void of_negativePage_throwsIllegalArgument() {
        assertThrows(IllegalArgumentException.class,
            () -> OffsetPageRequest.of(-5, 20, SORT_BY_ID_ASC));
    }

    @Test
    void of_validPage_usesProvided() {
        OffsetPageRequest pr = OffsetPageRequest.of(3, 20, SORT_BY_ID_ASC);
        assertEquals(3, pr.page());
    }

    // ========================================================================
    // toPageable()
    // ========================================================================

    @Test
    void toPageable_withSortCriteria() {
        OffsetPageRequest pr = OffsetPageRequest.of(2, 25, SORT_BY_CREATED_THEN_ID_DESC);
        var pageable = pr.toPageable();

        assertEquals(2, pageable.getPageNumber());
        assertEquals(25, pageable.getPageSize());
        assertTrue(pageable.getSort().isSorted());
    }

    @Test
    void toSort_mixedDirections_containsAllOrders() {
        List<SortCriteria> mixed = List.of(
            SortCriteria.desc(SortKey.ofTimestamp("createdAt")),
            SortCriteria.asc(SortKey.ofString("state")),
            SortCriteria.desc(SortKey.id()));
        OffsetPageRequest pr = OffsetPageRequest.of(0, 20, mixed);
        var sort = pr.toSort();

        assertTrue(sort.isSorted());
        assertNotNull(sort.getOrderFor("createdAt"));
        assertNotNull(sort.getOrderFor("state"));
        assertNotNull(sort.getOrderFor("id"));
    }

    // ========================================================================
    // of(): Edge Cases
    // ========================================================================

    @Test
    void of_pageSize1_valid() {
        OffsetPageRequest pr = OffsetPageRequest.of(0, 1, SORT_BY_ID_ASC);
        assertEquals(1, pr.pageSize());
    }
}
