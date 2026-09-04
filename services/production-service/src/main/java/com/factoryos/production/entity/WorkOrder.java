package com.factoryos.production.entity;

import java.time.Instant;
import java.util.UUID;

import jakarta.persistence.*;

import com.github.f4b6a3.uuid.UuidCreator;

/**
 * WorkOrder: the primary execution unit dispatched to the shop floor.
 * ISA-95 Reference: Part 2, Section 4 — Job Order.
 */
@Entity
@Table(name = "work_orders")
public class WorkOrder {
    @Id
    @Column(columnDefinition = "uuid")
    private UUID id;

    @Column(name = "material_definition_id", nullable = false, columnDefinition = "uuid")
    private UUID materialDefinitionId;

    @Column(name = "routing_spec_id", nullable = false, columnDefinition = "uuid")
    private UUID routingSpecId;

    @Column(name = "work_center_id", nullable = false, columnDefinition = "uuid")
    private UUID workCenterId;

    @Column(name = "target_quantity", nullable = false, length = 50)
    private String targetQuantity;

    @Column(name = "unit_of_measure", nullable = false, length = 50)
    private String unitOfMeasure;

    @Column(nullable = false, length = 20)
    private String state = "draft";

    @Column(nullable = false, length = 20)
    private String priority = "medium";

    @Column(columnDefinition = "text")
    private String description;

    @Column(name = "due_date")
    private Instant dueDate;

    @Column(name = "created_at", nullable = false, updatable = false)
    private Instant createdAt;

    @Column(name = "updated_at", nullable = false)
    private Instant updatedAt;

    @PrePersist
    protected void onCreate() {
        this.id = UuidCreator.getTimeOrderedEpoch();
        Instant now = Instant.now();
        this.createdAt = now;
        this.updatedAt = now;
    }

    @PreUpdate
    protected void onUpdate() {
        this.updatedAt = Instant.now();
    }

    public UUID getId() {
        return id;
    }

    public void setId(UUID id) {
        this.id = id;
    }

    public UUID getMaterialDefinitionId() {
        return materialDefinitionId;
    }

    public void setMaterialDefinitionId(UUID materialDefinitionId) {
        this.materialDefinitionId = materialDefinitionId;
    }

    public UUID getRoutingSpecId() {
        return routingSpecId;
    }

    public void setRoutingSpecId(UUID routingSpecId) {
        this.routingSpecId = routingSpecId;
    }

    public UUID getWorkCenterId() {
        return workCenterId;
    }

    public void setWorkCenterId(UUID workCenterId) {
        this.workCenterId = workCenterId;
    }

    public String getTargetQuantity() {
        return targetQuantity;
    }

    public void setTargetQuantity(String targetQuantity) {
        this.targetQuantity = targetQuantity;
    }

    public String getUnitOfMeasure() {
        return unitOfMeasure;
    }

    public void setUnitOfMeasure(String unitOfMeasure) {
        this.unitOfMeasure = unitOfMeasure;
    }

    public String getState() {
        return state;
    }

    public void setState(String state) {
        this.state = state;
    }

    public String getPriority() {
        return priority;
    }

    public void setPriority(String priority) {
        this.priority = priority;
    }

    public String getDescription() {
        return description;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public Instant getDueDate() {
        return dueDate;
    }

    public void setDueDate(Instant dueDate) {
        this.dueDate = dueDate;
    }

    public Instant getCreatedAt() {
        return createdAt;
    }

    public Instant getUpdatedAt() {
        return updatedAt;
    }
}
