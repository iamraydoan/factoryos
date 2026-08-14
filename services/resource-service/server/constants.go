package server

// Work Unit status constants used throughout the resource service.
// These correspond to the work_unit_status PostgreSQL enum and the
// WorkUnitStatus proto enum defined in equipment.proto.
const (
	StatusAvailable    = "available"
	StatusAllocated    = "allocated"
	StatusInProduction = "in_production"
	StatusFaulted      = "faulted"
)

// Physical Asset status constants.
// These correspond to the physical_asset_status PostgreSQL enum and the
// PhysicalAssetStatus proto enum defined in equipment.proto.
const (
	AssetStatusActive           = "active"
	AssetStatusFaulted          = "faulted"
	AssetStatusUnderMaintenance = "under_maintenance"
	AssetStatusDecommissioned   = "decommissioned"
)
