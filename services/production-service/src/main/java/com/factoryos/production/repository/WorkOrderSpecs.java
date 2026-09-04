package com.factoryos.production.repository;

import java.util.ArrayList;
import java.util.List;
import java.util.UUID;

import org.springframework.data.jpa.domain.Specification;

import com.factoryos.production.entity.WorkOrder;

import jakarta.persistence.criteria.Predicate;

public class WorkOrderSpecs {
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
}
