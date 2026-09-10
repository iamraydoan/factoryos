package com.factoryos.production.repository.pagination;

import static org.junit.jupiter.api.Assertions.*;

import java.util.List;

import org.junit.jupiter.api.Test;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageImpl;
import org.springframework.data.domain.PageRequest;

class OffsetPageTest {

    // ========================================================================
    // Construction
    // ========================================================================

    @Test
    void constructor_validParams_succeeds() {
        OffsetPage<String> page = new OffsetPage<>(List.of("a", "b"), 0, 20, 100, 5);
        assertEquals(List.of("a", "b"), page.items());
        assertEquals(0, page.page());
        assertEquals(20, page.pageSize());
        assertEquals(100, page.total());
        assertEquals(5, page.totalPages());
    }

    @Test
    void from_springPage_mapsAllFields() {
        Page<String> springPage = new PageImpl<>(
            List.of("a", "b"),
            PageRequest.of(2, 10),
            55
        );

        OffsetPage<String> page = OffsetPage.from(springPage);

        assertEquals(List.of("a", "b"), page.items());
        assertEquals(2, page.page());
        assertEquals(10, page.pageSize());
        assertEquals(55, page.total());
        assertEquals(6, page.totalPages());
    }

    @Test
    void from_emptySpringPage_mapsCorrectly() {
        Page<String> springPage = new PageImpl<>(
            List.of(),
            PageRequest.of(0, 20),
            0
        );

        OffsetPage<String> page = OffsetPage.from(springPage);

        assertTrue(page.items().isEmpty());
        assertEquals(0, page.page());
        assertEquals(0, page.total());
        assertEquals(0, page.totalPages());
    }

    // ========================================================================
    // hasNext / hasPrevious
    // ========================================================================

    @Test
    void hasNext_middlePage_returnsTrue() {
        OffsetPage<String> page = new OffsetPage<>(List.of("a"), 2, 10, 100, 10);
        assertTrue(page.hasNext());
    }

    @Test
    void hasNext_lastPage_returnsFalse() {
        OffsetPage<String> page = new OffsetPage<>(List.of("a"), 9, 10, 100, 10);
        assertFalse(page.hasNext());
    }

    @Test
    void hasNext_singlePage_returnsFalse() {
        OffsetPage<String> page = new OffsetPage<>(List.of("a"), 0, 10, 5, 1);
        assertFalse(page.hasNext());
    }

    @Test
    void hasPrevious_firstPage_returnsFalse() {
        OffsetPage<String> page = new OffsetPage<>(List.of("a"), 0, 10, 100, 10);
        assertFalse(page.hasPrevious());
    }

    @Test
    void hasPrevious_middlePage_returnsTrue() {
        OffsetPage<String> page = new OffsetPage<>(List.of("a"), 2, 10, 100, 10);
        assertTrue(page.hasPrevious());
    }

    @Test
    void hasPrevious_lastPage_returnsTrue() {
        OffsetPage<String> page = new OffsetPage<>(List.of("a"), 9, 10, 100, 10);
        assertTrue(page.hasPrevious());
    }

    // ========================================================================
    // Validation
    // ========================================================================

    @Test
    void constructor_negativePage_throwsIllegalArgument() {
        assertThrows(IllegalArgumentException.class,
            () -> new OffsetPage<>(List.of(), -1, 10, 0, 0));
    }

    @Test
    void constructor_zeroPageSize_throwsIllegalArgument() {
        assertThrows(IllegalArgumentException.class,
            () -> new OffsetPage<>(List.of(), 0, 0, 0, 0));
    }

    @Test
    void constructor_negativeTotal_throwsIllegalArgument() {
        assertThrows(IllegalArgumentException.class,
            () -> new OffsetPage<>(List.of(), 0, 10, -1, 0));
    }

    @Test
    void constructor_nullItems_throwsNullPointer() {
        assertThrows(NullPointerException.class,
            () -> new OffsetPage<>(null, 0, 10, 0, 0));
    }

    @Test
    void from_nullSpringPage_throwsNullPointer() {
        assertThrows(NullPointerException.class,
            () -> OffsetPage.from(null));
    }
}
