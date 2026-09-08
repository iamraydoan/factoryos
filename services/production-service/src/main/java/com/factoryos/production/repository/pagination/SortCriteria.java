package com.factoryos.production.repository.pagination;

import java.util.Objects;

import org.springframework.data.domain.Sort;

/**
 * Pairs a {@link SortKey} with a {@link Sort.Direction} for mixed-direction
 * sorting.
 *
 * <p>
 * Enables different sort keys to have different directions in the same query.
 * For example: sort by {@code createdAt DESC, name ASC, id DESC}.
 *
 * <p>
 * This class is immutable and thread-safe.
 *
 * @param key       the sort key (field definition)
 * @param direction the sort direction (ASC or DESC)
 */
public record SortCriteria(SortKey<?> key, Sort.Direction direction) {
    public SortCriteria {
        Objects.requireNonNull(key, "key must not be null");
        Objects.requireNonNull(direction, "direction must not be null");
    }

    public static SortCriteria asc(SortKey<?> key) {
        return new SortCriteria(key, Sort.Direction.ASC);
    }

    public static SortCriteria desc(SortKey<?> key) {
        return new SortCriteria(key, Sort.Direction.DESC);
    }
}
