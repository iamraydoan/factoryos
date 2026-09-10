package com.factoryos.production.repository.pagination;

import static org.junit.jupiter.api.Assertions.*;

import java.time.Instant;
import java.util.List;
import java.util.UUID;

import org.junit.jupiter.api.Test;

class CursorTest {

    // ========================================================================
    // Encode / Decode Round-Trip: Single Key
    // ========================================================================

    @Test
    void encodeDecode_singleKey_id() {
        UUID id = UUID.fromString("550e8400-e29b-41d4-a716-446655440000");
        Cursor original = Cursor.of(List.of(SortKey.id()), List.of(id));

        String token = original.encode();
        Cursor decoded = Cursor.decode(token, List.of(SortKey.id()));

        assertNotNull(decoded);
        assertEquals(List.of("id"), decoded.keyNames());
        assertEquals(id, decoded.values().get(0));
    }

    @Test
    void encodeDecode_singleKey_timestamp() {
        SortKey<Instant> createdAtKey = SortKey.ofTimestamp("createdAt");
        Instant now = Instant.parse("2026-09-04T10:30:00Z");
        Cursor original = Cursor.of(List.of(createdAtKey), List.of(now));

        String token = original.encode();
        Cursor decoded = Cursor.decode(token, List.of(createdAtKey));

        assertNotNull(decoded);
        assertEquals(List.of("createdAt"), decoded.keyNames());
        assertEquals(now, decoded.values().get(0));
    }

    @Test
    void encodeDecode_singleKey_string() {
        SortKey<String> nameKey = SortKey.ofString("name");
        Cursor original = Cursor.of(List.of(nameKey), List.of("Widget-A"));

        String token = original.encode();
        Cursor decoded = Cursor.decode(token, List.of(nameKey));

        assertNotNull(decoded);
        assertEquals("Widget-A", decoded.values().get(0));
    }

    // ========================================================================
    // Encode / Decode Round-Trip: Composite Key
    // ========================================================================

    @Test
    void encodeDecode_compositeKey_timestampAndId() {
        SortKey<Instant> createdAtKey = SortKey.ofTimestamp("createdAt");
        Instant createdAt = Instant.parse("2026-09-04T10:30:00Z");
        UUID id = UUID.fromString("550e8400-e29b-41d4-a716-446655440000");

        Cursor original = Cursor.of(List.of(createdAtKey, SortKey.id()), List.of(createdAt, id));

        String token = original.encode();
        Cursor decoded = Cursor.decode(token, List.of(createdAtKey, SortKey.id()));

        assertNotNull(decoded);
        assertEquals(List.of("createdAt", "id"), decoded.keyNames());
        assertEquals(createdAt, decoded.values().get(0));
        assertEquals(id, decoded.values().get(1));
    }

    @Test
    void encodeDecode_compositeKey_nameAndId() {
        SortKey<String> nameKey = SortKey.ofString("name");
        UUID id = UUID.fromString("550e8400-e29b-41d4-a716-446655440000");

        Cursor original = Cursor.of(List.of(nameKey, SortKey.id()), List.of("Widget-A", id));

        String token = original.encode();
        Cursor decoded = Cursor.decode(token, List.of(nameKey, SortKey.id()));

        assertNotNull(decoded);
        assertEquals("Widget-A", decoded.values().get(0));
        assertEquals(id, decoded.values().get(1));
    }

    // ========================================================================
    // Token Format
    // ========================================================================

    @Test
    void tokenStartsWithVersionPrefix() {
        Cursor cursor = Cursor.of(List.of(SortKey.id()), List.of(UUID.randomUUID()));
        assertTrue(cursor.encode().startsWith("v1."));
    }

    @Test
    void tokenIsUrlSafe() {
        Cursor cursor = Cursor.of(List.of(SortKey.id()), List.of(UUID.randomUUID()));
        String base64Part = cursor.encode().substring(3);
        assertTrue(base64Part.matches("[A-Za-z0-9_-]+"));
    }

    // ========================================================================
    // Decode: Null / Empty = First Page
    // ========================================================================

    @Test
    void decode_nullToken_returnsNull() {
        assertNull(Cursor.decode(null, List.of(SortKey.id())));
    }

    @Test
    void decode_emptyToken_returnsNull() {
        assertNull(Cursor.decode("", List.of(SortKey.id())));
    }

    // ========================================================================
    // Decode: Validation Errors
    // ========================================================================

    @Test
    void decode_missingVersionPrefix_throwsInvalidCursor() {
        InvalidCursorException ex = assertThrows(InvalidCursorException.class,
            () -> Cursor.decode("not-a-cursor", List.of(SortKey.id())));
        assertTrue(ex.getMessage().contains("version prefix"));
    }

    @Test
    void decode_wrongVersionPrefix_throwsInvalidCursor() {
        // Valid base64 JSON, but prefix is v2 instead of v1
        String base64 = java.util.Base64.getUrlEncoder().withoutPadding()
            .encodeToString("{\"keys\":[\"id\"],\"vals\":[\"550e8400-e29b-41d4-a716-446655440000\"]}".getBytes());
        InvalidCursorException ex = assertThrows(InvalidCursorException.class,
            () -> Cursor.decode("v2." + base64, List.of(SortKey.id())));
        assertTrue(ex.getMessage().contains("version prefix"));
    }

    @Test
    void decode_malformedBase64_throwsInvalidCursor() {
        assertThrows(InvalidCursorException.class,
            () -> Cursor.decode("v1.!!!invalid!!!", List.of(SortKey.id())));
    }

    @Test
    void decode_malformedJson_throwsInvalidCursor() {
        String base64 = java.util.Base64.getUrlEncoder().withoutPadding()
            .encodeToString("not json".getBytes());
        assertThrows(InvalidCursorException.class,
            () -> Cursor.decode("v1." + base64, List.of(SortKey.id())));
    }

    @Test
    void decode_keysMismatch_throwsInvalidCursor() {
        String token = Cursor.of(List.of(SortKey.id()), List.of(UUID.randomUUID())).encode();
        InvalidCursorException ex = assertThrows(InvalidCursorException.class,
            () -> Cursor.decode(token, List.of(SortKey.ofTimestamp("createdAt"), SortKey.id())));
        assertTrue(ex.getMessage().contains("mismatch"));
    }

    @Test
    void decode_invalidUuidValue_throwsInvalidCursor() {
        String base64 = java.util.Base64.getUrlEncoder().withoutPadding()
            .encodeToString("{\"keys\":[\"id\"],\"vals\":[\"not-a-uuid\"]}".getBytes());
        assertThrows(InvalidCursorException.class,
            () -> Cursor.decode("v1." + base64, List.of(SortKey.id())));
    }

    @Test
    void decode_invalidInstantValue_throwsInvalidCursor() {
        SortKey<Instant> key = SortKey.ofTimestamp("createdAt");
        String base64 = java.util.Base64.getUrlEncoder().withoutPadding()
            .encodeToString("{\"keys\":[\"createdAt\"],\"vals\":[\"not-a-time\"]}".getBytes());
        assertThrows(InvalidCursorException.class,
            () -> Cursor.decode("v1." + base64, List.of(key)));
    }

    // ========================================================================
    // of(): Validation
    // ========================================================================

    @Test
    void of_emptyKeys_throwsIllegalArgument() {
        assertThrows(IllegalArgumentException.class,
            () -> Cursor.of(List.of(), List.of()));
    }

    @Test
    void of_sizeMismatch_throwsIllegalArgument() {
        assertThrows(IllegalArgumentException.class,
            () -> Cursor.of(List.of(SortKey.id()), List.of()));
    }

    @Test
    void of_nullKeys_throwsNullPointer() {
        assertThrows(NullPointerException.class,
            () -> Cursor.of(null, List.of(UUID.randomUUID())));
    }

    @Test
    void of_nullValues_throwsNullPointer() {
        assertThrows(NullPointerException.class,
            () -> Cursor.of(List.of(SortKey.id()), null));
    }

    // ========================================================================
    // Equality
    // ========================================================================

    @Test
    void equalCursors_areEqual() {
        UUID id = UUID.fromString("550e8400-e29b-41d4-a716-446655440000");
        Cursor c1 = Cursor.of(List.of(SortKey.id()), List.of(id));
        Cursor c2 = Cursor.of(List.of(SortKey.id()), List.of(id));
        assertEquals(c1, c2);
        assertEquals(c1.hashCode(), c2.hashCode());
    }

    @Test
    void differentCursors_areNotEqual() {
        Cursor c1 = Cursor.of(List.of(SortKey.id()), List.of(UUID.randomUUID()));
        Cursor c2 = Cursor.of(List.of(SortKey.id()), List.of(UUID.randomUUID()));
        assertNotEquals(c1, c2);
    }
}
