package com.factoryos.production.repository.pagination;

import static org.junit.jupiter.api.Assertions.*;

import java.time.Instant;
import java.util.List;
import java.util.UUID;
import java.util.function.Function;

import org.junit.jupiter.api.Test;

import com.factoryos.production.entity.WorkOrder;
import com.github.f4b6a3.uuid.UuidCreator;

class CursorPageTest {

    private static final List<SortCriteria> SORT_BY_ID_ASC = List.of(
            SortCriteria.asc(SortKey.id()));
    private static final List<SortCriteria> SORT_BY_CREATED_THEN_ID_ASC = List.of(
            SortCriteria.asc(SortKey.ofTimestamp("createdAt")),
            SortCriteria.asc(SortKey.id()));

    private WorkOrder buildWorkOrder(Instant createdAt) {
        WorkOrder wo = new WorkOrder();
        wo.setId(UuidCreator.getTimeOrderedEpoch());
        wo.setMaterialDefinitionId(UUID.randomUUID());
        wo.setRoutingSpecId(UUID.randomUUID());
        wo.setWorkCenterId(UUID.randomUUID());
        wo.setTargetQuantity("100");
        wo.setUnitOfMeasure("pcs");
        wo.setState("draft");
        wo.setPriority("medium");
        // createdAt is set by @PrePersist in production; set explicitly for unit tests
        try {
            var field = WorkOrder.class.getDeclaredField("createdAt");
            field.setAccessible(true);
            field.set(wo, createdAt);
        } catch (ReflectiveOperationException e) {
            throw new RuntimeException(e);
        }
        return wo;
    }

    private List<Object> extractId(WorkOrder wo) {
        return List.of(wo.getId());
    }

    // ========================================================================
    // Last Page (rawItems <= pageSize)
    // ========================================================================

    @Test
    void of_fewerItemsThanPageSize_returnsEmptyNextToken() {
        List<WorkOrder> items = List.of(
            buildWorkOrder(Instant.now()),
            buildWorkOrder(Instant.now())
        );

        CursorPage<WorkOrder> page = CursorPage.of(items, 10, SORT_BY_ID_ASC, this::extractId);

        assertEquals(2, page.items().size());
        assertEquals("", page.nextPageToken());
    }

    @Test
    void of_exactPageSize_returnsEmptyNextToken() {
        List<WorkOrder> items = List.of(
            buildWorkOrder(Instant.now()),
            buildWorkOrder(Instant.now()),
            buildWorkOrder(Instant.now())
        );

        CursorPage<WorkOrder> page = CursorPage.of(items, 3, SORT_BY_ID_ASC, this::extractId);

        assertEquals(3, page.items().size());
        assertEquals("", page.nextPageToken());
    }

    @Test
    void of_emptyItems_returnsEmptyNextToken() {
        CursorPage<WorkOrder> page = CursorPage.of(List.of(), 10, SORT_BY_ID_ASC, this::extractId);

        assertEquals(0, page.items().size());
        assertEquals("", page.nextPageToken());
    }

    // ========================================================================
    // Has Next Page (rawItems > pageSize)
    // ========================================================================

    @Test
    void of_moreItemsThanPageSize_trimsAndReturnsNextToken() {
        WorkOrder w1 = buildWorkOrder(Instant.parse("2026-09-01T00:00:00Z"));
        WorkOrder w2 = buildWorkOrder(Instant.parse("2026-09-02T00:00:00Z"));
        WorkOrder w3 = buildWorkOrder(Instant.parse("2026-09-03T00:00:00Z"));
        WorkOrder extra = buildWorkOrder(Instant.parse("2026-09-04T00:00:00Z"));

        CursorPage<WorkOrder> page = CursorPage.of(List.of(w1, w2, w3, extra), 3, SORT_BY_ID_ASC, this::extractId);

        assertEquals(3, page.items().size());
        assertFalse(page.nextPageToken().isEmpty());

        Cursor decoded = Cursor.decode(page.nextPageToken(), List.of(SortKey.id()));
        assertNotNull(decoded);
        assertEquals(w3.getId(), decoded.values().get(0));
    }

    // ========================================================================
    // Composite Key Extraction
    // ========================================================================

    @Test
    void of_compositeKey_nextTokenEncodesBothValues() {
        WorkOrder w1 = buildWorkOrder(Instant.parse("2026-09-01T00:00:00Z"));
        WorkOrder w2 = buildWorkOrder(Instant.parse("2026-09-02T00:00:00Z"));
        WorkOrder extra = buildWorkOrder(Instant.parse("2026-09-03T00:00:00Z"));

        Function<WorkOrder, List<Object>> compositeExtractor =
            wo -> List.of(wo.getCreatedAt(), wo.getId());

        CursorPage<WorkOrder> page = CursorPage.of(
            List.of(w1, w2, extra), 2, SORT_BY_CREATED_THEN_ID_ASC, compositeExtractor);

        assertEquals(2, page.items().size());
        assertFalse(page.nextPageToken().isEmpty());

        Cursor decoded = Cursor.decode(page.nextPageToken(),
            List.of(SortKey.ofTimestamp("createdAt"), SortKey.id()));
        assertNotNull(decoded);
        assertEquals(w2.getCreatedAt(), decoded.values().get(0));
        assertEquals(w2.getId(), decoded.values().get(1));
    }

    @Test
    void of_singleItemPage_trimmedCorrectly() {
        WorkOrder w1 = buildWorkOrder(Instant.now());
        WorkOrder extra = buildWorkOrder(Instant.now());

        CursorPage<WorkOrder> page = CursorPage.of(List.of(w1, extra), 1, SORT_BY_ID_ASC, this::extractId);

        assertEquals(1, page.items().size());
        assertFalse(page.nextPageToken().isEmpty());
        assertEquals(w1.getId(), page.items().get(0).getId());
    }

    @Test
    void of_cursorValuesComeFromLastItem_notExtra() {
        WorkOrder w1 = buildWorkOrder(Instant.parse("2026-09-01T00:00:00Z"));
        WorkOrder w2 = buildWorkOrder(Instant.parse("2026-09-02T00:00:00Z"));
        WorkOrder extra = buildWorkOrder(Instant.parse("2026-09-03T00:00:00Z"));

        CursorPage<WorkOrder> page = CursorPage.of(
            List.of(w1, w2, extra), 2, SORT_BY_ID_ASC, this::extractId);

        Cursor decoded = Cursor.decode(page.nextPageToken(), List.of(SortKey.id()));
        // Cursor should point to w2 (last item on page), not extra
        assertEquals(w2.getId(), decoded.values().get(0));
    }

    // ========================================================================
    // Validation
    // ========================================================================

    @Test
    void of_nullItems_throwsNullPointer() {
        assertThrows(NullPointerException.class,
            () -> CursorPage.of(null, 10, SORT_BY_ID_ASC, this::extractId));
    }

    @Test
    void of_nullSortCriteria_throwsNullPointer() {
        assertThrows(NullPointerException.class,
            () -> CursorPage.of(List.of(), 10, null, this::extractId));
    }

    @Test
    void of_nullExtractor_throwsNullPointer() {
        assertThrows(NullPointerException.class,
            () -> CursorPage.of(List.of(), 10, SORT_BY_ID_ASC, null));
    }
}
