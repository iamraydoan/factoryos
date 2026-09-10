package com.factoryos.production.grpc;

import java.util.List;
import java.util.UUID;

import org.springframework.data.domain.PageRequest;
import org.springframework.data.jpa.domain.Specification;
import org.springframework.transaction.annotation.Transactional;

import com.factoryos.production.entity.WorkOrder;
import com.factoryos.production.proto.ListWorkOrdersRequest;
import com.factoryos.production.proto.ListWorkOrdersResponse;
import com.factoryos.production.proto.ProductionServiceGrpc;
import com.factoryos.production.repository.WorkOrderRepository;
import com.factoryos.production.repository.WorkOrderSpecs;
import com.factoryos.production.repository.pagination.CursorPage;
import com.factoryos.production.repository.pagination.CursorPageRequest;
import com.google.protobuf.Timestamp;

import io.grpc.stub.StreamObserver;
import net.devh.boot.grpc.server.service.GrpcService;

@GrpcService
public class WorkOrderGrpcService extends ProductionServiceGrpc.ProductionServiceImplBase {
    private final WorkOrderRepository workOrderRepository;

    public WorkOrderGrpcService(WorkOrderRepository workOrderRepository) {
        this.workOrderRepository = workOrderRepository;
    }

    @Override
    @Transactional(readOnly = true)
    public void listWorkOrders(ListWorkOrdersRequest req, StreamObserver<ListWorkOrdersResponse> responseObserver) {
        UUID workCenterId = req.getWorkCenterId().isEmpty() ? null : UUID.fromString(req.getWorkCenterId());
        String state = req.getState().isEmpty() ? null : req.getState();

        // TODO: Use sort from gRPC
        CursorPageRequest pageReq = CursorPageRequest.of(req.getPageSize(), req.getPageToken(),
                WorkOrderSpecs.SORT_NEWEST_FIRST);

        Specification<WorkOrder> spec = WorkOrderSpecs.withFiltersAndCursor(workCenterId, state, pageReq.cursor(),
                pageReq.sortCriteriaList());

        List<WorkOrder> raw = workOrderRepository
                .findAll(spec, PageRequest.of(0, pageReq.queryLimit(), pageReq.toSort())).getContent();

        CursorPage<WorkOrder> page = CursorPage.of(raw, pageReq.pageSize(), pageReq.sortCriteriaList(),
                wo -> List.of(wo.getCreatedAt(), wo.getId()));

        ListWorkOrdersResponse.Builder response = ListWorkOrdersResponse.newBuilder();
        for (WorkOrder wo : page.items()) {
            response.addWorkOrders(toProtoWorkOrder(wo));
        }
        response.setNextPageToken(page.nextPageToken());
        responseObserver.onNext(response.build());
        responseObserver.onCompleted();
    }

    private com.factoryos.production.proto.WorkOrder toProtoWorkOrder(WorkOrder wo) {
        com.factoryos.production.proto.WorkOrder.Builder builder = com.factoryos.production.proto.WorkOrder
                .newBuilder();
        builder.setId(wo.getId().toString());
        builder.setMaterialDefinitionId(wo.getMaterialDefinitionId().toString());
        builder.setRoutingSpecId(wo.getRoutingSpecId().toString());
        builder.setWorkCenterId(wo.getWorkCenterId().toString());
        builder.setTargetQuantity(wo.getTargetQuantity());
        builder.setUnitOfMeasure(wo.getUnitOfMeasure());
        builder.setState(wo.getState());
        builder.setPriority(wo.getPriority());
        if (wo.getDescription() != null) {
            builder.setDescription(wo.getDescription());
        }
        if (wo.getDueDate() != null) {
            builder.setDueDate(instantToTimestamp(wo.getDueDate()));
        }
        if (wo.getCreatedAt() != null) {
            builder.setCreatedAt(instantToTimestamp(wo.getCreatedAt()));
        }
        if (wo.getUpdatedAt() != null) {
            builder.setUpdatedAt(instantToTimestamp(wo.getUpdatedAt()));
        }
        return builder.build();
    }

    private static Timestamp instantToTimestamp(java.time.Instant instant) {
        return Timestamp.newBuilder()
                .setSeconds(instant.getEpochSecond())
                .setNanos(instant.getNano())
                .build();
    }
}
