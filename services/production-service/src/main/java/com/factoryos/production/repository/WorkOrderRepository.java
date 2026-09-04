package com.factoryos.production.repository;

import java.util.UUID;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import com.factoryos.production.entity.WorkOrder;

@Repository
public interface WorkOrderRepository extends JpaRepository<WorkOrder, UUID>, JpaSpecificationExecutor<WorkOrder> {
    @Modifying
    @Query("UPDATE WorkOrder wo SET wo.state = :newState, wo.updatedAt = CURRENT_TIMESTAMP WHERE wo.id = :id AND wo.state = :expectedState")
    int updateStateIfMatches(
            @Param("id") UUID id,
            @Param("expectedState") String expectedState,
            @Param("newState") String newState);
}
