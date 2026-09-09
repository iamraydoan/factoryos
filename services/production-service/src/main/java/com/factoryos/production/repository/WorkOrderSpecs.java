package com.factoryos.production.repository;

import java.util.ArrayList;
import java.util.List;
import java.util.Objects;
import java.util.UUID;

import org.springframework.data.jpa.domain.Specification;

import com.factoryos.production.entity.WorkOrder;
import com.factoryos.production.repository.pagination.Cursor;
import com.factoryos.production.repository.pagination.KeysetCondition;
import com.factoryos.production.repository.pagination.SortCriteria;
import com.factoryos.production.repository.pagination.SortKey;

import jakarta.persistence.criteria.Predicate;

/**
 * JPA {@link Specification} builders for {@link WorkOrder} queries.
 *
 * <p>Provides two patterns:
 * <ul>
 *   <li><b>Offset pagination:</b> {@link #withFilters} — filter-only, use with
 *       {@code Pageable}</li>
 *   <li><b>Cursor pagination:</b> {@link #withFiltersAndCursor} — filter + keyset
 *       condition, use with {@link CursorPageRequest}</li>
 * </ul>
 */
public class WorkOrderSpecs {

    /** Newest first (DESC by createdAt), tiebreaker by ID (DESC). */
    public static final List<SortCriteria> SORT_NEWEST_FIRST = List.of(
            SortCriteria.desc(SortKey.ofTimestamp("createdAt")),
            SortCriteria.desc(SortKey.id()));

    /** Oldest first (ASC by createdAt), tiebreaker by ID (ASC). */
    public static final List<SortCriteria> SORT_OLDEST_FIRST = List.of(
            SortCriteria.asc(SortKey.ofTimestamp("createdAt")),
            SortCriteria.asc(SortKey.id()));

    /** Alphabetical by state (ASC), then newest first (DESC), tiebreaker by ID (DESC). */
    public static final List<SortCriteria> SORT_BY_STATE_THEN_NEWEST = List.of(
            SortCriteria.asc(SortKey.ofString("state")),
            SortCriteria.desc(SortKey.ofTimestamp("createdAt")),
            SortCriteria.desc(SortKey.id()));

    /**
     * Builds a filter-only specification for offset-based pagination.
     *
     * @param workCenterId optional work center filter (null = no filter)
     * @param state        optional state filter (null = no filter)
     * @return a specification combining the non-null filters
     */
    public static Specification<WorkOrder> withFilters(UUID workCenterId, String state) {
        return (root, query, cb) -> {
            List<Predicate> predicates = new ArrayList<>();
            if (workCenterId != null) {
                predicates.add(cb.equal(root.get("workCenterId"), workCenterId));
            }
            if (state != null) {
                predicates.add(cb.equal(root.get("state"), state));
            }
            return cb.and(predicates.toArray(new Predicate[0]));
        };
    }

    /**
     * Builds a filter + keyset condition specification for cursor-based pagination.
     *
     * @param workCenterId    optional work center filter (null = no filter)
     * @param state           optional state filter (null = no filter)
     * @param cursor          the decoded cursor (null for first page)
     * @param sortCriteriaList the sort criteria defining keyset fields and directions
     * @return a specification combining filters and the keyset condition
     * @throws NullPointerException if sortCriteriaList is null
     */
    public static Specification<WorkOrder> withFiltersAndCursor(
            UUID workCenterId, String state, Cursor cursor, List<SortCriteria> sortCriteriaList) {
        Objects.requireNonNull(sortCriteriaList, "sortCriteriaList must not be null");
        Specification<WorkOrder> filters = withFilters(workCenterId, state);
        if (cursor == null) {
            return filters;
        }
        return filters.and((root, query, cb) -> KeysetCondition.build(root, cb, cursor, sortCriteriaList));
    }
}
