package com.factoryos.production.repository.pagination;

import static org.junit.jupiter.api.Assertions.*;

import java.time.Instant;
import java.util.List;
import java.util.UUID;

import org.junit.jupiter.api.Test;

class CursorPageRequestTest {

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
        CursorPageRequest pr = CursorPageRequest.of(0, null, SORT_BY_ID_ASC);
        assertEquals(20, pr.pageSize());
    }

    @Test
    void of_negativePageSize_throwsIllegalArgument() {
        assertThrows(IllegalArgumentException.class,
            () -> CursorPageRequest.of(-5, null, SORT_BY_ID_ASC));
    }

    @Test
    void of_pageSizeOver100_throwsIllegalArgument() {
        assertThrows(IllegalArgumentException.class,
            () -> CursorPageRequest.of(150, null, SORT_BY_ID_ASC));
    }

    @Test
    void of_validPageSize_usesProvided() {
        CursorPageRequest pr = CursorPageRequest.of(50, null, SORT_BY_ID_ASC);
        assertEquals(50, pr.pageSize());
    }

    // ========================================================================
    // of(): Cursor Parsing
    // ========================================================================

    @Test
    void of_nullPageToken_firstPage() {
        CursorPageRequest pr = CursorPageRequest.of(20, null, SORT_BY_ID_ASC);
        assertNull(pr.cursor());
    }

    @Test
    void of_emptyPageToken_firstPage() {
        CursorPageRequest pr = CursorPageRequest.of(20, "", SORT_BY_ID_ASC);
        assertNull(pr.cursor());
    }

    @Test
    void of_validPageToken_decodesCursor() {
        UUID id = UUID.fromString("550e8400-e29b-41d4-a716-446655440000");
        String token = Cursor.of(List.of(SortKey.id()), List.of(id)).encode();

        CursorPageRequest pr = CursorPageRequest.of(20, token, SORT_BY_ID_ASC);

        assertNotNull(pr.cursor());
        assertEquals(id, pr.cursor().values().get(0));
    }

    @Test
    void of_invalidPageToken_throwsInvalidCursor() {
        assertThrows(InvalidCursorException.class,
            () -> CursorPageRequest.of(20, "invalid-token", SORT_BY_ID_ASC));
    }

    // ========================================================================
    // queryLimit()
    // ========================================================================

    @Test
    void queryLimit_addsOne() {
        CursorPageRequest pr = CursorPageRequest.of(20, null, SORT_BY_ID_ASC);
        assertEquals(21, pr.queryLimit());
    }

    // ========================================================================
    // toSort()
    // ========================================================================

    @Test
    void toSort_singleKeyAsc() {
        CursorPageRequest pr = CursorPageRequest.of(20, null, SORT_BY_ID_ASC);
        var sort = pr.toSort();
        assertTrue(sort.isSorted());
        assertEquals("id", sort.getOrderFor("id").getProperty());
    }

    // ========================================================================
    // toSort() — Mixed Directions
    // ========================================================================

    @Test
    void toSort_mixedDirections_containsAllOrders() {
        CursorPageRequest pr = CursorPageRequest.of(20, null, SORT_BY_CREATED_THEN_ID_DESC);
        var sort = pr.toSort();

        assertTrue(sort.isSorted());
        assertNotNull(sort.getOrderFor("createdAt"));
        assertNotNull(sort.getOrderFor("id"));
    }

    // ========================================================================
    // of(): Composite Cursor
    // ========================================================================

    @Test
    void of_compositeCursor_decodesCorrectly() {
        Instant ts = Instant.parse("2026-09-04T10:30:00Z");
        UUID id = UUID.fromString("550e8400-e29b-41d4-a716-446655440000");
        String token = Cursor.of(
            List.of(SortKey.ofTimestamp("createdAt"), SortKey.id()),
            List.of(ts, id)).encode();

        CursorPageRequest pr = CursorPageRequest.of(20, token, SORT_BY_CREATED_THEN_ID_DESC);

        assertNotNull(pr.cursor());
        assertEquals(ts, pr.cursor().values().get(0));
        assertEquals(id, pr.cursor().values().get(1));
    }

    @Test
    void of_pageSize1_valid() {
        CursorPageRequest pr = CursorPageRequest.of(1, null, SORT_BY_ID_ASC);
        assertEquals(1, pr.pageSize());
        assertEquals(2, pr.queryLimit());
    }

    @Test
    void of_pageSize100_valid() {
        CursorPageRequest pr = CursorPageRequest.of(100, null, SORT_BY_ID_ASC);
        assertEquals(100, pr.pageSize());
    }

    @Test
    void of_pageSize101_throwsIllegalArgument() {
        assertThrows(IllegalArgumentException.class,
            () -> CursorPageRequest.of(101, null, SORT_BY_ID_ASC));
    }

    // ========================================================================
    // Validation
    // ========================================================================

    @Test
    void of_emptySortCriteria_throwsIllegalArgument() {
        assertThrows(IllegalArgumentException.class,
            () -> CursorPageRequest.of(20, null, List.of()));
    }

    @Test
    void of_nullSortCriteria_throwsNullPointer() {
        assertThrows(NullPointerException.class,
            () -> CursorPageRequest.of(20, null, null));
    }
}
