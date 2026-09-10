package com.factoryos.production.repository.pagination;

import static org.junit.jupiter.api.Assertions.*;

import java.time.Instant;
import java.util.UUID;

import org.junit.jupiter.api.Test;
import org.springframework.data.domain.Sort;

/**
 * Tests for {@link SortCriteria} construction and factory methods.
 */
class SortCriteriaTest {

    private static final SortKey<UUID> ID_KEY = SortKey.id();
    private static final SortKey<Instant> CREATED_KEY = SortKey.ofTimestamp("createdAt");
    private static final SortKey<String> NAME_KEY = SortKey.ofString("name");

    // ========================================================================
    // Factory Methods
    // ========================================================================

    @Test
    void asc_pairsWithAscDirection() {
        SortCriteria criteria = SortCriteria.asc(ID_KEY);
        assertEquals(ID_KEY, criteria.key());
        assertEquals(Sort.Direction.ASC, criteria.direction());
    }

    @Test
    void desc_pairsWithDescDirection() {
        SortCriteria criteria = SortCriteria.desc(ID_KEY);
        assertEquals(ID_KEY, criteria.key());
        assertEquals(Sort.Direction.DESC, criteria.direction());
    }

    @Test
    void asc_differentKeyTypes() {
        SortCriteria idCriteria = SortCriteria.asc(ID_KEY);
        SortCriteria timeCriteria = SortCriteria.asc(CREATED_KEY);
        SortCriteria nameCriteria = SortCriteria.asc(NAME_KEY);

        assertEquals("id", idCriteria.key().fieldName());
        assertEquals("createdAt", timeCriteria.key().fieldName());
        assertEquals("name", nameCriteria.key().fieldName());
    }

    @Test
    void sameKeyDifferentDirections() {
        SortCriteria asc = SortCriteria.asc(CREATED_KEY);
        SortCriteria desc = SortCriteria.desc(CREATED_KEY);

        assertEquals(Sort.Direction.ASC, asc.direction());
        assertEquals(Sort.Direction.DESC, desc.direction());
        assertEquals(asc.key(), desc.key());
    }

    // ========================================================================
    // Validation
    // ========================================================================

    @Test
    void nullKey_throwsNullPointer() {
        assertThrows(NullPointerException.class,
            () -> new SortCriteria(null, Sort.Direction.ASC));
    }

    @Test
    void nullDirection_throwsNullPointer() {
        assertThrows(NullPointerException.class,
            () -> new SortCriteria(ID_KEY, null));
    }

    // ========================================================================
    // Equality
    // ========================================================================

    @Test
    void sameKeyAndDirection_areEqual() {
        SortCriteria a = SortCriteria.desc(CREATED_KEY);
        SortCriteria b = SortCriteria.desc(CREATED_KEY);
        assertEquals(a, b);
        assertEquals(a.hashCode(), b.hashCode());
    }

    @Test
    void differentDirection_notEqual() {
        SortCriteria a = SortCriteria.asc(ID_KEY);
        SortCriteria b = SortCriteria.desc(ID_KEY);
        assertNotEquals(a, b);
    }
}
