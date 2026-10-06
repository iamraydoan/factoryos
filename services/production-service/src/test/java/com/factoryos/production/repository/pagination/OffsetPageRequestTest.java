package com.factoryos.production.repository.pagination;

import static org.junit.jupiter.api.Assertions.*;

import java.util.List;

import org.junit.jupiter.api.Test;

import com.factoryos.common.error.CommonErrorCode;
import com.factoryos.common.error.DomainException;
import com.factoryos.common.error.ErrorCategory;

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
    void of_negativePageSize_throwsPageSizeOutOfRange() {
        DomainException ex = assertThrows(DomainException.class,
            () -> OffsetPageRequest.of(0, -5, SORT_BY_ID_ASC));
        assertEquals(CommonErrorCode.PAGE_SIZE_OUT_OF_RANGE, ex.code());
        assertEquals(ErrorCategory.VALIDATION, ex.code().category());
    }

    @Test
    void of_pageSizeOver100_throwsPageSizeOutOfRange() {
        DomainException ex = assertThrows(DomainException.class,
            () -> OffsetPageRequest.of(0, 150, SORT_BY_ID_ASC));
        assertEquals(CommonErrorCode.PAGE_SIZE_OUT_OF_RANGE, ex.code());
    }

    @Test
    void of_validPageSize_usesProvided() {
        OffsetPageRequest pr = OffsetPageRequest.of(0, 50, SORT_BY_ID_ASC);
        assertEquals(50, pr.pageSize());
    }

    @Test
    void of_negativePage_throwsPageSizeOutOfRange() {
        DomainException ex = assertThrows(DomainException.class,
            () -> OffsetPageRequest.of(-5, 20, SORT_BY_ID_ASC));
        assertEquals(CommonErrorCode.PAGE_SIZE_OUT_OF_RANGE, ex.code());
        assertEquals(ErrorCategory.VALIDATION, ex.code().category());
    }

    @Test
    void of_validPage_usesProvided() {
        OffsetPageRequest pr = OffsetPageRequest.of(3, 20, SORT_BY_ID_ASC);
        assertEquals(3, pr.page());
    }

    // ========================================================================
    // ofApiPage(): 1-indexed wire adapter
    // ========================================================================

    @Test
    void ofApiPage_firstPage_mapsToZeroIndexed() {
        OffsetPageRequest pr = OffsetPageRequest.ofApiPage(1, 20, SORT_BY_ID_ASC);
        assertEquals(0, pr.page());
    }

    @Test
    void ofApiPage_thirdPage_mapsToZeroIndexed() {
        OffsetPageRequest pr = OffsetPageRequest.ofApiPage(3, 20, SORT_BY_ID_ASC);
        assertEquals(2, pr.page());
    }

    @Test
    void ofApiPage_zero_throwsPageSizeOutOfRange() {
        DomainException ex = assertThrows(DomainException.class,
            () -> OffsetPageRequest.ofApiPage(0, 20, SORT_BY_ID_ASC));
        assertEquals(CommonErrorCode.PAGE_SIZE_OUT_OF_RANGE, ex.code());
        assertEquals(ErrorCategory.VALIDATION, ex.code().category());
    }

    @Test
    void ofApiPage_negative_throwsPageSizeOutOfRange() {
        DomainException ex = assertThrows(DomainException.class,
            () -> OffsetPageRequest.ofApiPage(-2, 20, SORT_BY_ID_ASC));
        assertEquals(CommonErrorCode.PAGE_SIZE_OUT_OF_RANGE, ex.code());
    }

    @Test
    void ofApiPage_zeroPageSize_usesDefault() {
        OffsetPageRequest pr = OffsetPageRequest.ofApiPage(1, 0, SORT_BY_ID_ASC);
        assertEquals(20, pr.pageSize());
    }

    @Test
    void ofApiPage_toPageable_usesZeroIndexedPageNumber() {
        OffsetPageRequest pr = OffsetPageRequest.ofApiPage(2, 25, SORT_BY_CREATED_THEN_ID_DESC);
        assertEquals(1, pr.toPageable().getPageNumber());
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

    // ========================================================================
    // Classification: a bad page size is the caller's mistake, not a 5xx
    // ========================================================================

    @Test
    void badPageSize_isClientError_notServerError() {
        DomainException ex = assertThrows(DomainException.class,
            () -> OffsetPageRequest.of(0, 500, SORT_BY_ID_ASC));
        // The category is the client-error class; the HTTP status itself (400, not a
        // 5xx) is the adapter's projection and is asserted in Part 7.
        assertEquals(ErrorCategory.VALIDATION, ex.code().category(),
            "an out-of-range page size is the caller's mistake, so must not be a 5xx");
        assertFalse(ex.code().retryable(), "retrying the identical request cannot succeed");
    }

    @Test
    void badPageSize_carriesFieldLevelDetail() {
        DomainException ex = assertThrows(DomainException.class,
            () -> OffsetPageRequest.of(0, 500, SORT_BY_ID_ASC));
        assertEquals(1, ex.fieldErrors().size());
        assertEquals("limit", ex.fieldErrors().get(0).field());
        assertEquals("500", ex.fieldErrors().get(0).current());
    }
}
