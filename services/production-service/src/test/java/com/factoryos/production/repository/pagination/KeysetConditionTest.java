package com.factoryos.production.repository.pagination;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

import java.time.Instant;
import java.util.List;
import java.util.UUID;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.mockito.junit.jupiter.MockitoSettings;
import org.mockito.quality.Strictness;
import org.springframework.data.domain.Sort;

import jakarta.persistence.criteria.CriteriaBuilder;
import jakarta.persistence.criteria.Expression;
import jakarta.persistence.criteria.Path;
import jakarta.persistence.criteria.Predicate;
import jakarta.persistence.criteria.Root;

/**
 * Tests for {@link KeysetCondition#build} — JPA Criteria WHERE clause generation
 * for keyset (cursor) pagination with mixed sort directions.
 *
 * <p>Uses Mockito with lenient stubs for the JPA Criteria API (which has many
 * overloaded methods that cause ambiguity). Tests focus on behavioral outcomes:
 * null handling, root field access, and predicate structure.
 */
@ExtendWith(MockitoExtension.class)
@MockitoSettings(strictness = Strictness.LENIENT)
class KeysetConditionTest {

    @Mock private Root<?> root;
    @Mock private CriteriaBuilder cb;

    @Mock private Path<UUID> idPath;
    @Mock private Path<Instant> createdAtPath;
    @Mock private Path<String> namePath;

    @Mock private Predicate predicate;
    @Mock private Predicate andPredicate;
    @Mock private Predicate orPredicate;

    private static final UUID TEST_ID = UUID.fromString("550e8400-e29b-41d4-a716-446655440000");
    private static final Instant TEST_TIME = Instant.parse("2026-09-04T10:30:00Z");

    @BeforeEach
    void setUp() {
        // Typed doReturn avoids JPA overload ambiguity on Root.get()
        doReturn(idPath).when(root).get("id");
        doReturn(createdAtPath).when(root).get("createdAt");
        doReturn(namePath).when(root).get("name");

        // Stubs for CriteriaBuilder comparison methods
        // doReturn avoids ambiguity on lessThan/greaterThan overloads
        doReturn(predicate).when(cb).equal(any(Expression.class), any());
        doReturn(predicate).when(cb).lessThan(any(Expression.class), any(Comparable.class));
        doReturn(predicate).when(cb).greaterThan(any(Expression.class), any(Comparable.class));
        doReturn(andPredicate).when(cb).and(any(Predicate[].class));
        doReturn(orPredicate).when(cb).or(any(Predicate[].class));
    }

    // ========================================================================
    // Null / First Page
    // ========================================================================

    @Test
    void nullCursor_returnsNull() {
        List<SortCriteria> sortCriteriaList = List.of(SortCriteria.desc(SortKey.id()));
        Predicate result = KeysetCondition.build(root, cb, null, sortCriteriaList);
        assertNull(result);
    }

    // ========================================================================
    // Validation
    // ========================================================================

    @Test
    void nullRoot_throwsNullPointer() {
        Cursor cursor = Cursor.of(List.of(SortKey.id()), List.of(TEST_ID));
        assertThrows(NullPointerException.class,
            () -> KeysetCondition.build(null, cb, cursor, List.of(SortCriteria.desc(SortKey.id()))));
    }

    @Test
    void nullCb_throwsNullPointer() {
        Cursor cursor = Cursor.of(List.of(SortKey.id()), List.of(TEST_ID));
        assertThrows(NullPointerException.class,
            () -> KeysetCondition.build(root, null, cursor, List.of(SortCriteria.desc(SortKey.id()))));
    }

    @Test
    void nullSortCriteriaList_throwsNullPointer() {
        Cursor cursor = Cursor.of(List.of(SortKey.id()), List.of(TEST_ID));
        assertThrows(NullPointerException.class,
            () -> KeysetCondition.build(root, cb, cursor, null));
    }

    // ========================================================================
    // Single Key — DESC
    // ========================================================================

    @Test
    void singleKey_desc_returnsOrPredicate() {
        List<SortCriteria> sort = List.of(SortCriteria.desc(SortKey.id()));
        Cursor cursor = Cursor.of(List.of(SortKey.id()), List.of(TEST_ID));

        Predicate result = KeysetCondition.build(root, cb, cursor, sort);

        assertSame(orPredicate, result);
    }

    @Test
    void singleKey_desc_accessesIdField() {
        List<SortCriteria> sort = List.of(SortCriteria.desc(SortKey.id()));
        Cursor cursor = Cursor.of(List.of(SortKey.id()), List.of(TEST_ID));

        KeysetCondition.build(root, cb, cursor, sort);

        verify(root, atLeastOnce()).get("id");
    }

    // ========================================================================
    // Single Key — ASC
    // ========================================================================

    @Test
    void singleKey_asc_returnsOrPredicate() {
        List<SortCriteria> sort = List.of(SortCriteria.asc(SortKey.id()));
        Cursor cursor = Cursor.of(List.of(SortKey.id()), List.of(TEST_ID));

        Predicate result = KeysetCondition.build(root, cb, cursor, sort);

        assertSame(orPredicate, result);
    }

    @Test
    void singleKey_asc_accessesIdField() {
        List<SortCriteria> sort = List.of(SortCriteria.asc(SortKey.id()));
        Cursor cursor = Cursor.of(List.of(SortKey.id()), List.of(TEST_ID));

        KeysetCondition.build(root, cb, cursor, sort);

        verify(root, atLeastOnce()).get("id");
    }

    // ========================================================================
    // Composite Keys — Two keys
    // ========================================================================

    @Test
    void compositeKey_twoKeys_returnsOrPredicate() {
        List<SortCriteria> sort = List.of(
            SortCriteria.desc(SortKey.ofTimestamp("createdAt")),
            SortCriteria.desc(SortKey.id()));
        Cursor cursor = Cursor.of(
            List.of(SortKey.ofTimestamp("createdAt"), SortKey.id()),
            List.of(TEST_TIME, TEST_ID));

        Predicate result = KeysetCondition.build(root, cb, cursor, sort);

        assertSame(orPredicate, result);
    }

    @Test
    void compositeKey_twoKeys_accessesBothFields() {
        List<SortCriteria> sort = List.of(
            SortCriteria.desc(SortKey.ofTimestamp("createdAt")),
            SortCriteria.desc(SortKey.id()));
        Cursor cursor = Cursor.of(
            List.of(SortKey.ofTimestamp("createdAt"), SortKey.id()),
            List.of(TEST_TIME, TEST_ID));

        KeysetCondition.build(root, cb, cursor, sort);

        verify(root, atLeastOnce()).get("createdAt");
        verify(root, atLeastOnce()).get("id");
    }

    // ========================================================================
    // Composite Keys — Three keys
    // ========================================================================

    @Test
    void compositeKey_threeKeys_returnsOrPredicate() {
        List<SortCriteria> sort = List.of(
            SortCriteria.asc(SortKey.ofString("name")),
            SortCriteria.desc(SortKey.ofTimestamp("createdAt")),
            SortCriteria.desc(SortKey.id()));
        Cursor cursor = Cursor.of(
            List.of(SortKey.ofString("name"), SortKey.ofTimestamp("createdAt"), SortKey.id()),
            List.of("Alice", TEST_TIME, TEST_ID));

        Predicate result = KeysetCondition.build(root, cb, cursor, sort);

        assertSame(orPredicate, result);
    }

    @Test
    void compositeKey_threeKeys_accessesAllFields() {
        List<SortCriteria> sort = List.of(
            SortCriteria.asc(SortKey.ofString("name")),
            SortCriteria.desc(SortKey.ofTimestamp("createdAt")),
            SortCriteria.desc(SortKey.id()));
        Cursor cursor = Cursor.of(
            List.of(SortKey.ofString("name"), SortKey.ofTimestamp("createdAt"), SortKey.id()),
            List.of("Alice", TEST_TIME, TEST_ID));

        KeysetCondition.build(root, cb, cursor, sort);

        verify(root, atLeastOnce()).get("name");
        verify(root, atLeastOnce()).get("createdAt");
        verify(root, atLeastOnce()).get("id");
    }

    // ========================================================================
    // Mixed Directions
    // ========================================================================

    @Test
    void mixedDirections_returnsOrPredicate() {
        List<SortCriteria> sort = List.of(
            SortCriteria.asc(SortKey.ofString("name")),
            SortCriteria.desc(SortKey.id()));
        Cursor cursor = Cursor.of(
            List.of(SortKey.ofString("name"), SortKey.id()),
            List.of("Alice", TEST_ID));

        Predicate result = KeysetCondition.build(root, cb, cursor, sort);

        assertSame(orPredicate, result);
    }

    @Test
    void mixedDirections_accessesCorrectFields() {
        List<SortCriteria> sort = List.of(
            SortCriteria.asc(SortKey.ofString("name")),
            SortCriteria.desc(SortKey.id()));
        Cursor cursor = Cursor.of(
            List.of(SortKey.ofString("name"), SortKey.id()),
            List.of("Alice", TEST_ID));

        KeysetCondition.build(root, cb, cursor, sort);

        verify(root, atLeastOnce()).get("name");
        verify(root, atLeastOnce()).get("id");
    }

    // ========================================================================
    // Root field access uses cursor keyNames
    // ========================================================================

    @Test
    void build_singleKey_accessesOneField() {
        List<SortCriteria> sort = List.of(SortCriteria.asc(SortKey.id()));
        Cursor cursor = Cursor.of(List.of(SortKey.id()), List.of(TEST_ID));

        KeysetCondition.build(root, cb, cursor, sort);

        verify(root, atLeastOnce()).get("id");
        verify(root, times(1)).get(anyString());
    }

    // ========================================================================
    // DESC-only composite — returns predicate
    // ========================================================================

    @Test
    void descComposite_returnsOrPredicate() {
        List<SortCriteria> sort = List.of(
            SortCriteria.desc(SortKey.ofTimestamp("createdAt")),
            SortCriteria.desc(SortKey.id()));
        Cursor cursor = Cursor.of(
            List.of(SortKey.ofTimestamp("createdAt"), SortKey.id()),
            List.of(TEST_TIME, TEST_ID));

        Predicate result = KeysetCondition.build(root, cb, cursor, sort);

        assertSame(orPredicate, result);
    }

    // ========================================================================
    // ASC-only composite — returns predicate
    // ========================================================================

    @Test
    void ascComposite_returnsOrPredicate() {
        List<SortCriteria> sort = List.of(
            SortCriteria.asc(SortKey.ofTimestamp("createdAt")),
            SortCriteria.asc(SortKey.id()));
        Cursor cursor = Cursor.of(
            List.of(SortKey.ofTimestamp("createdAt"), SortKey.id()),
            List.of(TEST_TIME, TEST_ID));

        Predicate result = KeysetCondition.build(root, cb, cursor, sort);

        assertSame(orPredicate, result);
    }

    // ========================================================================
    // Verify cb.or is invoked (produces an OR-chain)
    // ========================================================================

    @Test
    void build_callsCbOr() {
        List<SortCriteria> sort = List.of(
            SortCriteria.desc(SortKey.ofTimestamp("createdAt")),
            SortCriteria.desc(SortKey.id()));
        Cursor cursor = Cursor.of(
            List.of(SortKey.ofTimestamp("createdAt"), SortKey.id()),
            List.of(TEST_TIME, TEST_ID));

        KeysetCondition.build(root, cb, cursor, sort);

        verify(cb).or(any(Predicate[].class));
    }

    @Test
    void build_callsCbAndForEachPosition() {
        List<SortCriteria> sort = List.of(
            SortCriteria.desc(SortKey.ofTimestamp("createdAt")),
            SortCriteria.desc(SortKey.id()));
        Cursor cursor = Cursor.of(
            List.of(SortKey.ofTimestamp("createdAt"), SortKey.id()),
            List.of(TEST_TIME, TEST_ID));

        KeysetCondition.build(root, cb, cursor, sort);

        // 2 sort keys → 2 AND groups
        verify(cb, times(2)).and(any(Predicate[].class));
    }

    // ========================================================================
    // Three-key composite — 3 AND groups
    // ========================================================================

    @Test
    void compositeKey_threeKeys_callsCbAndThreeTimes() {
        List<SortCriteria> sort = List.of(
            SortCriteria.asc(SortKey.ofString("name")),
            SortCriteria.desc(SortKey.ofTimestamp("createdAt")),
            SortCriteria.desc(SortKey.id()));
        Cursor cursor = Cursor.of(
            List.of(SortKey.ofString("name"), SortKey.ofTimestamp("createdAt"), SortKey.id()),
            List.of("Alice", TEST_TIME, TEST_ID));

        KeysetCondition.build(root, cb, cursor, sort);

        verify(cb, times(3)).and(any(Predicate[].class));
    }
}
