package server

import (
	"encoding/json"

	"google.golang.org/protobuf/types/known/timestamppb"

	resourcev1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/resource/v1"
	"github.com/iamraydoan/factoryos/services/resource-service/db"
)

// statusUnknown is the sentinel value for unknown status from proto enum conversion.
const statusUnknown = "unknown"

// ============================================================================
// Proto Conversion Helpers
// ============================================================================

// toProtoWorkUnit converts a domain WorkUnit to a proto WorkUnit.
func toProtoWorkUnit(wu *db.WorkUnit) *resourcev1.WorkUnit {
	if wu == nil {
		return nil
	}

	return &resourcev1.WorkUnit{
		Id:              wu.ID,
		WorkCenterId:    wu.WorkCenterID,
		Name:            wu.Name,
		Status:          toProtoStatus(wu.Status),
		PhysicalAssetId: derefString(wu.PhysicalAssetID),
		CreatedAt:       timestamppb.New(wu.CreatedAt),
		UpdatedAt:       timestamppb.New(wu.UpdatedAt),
	}
}

// toProtoEquipmentClass converts a domain EquipmentClass to a proto EquipmentClass.
func toProtoEquipmentClass(ec *db.EquipmentClass) *resourcev1.EquipmentClass {
	if ec == nil {
		return nil
	}

	return &resourcev1.EquipmentClass{
		Id:          ec.ID,
		Name:        ec.Name,
		Description: derefString(ec.Description),
		CreatedAt:   timestamppb.New(ec.CreatedAt),
		UpdatedAt:   timestamppb.New(ec.UpdatedAt),
	}
}

// toProtoCapability converts a domain WorkUnitCapability to a proto WorkUnitCapability.
func toProtoCapability(cap *db.WorkUnitCapability) *resourcev1.WorkUnitCapability {
	if cap == nil {
		return nil
	}

	propsJSON := "{}"
	if cap.Properties != nil {
		if b, err := json.Marshal(cap.Properties); err == nil {
			propsJSON = string(b)
		}
	}

	return &resourcev1.WorkUnitCapability{
		Id:               cap.ID,
		WorkUnitId:       cap.WorkUnitID,
		EquipmentClassId: cap.EquipmentClassID,
		PropertiesJson:   propsJSON,
		CreatedAt:        timestamppb.New(cap.CreatedAt),
		UpdatedAt:        timestamppb.New(cap.UpdatedAt),
	}
}

// toProtoPhysicalAsset converts a domain PhysicalAsset to a proto PhysicalAsset.
func toProtoPhysicalAsset(pa *db.PhysicalAsset) *resourcev1.PhysicalAsset {
	if pa == nil {
		return nil
	}

	proto := &resourcev1.PhysicalAsset{
		Id:           pa.ID,
		Name:         pa.Name,
		SerialNumber: pa.SerialNumber,
		Manufacturer: pa.Manufacturer,
		Model:        pa.Model,
		AssetType:    pa.AssetType,
		Status:       toProtoAssetStatus(pa.Status),
		CreatedAt:    timestamppb.New(pa.CreatedAt),
		UpdatedAt:    timestamppb.New(pa.UpdatedAt),
	}
	if pa.InstalledAt != nil {
		proto.InstalledAt = timestamppb.New(*pa.InstalledAt)
	}
	return proto
}

// toProtoInstallation converts a domain PhysicalAssetInstallation to a proto PhysicalAssetInstallation.
func toProtoInstallation(inst *db.PhysicalAssetInstallation) *resourcev1.PhysicalAssetInstallation {
	if inst == nil {
		return nil
	}

	proto := &resourcev1.PhysicalAssetInstallation{
		Id:              inst.ID,
		PhysicalAssetId: inst.PhysicalAssetID,
		WorkUnitId:      inst.WorkUnitID,
		InstalledAt:     timestamppb.New(inst.InstalledAt),
		CreatedAt:       timestamppb.New(inst.CreatedAt),
		UpdatedAt:       timestamppb.New(inst.UpdatedAt),
	}
	if inst.RemovedAt != nil {
		proto.RemovedAt = timestamppb.New(*inst.RemovedAt)
	}
	return proto
}

// toProtoPersonClass converts a domain PersonClass to a proto PersonClass.
func toProtoPersonClass(pc *db.PersonClass) *resourcev1.PersonClass {
	if pc == nil {
		return nil
	}

	return &resourcev1.PersonClass{
		Id:          pc.ID,
		Name:        pc.Name,
		Description: derefString(pc.Description),
		CreatedAt:   timestamppb.New(pc.CreatedAt),
		UpdatedAt:   timestamppb.New(pc.UpdatedAt),
	}
}

// toProtoPerson converts a domain Person to a proto Person.
func toProtoPerson(p *db.Person) *resourcev1.Person {
	if p == nil {
		return nil
	}

	return &resourcev1.Person{
		Id:            p.ID,
		PersonClassId: p.PersonClassID,
		EmployeeId:    p.EmployeeID,
		FirstName:     p.FirstName,
		LastName:      p.LastName,
		Email:         derefString(p.Email),
		Status:        toProtoPersonStatus(p.Status),
		CreatedAt:     timestamppb.New(p.CreatedAt),
		UpdatedAt:     timestamppb.New(p.UpdatedAt),
	}
}

// ============================================================================
// Status Conversion Helpers
// ============================================================================

// toProtoStatus converts a DB status string to a proto WorkUnitStatus enum.
func toProtoStatus(s string) resourcev1.WorkUnitStatus {
	switch s {
	case StatusAvailable:
		return resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_AVAILABLE
	case StatusAllocated:
		return resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_ALLOCATED
	case StatusInProduction:
		return resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_IN_PRODUCTION
	case StatusFaulted:
		return resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_FAULTED
	default:
		return resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_UNSPECIFIED
	}
}

// fromProtoStatus converts a proto WorkUnitStatus enum to a DB status string.
func fromProtoStatus(s resourcev1.WorkUnitStatus) string {
	switch s {
	case resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_AVAILABLE:
		return StatusAvailable
	case resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_ALLOCATED:
		return StatusAllocated
	case resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_IN_PRODUCTION:
		return StatusInProduction
	case resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_FAULTED:
		return StatusFaulted
	default:
		return statusUnknown
	}
}

// toProtoAssetStatus converts a DB status string to a proto PhysicalAssetStatus enum.
func toProtoAssetStatus(s string) resourcev1.PhysicalAssetStatus {
	switch s {
	case AssetStatusActive:
		return resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_ACTIVE
	case AssetStatusFaulted:
		return resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_FAULTED
	case AssetStatusUnderMaintenance:
		return resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_UNDER_MAINTENANCE
	case AssetStatusDecommissioned:
		return resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_DECOMMISSIONED
	default:
		return resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_UNSPECIFIED
	}
}

// fromProtoAssetStatus converts a proto PhysicalAssetStatus enum to a DB status string.
func fromProtoAssetStatus(s resourcev1.PhysicalAssetStatus) string {
	switch s {
	case resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_ACTIVE:
		return AssetStatusActive
	case resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_FAULTED:
		return AssetStatusFaulted
	case resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_UNDER_MAINTENANCE:
		return AssetStatusUnderMaintenance
	case resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_DECOMMISSIONED:
		return AssetStatusDecommissioned
	default:
		return statusUnknown
	}
}

// toProtoPersonStatus converts a DB status string to a proto PersonStatus enum.
func toProtoPersonStatus(s string) resourcev1.PersonStatus {
	switch s {
	case "active":
		return resourcev1.PersonStatus_PERSON_STATUS_ACTIVE
	case "inactive":
		return resourcev1.PersonStatus_PERSON_STATUS_INACTIVE
	case "on_leave":
		return resourcev1.PersonStatus_PERSON_STATUS_ON_LEAVE
	default:
		return resourcev1.PersonStatus_PERSON_STATUS_UNSPECIFIED
	}
}

// fromProtoPersonStatus converts a proto PersonStatus enum to a DB status string.
func fromProtoPersonStatus(s resourcev1.PersonStatus) string {
	switch s {
	case resourcev1.PersonStatus_PERSON_STATUS_ACTIVE:
		return "active"
	case resourcev1.PersonStatus_PERSON_STATUS_INACTIVE:
		return "inactive"
	case resourcev1.PersonStatus_PERSON_STATUS_ON_LEAVE:
		return "on_leave"
	default:
		return statusUnknown
	}
}

// ============================================================================
// Utility Helpers
// ============================================================================

// derefString safely dereferences a *string, returns "" if nil.
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
