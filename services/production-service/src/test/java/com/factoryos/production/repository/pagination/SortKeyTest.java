package com.factoryos.production.repository.pagination;

import static org.junit.jupiter.api.Assertions.*;

import java.time.Instant;
import java.util.UUID;
import java.util.function.Function;

import org.junit.jupiter.api.Test;

/**
 * Tests for {@link SortKey} factories, value parsing, and construction validation.
 */
class SortKeyTest {

    // ========================================================================
    // SortKey.id()
    // ========================================================================

    @Test
    void id_fieldName() {
        assertEquals("id", SortKey.id().fieldName());
    }

    @Test
    void id_type() {
        assertEquals(UUID.class, SortKey.id().type());
    }

    @Test
    void id_parseValid() {
        UUID expected = UUID.fromString("550e8400-e29b-41d4-a716-446655440000");
        assertEquals(expected, SortKey.id().parseValue(expected.toString()));
    }

    @Test
    void id_parseInvalid_throws() {
        assertThrows(IllegalArgumentException.class,
            () -> SortKey.id().parseValue("not-a-uuid"));
    }

    @Test
    void id_parseNull_throws() {
        assertThrows(IllegalArgumentException.class,
            () -> SortKey.id().parseValue(null));
    }

    @Test
    void id_parseEmpty_throws() {
        assertThrows(IllegalArgumentException.class,
            () -> SortKey.id().parseValue(""));
    }

    // ========================================================================
    // SortKey.ofTimestamp()
    // ========================================================================

    @Test
    void ofTimestamp_fieldName() {
        assertEquals("createdAt", SortKey.ofTimestamp("createdAt").fieldName());
    }

    @Test
    void ofTimestamp_type() {
        assertEquals(Instant.class, SortKey.ofTimestamp("createdAt").type());
    }

    @Test
    void ofTimestamp_parseValid() {
        Instant expected = Instant.parse("2026-09-04T10:30:00Z");
        assertEquals(expected, SortKey.ofTimestamp("createdAt").parseValue("2026-09-04T10:30:00Z"));
    }

    @Test
    void ofTimestamp_parseInvalid_throws() {
        assertThrows(IllegalArgumentException.class,
            () -> SortKey.ofTimestamp("createdAt").parseValue("not-a-time"));
    }

    // ========================================================================
    // SortKey.ofString()
    // ========================================================================

    @Test
    void ofString_fieldName() {
        assertEquals("name", SortKey.ofString("name").fieldName());
    }

    @Test
    void ofString_type() {
        assertEquals(String.class, SortKey.ofString("name").type());
    }

    @Test
    void ofString_parseReturnsIdentity() {
        assertEquals("hello", SortKey.ofString("name").parseValue("hello"));
    }

    // ========================================================================
    // SortKey.of() — custom
    // ========================================================================

    @Test
    void custom_parseValid() {
        SortKey<Integer> key = SortKey.of("priority", Integer.class, Integer::parseInt);
        assertEquals(42, key.parseValue("42"));
    }

    @Test
    void custom_parseInvalid_throws() {
        SortKey<Integer> key = SortKey.of("priority", Integer.class, Integer::parseInt);
        assertThrows(IllegalArgumentException.class,
            () -> key.parseValue("not-a-number"));
    }

    // ========================================================================
    // Construction validation
    // ========================================================================

    @Test
    void nullFieldName_throwsNullPointer() {
        assertThrows(NullPointerException.class,
            () -> new SortKey<>(null, String.class, Function.identity()));
    }
}
