package com.factoryos.production.repository.pagination;

import java.util.List;
import java.util.Objects;

import org.springframework.data.domain.Sort;

import jakarta.persistence.criteria.CriteriaBuilder;
import jakarta.persistence.criteria.Path;
import jakarta.persistence.criteria.Predicate;
import jakarta.persistence.criteria.Root;

/**
 * Builds JPA Criteria WHERE clauses for keyset (cursor) pagination with mixed sort directions.
 *
 * <p>Converts a cursor's key values into an OR-chain of AND-conditions that implements
 * tuple comparison without requiring SQL row-value constructors.
 *
 * <p>Example for SortCriteria=[createdAt DESC, name ASC, id DESC] and values=[t1, "Alice", id1]:
 * <pre>
 *   (createdAt &lt; t1)
 *   OR (createdAt = t1 AND name &gt; 'Alice')
 *   OR (createdAt = t1 AND name = 'Alice' AND id &lt; id1)
 * </pre>
 */
public final class KeysetCondition {

    private KeysetCondition() {
        // Utility class
    }

    /**
     * Builds a keyset predicate from a cursor with per-key sort criteria.
     *
     * @param root           the JPA root (FROM clause)
     * @param cb             the criteria builder
     * @param cursor         the decoded cursor (null for first page)
     * @param sortCriteriaList the sort criteria defining fields and directions
     * @return a predicate for the WHERE clause, or null if cursor is null (first page)
     */
    @SuppressWarnings("unchecked")
    public static Predicate build(Root<?> root, CriteriaBuilder cb,
                                  Cursor cursor,
                                  List<SortCriteria> sortCriteriaList) {
        Objects.requireNonNull(root, "root must not be null");
        Objects.requireNonNull(cb, "cb must not be null");
        Objects.requireNonNull(sortCriteriaList, "sortCriteriaList must not be null");

        if (cursor == null) {
            return null; // First page — no WHERE clause
        }

        List<String> keyNames = cursor.keyNames();
        List<Object> values = cursor.values();
        int size = keyNames.size();

        Predicate[] orParts = new Predicate[size];

        for (int i = 0; i < size; i++) {
            Predicate[] andParts = new Predicate[i + 1];

            // Preceding keys: key[j] = val[j]
            for (int j = 0; j < i; j++) {
                Path<Comparable<Object>> path = (Path<Comparable<Object>>) (Path<?>) root.get(keyNames.get(j));
                andParts[j] = cb.equal(path, values.get(j));
            }

            // Current key: compare based on its direction
            Path<Comparable<Object>> currentPath = (Path<Comparable<Object>>) (Path<?>) root.get(keyNames.get(i));
            Sort.Direction dir = sortCriteriaList.get(i).direction();
            Comparable<Object> val = (Comparable<Object>) values.get(i);

            if (dir == Sort.Direction.DESC) {
                andParts[i] = cb.lessThan(currentPath, val);
            } else {
                andParts[i] = cb.greaterThan(currentPath, val);
            }

            orParts[i] = cb.and(andParts);
        }

        return cb.or(orParts);
    }
}
