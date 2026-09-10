package com.factoryos.production.rest;

import java.util.UUID;

import org.springframework.data.jpa.domain.Specification;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import com.factoryos.production.entity.WorkOrder;
import com.factoryos.production.repository.WorkOrderRepository;
import com.factoryos.production.repository.WorkOrderSpecs;
import com.factoryos.production.repository.pagination.OffsetPage;
import com.factoryos.production.repository.pagination.OffsetPageRequest;

@RestController
@RequestMapping("/api/v1")
public class WorkOrderController {
    private final WorkOrderRepository workOrderRepository;

    public WorkOrderController(WorkOrderRepository workOrderRepository) {
        this.workOrderRepository = workOrderRepository;
    }

    @GetMapping("/work-orders")
    public OffsetPage<WorkOrder> listWorkOrders(@RequestParam(defaultValue = "0") int page,
            @RequestParam(defaultValue = "20") int limit,
            @RequestParam(required = false) UUID workCenterId,
            @RequestParam(required = false) String state) {

        OffsetPageRequest pageReq = OffsetPageRequest.of(page, limit, WorkOrderSpecs.SORT_NEWEST_FIRST);
        Specification<WorkOrder> spec = WorkOrderSpecs.withFilters(workCenterId, state);
        return OffsetPage.from(workOrderRepository.findAll(spec, pageReq.toPageable()));
    }

}
