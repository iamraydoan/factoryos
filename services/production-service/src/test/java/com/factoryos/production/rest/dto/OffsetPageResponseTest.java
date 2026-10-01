package com.factoryos.production.rest.dto;

import static org.junit.jupiter.api.Assertions.*;

import java.util.List;

import org.junit.jupiter.api.Test;

import com.factoryos.production.repository.pagination.OffsetPage;

class OffsetPageResponseTest {

    @Test
    void from_firstInternalPage_mapsToWirePageOne() {
        OffsetPage<String> page = new OffsetPage<>(List.of("a", "b"), 0, 20, 100, 5);

        OffsetPageResponse<String> response = OffsetPageResponse.from(page);

        assertEquals(List.of("a", "b"), response.data());
        assertEquals(1, response.meta().page());
        assertEquals(20, response.meta().limit());
        assertEquals(100, response.meta().total());
        assertEquals(5, response.meta().totalPages());
    }

    @Test
    void from_thirdInternalPage_mapsToWirePageThree() {
        OffsetPage<String> page = new OffsetPage<>(List.of("c"), 2, 20, 100, 5);

        OffsetPageResponse<String> response = OffsetPageResponse.from(page);

        assertEquals(3, response.meta().page());
    }

    @Test
    void from_emptyPage_mapsToWirePageOneWithZeroTotals() {
        OffsetPage<String> page = new OffsetPage<>(List.of(), 0, 20, 0, 0);

        OffsetPageResponse<String> response = OffsetPageResponse.from(page);

        assertTrue(response.data().isEmpty());
        assertEquals(1, response.meta().page());
        assertEquals(0, response.meta().total());
        assertEquals(0, response.meta().totalPages());
    }
}
