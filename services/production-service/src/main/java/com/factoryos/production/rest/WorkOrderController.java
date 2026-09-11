package com.factoryos.production.rest;

import java.util.List;
import java.util.UUID;

import org.springframework.data.domain.PageRequest;
import org.springframework.data.jpa.domain.Specification;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import com.factoryos.production.entity.WorkOrder;
import com.factoryos.production.repository.WorkOrderRepository;
import com.factoryos.production.repository.WorkOrderSpecs;
import com.factoryos.production.repository.pagination.CursorPageRequest;
import com.factoryos.production.repository.pagination.OffsetPageRequest;
import com.factoryos.production.rest.dto.CursorPageResponse;
import com.factoryos.production.rest.dto.OffsetPageResponse;

@RestController
@RequestMapping("/api/v1/work-orders")
public class WorkOrderController {
    private final WorkOrderRepository workOrderRepository;

    public WorkOrderController(WorkOrderRepository workOrderRepository) {
        this.workOrderRepository = workOrderRepository;
    }

    @GetMapping("")
    public OffsetPageResponse<WorkOrder> listWorkOrders(@RequestParam(defaultValue = "0") int page,
            @RequestParam(defaultValue = "20") int limit,
            @RequestParam(required = false) UUID workCenterId,
            @RequestParam(required = false) String state) {
        OffsetPageRequest pageReq = OffsetPageRequest.of(page, limit, WorkOrderSpecs.SORT_NEWEST_FIRST);
        Specification<WorkOrder> spec = WorkOrderSpecs.withFilters(workCenterId, state);

        return OffsetPageResponse.of(workOrderRepository.findAll(spec, pageReq.toPageable()));
    }

    @GetMapping(value = "", params = "cursor")
    public CursorPageResponse<WorkOrder> listWorkOrdersCursor(
            @RequestParam(defaultValue = "20") int limit,
            @RequestParam(required = false) String cursor,
            @RequestParam(required = false) UUID workCenterId,
            @RequestParam(required = false) String state) {
        CursorPageRequest pageReq = CursorPageRequest.of(limit, cursor, WorkOrderSpecs.SORT_NEWEST_FIRST);
        Specification<WorkOrder> spec = WorkOrderSpecs.withFiltersAndCursor(workCenterId, state, pageReq.cursor(),
                pageReq.sortCriteriaList());

        List<WorkOrder> raw = workOrderRepository
                .findAll(spec, PageRequest.of(0, pageReq.queryLimit(), pageReq.toSort()))
                .getContent();
        return CursorPageResponse.of(raw, pageReq.pageSize(), pageReq.sortCriteriaList(),
                wo -> List.of(wo.getCreatedAt(), wo.getId()));
    }
}
