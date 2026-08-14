package server

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	resourcev1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/resource/v1"
	"github.com/iamraydoan/factoryos/services/resource-service/db"
)

// statusUnknown is the sentinel value for unknown status from proto enum conversion.
const statusUnknown = "unknown"

// EquipmentServer implements the gRPC EquipmentServiceServer interface.
type EquipmentServer struct {
	resourcev1.UnimplementedEquipmentServiceServer
	repo db.EquipmentRepository
}

// NewEquipmentServer creates a new EquipmentServer with the given repository.
func NewEquipmentServer(repo db.EquipmentRepository) *EquipmentServer {
	return &EquipmentServer{repo: repo}
}

// ============================================================================
// Work Unit RPCs
// ============================================================================

// CreateWorkUnit creates a new Work Unit in the specified Work Center.
func (s *EquipmentServer) CreateWorkUnit(ctx context.Context, req *resourcev1.CreateWorkUnitRequest) (*resourcev1.CreateWorkUnitResponse, error) {
	if req.GetWorkCenterId() == "" {
		return nil, status.Error(codes.InvalidArgument, "work_center_id is required")
	}
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if len(req.GetName()) > 255 {
		return nil, status.Error(codes.InvalidArgument, "name must be 255 characters or less")
	}

	wu, err := s.repo.CreateWorkUnit(ctx, req.GetWorkCenterId(), req.GetName())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create work unit: %v", err)
	}

	return &resourcev1.CreateWorkUnitResponse{
		WorkUnit: toProtoWorkUnit(wu),
	}, nil
}

// GetWorkUnit retrieves a Work Unit by ID.
func (s *EquipmentServer) GetWorkUnit(ctx context.Context, req *resourcev1.GetWorkUnitRequest) (*resourcev1.GetWorkUnitResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	wu, err := s.repo.GetWorkUnit(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get work unit: %v", err)
	}
	if wu == nil {
		return nil, status.Error(codes.NotFound, "work unit not found")
	}

	return &resourcev1.GetWorkUnitResponse{
		WorkUnit: toProtoWorkUnit(wu),
	}, nil
}

// ListWorkUnits returns all Work Units in a Work Center.
func (s *EquipmentServer) ListWorkUnits(ctx context.Context, req *resourcev1.ListWorkUnitsRequest) (*resourcev1.ListWorkUnitsResponse, error) {
	if req.GetWorkCenterId() == "" {
		return nil, status.Error(codes.InvalidArgument, "work_center_id is required")
	}

	workUnits, err := s.repo.ListWorkUnits(ctx, req.GetWorkCenterId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list work units: %v", err)
	}

	protoWorkUnits := make([]*resourcev1.WorkUnit, len(workUnits))
	for i, wu := range workUnits {
		protoWorkUnits[i] = toProtoWorkUnit(wu)
	}

	return &resourcev1.ListWorkUnitsResponse{
		WorkUnits: protoWorkUnits,
	}, nil
}

// UpdateWorkUnitStatus atomically changes a Work Unit's status after validating the transition.
func (s *EquipmentServer) UpdateWorkUnitStatus(ctx context.Context, req *resourcev1.UpdateWorkUnitStatusRequest) (*resourcev1.UpdateWorkUnitStatusResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	newStatus := fromProtoStatus(req.GetStatus())
	if newStatus == statusUnknown {
		return nil, status.Error(codes.InvalidArgument, "invalid status value")
	}

	// Get current work unit to validate the transition
	current, err := s.repo.GetWorkUnit(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get work unit: %v", err)
	}
	if current == nil {
		return nil, status.Error(codes.NotFound, "work unit not found")
	}

	// Validate state transition
	if err := ValidateTransition(current.Status, newStatus); err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	// Atomic update: only succeeds if current status matches
	updated, err := s.repo.UpdateWorkUnitStatus(ctx, req.GetId(), current.Status, newStatus)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update status: %v", err)
	}
	if updated == nil {
		return nil, status.Error(codes.Aborted, "status changed concurrently, please retry")
	}

	return &resourcev1.UpdateWorkUnitStatusResponse{
		WorkUnit: toProtoWorkUnit(updated),
	}, nil
}

// ============================================================================
// Equipment Class RPCs
// ============================================================================

// CreateEquipmentClass creates a new Equipment Class.
func (s *EquipmentServer) CreateEquipmentClass(ctx context.Context, req *resourcev1.CreateEquipmentClassRequest) (*resourcev1.CreateEquipmentClassResponse, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	var desc *string
	if req.GetDescription() != "" {
		d := req.GetDescription()
		desc = &d
	}

	ec, err := s.repo.CreateEquipmentClass(ctx, req.GetName(), desc)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create equipment class: %v", err)
	}

	return &resourcev1.CreateEquipmentClassResponse{
		EquipmentClass: toProtoEquipmentClass(ec),
	}, nil
}

// GetEquipmentClass retrieves an Equipment Class by ID.
func (s *EquipmentServer) GetEquipmentClass(ctx context.Context, req *resourcev1.GetEquipmentClassRequest) (*resourcev1.GetEquipmentClassResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	ec, err := s.repo.GetEquipmentClass(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get equipment class: %v", err)
	}
	if ec == nil {
		return nil, status.Error(codes.NotFound, "equipment class not found")
	}

	return &resourcev1.GetEquipmentClassResponse{
		EquipmentClass: toProtoEquipmentClass(ec),
	}, nil
}

// ListEquipmentClasses returns all Equipment Classes.
func (s *EquipmentServer) ListEquipmentClasses(ctx context.Context, req *resourcev1.ListEquipmentClassesRequest) (*resourcev1.ListEquipmentClassesResponse, error) {
	classes, err := s.repo.ListEquipmentClasses(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list equipment classes: %v", err)
	}

	protoClasses := make([]*resourcev1.EquipmentClass, len(classes))
	for i, ec := range classes {
		protoClasses[i] = toProtoEquipmentClass(ec)
	}

	return &resourcev1.ListEquipmentClassesResponse{
		EquipmentClasses: protoClasses,
	}, nil
}

// ============================================================================
// Capability Assignment RPCs
// ============================================================================

// AssignCapability links a Work Unit to an Equipment Class.
func (s *EquipmentServer) AssignCapability(ctx context.Context, req *resourcev1.AssignCapabilityRequest) (*resourcev1.AssignCapabilityResponse, error) {
	if req.GetWorkUnitId() == "" {
		return nil, status.Error(codes.InvalidArgument, "work_unit_id is required")
	}
	if req.GetEquipmentClassId() == "" {
		return nil, status.Error(codes.InvalidArgument, "equipment_class_id is required")
	}

	// Verify Work Unit exists
	wu, err := s.repo.GetWorkUnit(ctx, req.GetWorkUnitId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to verify work unit: %v", err)
	}
	if wu == nil {
		return nil, status.Error(codes.NotFound, "work unit not found")
	}

	// Verify Equipment Class exists
	ec, err := s.repo.GetEquipmentClass(ctx, req.GetEquipmentClassId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to verify equipment class: %v", err)
	}
	if ec == nil {
		return nil, status.Error(codes.NotFound, "equipment class not found")
	}

	// Parse properties JSON
	props := make(map[string]interface{})
	if req.GetPropertiesJson() != "" {
		if err := json.Unmarshal([]byte(req.GetPropertiesJson()), &props); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid properties_json")
		}
	}

	cap, err := s.repo.AssignCapability(ctx, req.GetWorkUnitId(), req.GetEquipmentClassId(), props)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to assign capability: %v", err)
	}

	return &resourcev1.AssignCapabilityResponse{
		Capability: toProtoCapability(cap),
	}, nil
}

// ListWorkUnitCapabilities returns all capabilities for a Work Unit.
func (s *EquipmentServer) ListWorkUnitCapabilities(ctx context.Context, req *resourcev1.ListWorkUnitCapabilitiesRequest) (*resourcev1.ListWorkUnitCapabilitiesResponse, error) {
	if req.GetWorkUnitId() == "" {
		return nil, status.Error(codes.InvalidArgument, "work_unit_id is required")
	}

	caps, err := s.repo.ListWorkUnitCapabilities(ctx, req.GetWorkUnitId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list capabilities: %v", err)
	}

	protoCaps := make([]*resourcev1.WorkUnitCapability, len(caps))
	for i, cap := range caps {
		protoCaps[i] = toProtoCapability(cap)
	}

	return &resourcev1.ListWorkUnitCapabilitiesResponse{
		Capabilities: protoCaps,
	}, nil
}

// RemoveCapability removes a Work Unit ↔ Equipment Class link.
func (s *EquipmentServer) RemoveCapability(ctx context.Context, req *resourcev1.RemoveCapabilityRequest) (*resourcev1.RemoveCapabilityResponse, error) {
	if req.GetWorkUnitId() == "" {
		return nil, status.Error(codes.InvalidArgument, "work_unit_id is required")
	}
	if req.GetEquipmentClassId() == "" {
		return nil, status.Error(codes.InvalidArgument, "equipment_class_id is required")
	}

	removed, err := s.repo.RemoveCapability(ctx, req.GetWorkUnitId(), req.GetEquipmentClassId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to remove capability: %v", err)
	}

	return &resourcev1.RemoveCapabilityResponse{
		Removed: removed,
	}, nil
}

// ============================================================================
// Physical Asset RPCs
// ============================================================================

// CreatePhysicalAsset creates a new Physical Asset.
func (s *EquipmentServer) CreatePhysicalAsset(ctx context.Context, req *resourcev1.CreatePhysicalAssetRequest) (*resourcev1.CreatePhysicalAssetResponse, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.GetSerialNumber() == "" {
		return nil, status.Error(codes.InvalidArgument, "serial_number is required")
	}

	asset := &db.PhysicalAsset{
		Name:         req.GetName(),
		SerialNumber: req.GetSerialNumber(),
		Manufacturer: req.GetManufacturer(),
		Model:        req.GetModel(),
		AssetType:    req.GetAssetType(),
		Status:       "active",
	}

	pa, err := s.repo.CreatePhysicalAsset(ctx, asset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create physical asset: %v", err)
	}

	return &resourcev1.CreatePhysicalAssetResponse{
		PhysicalAsset: toProtoPhysicalAsset(pa),
	}, nil
}

// GetPhysicalAsset retrieves a Physical Asset by ID.
func (s *EquipmentServer) GetPhysicalAsset(ctx context.Context, req *resourcev1.GetPhysicalAssetRequest) (*resourcev1.GetPhysicalAssetResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	pa, err := s.repo.GetPhysicalAsset(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get physical asset: %v", err)
	}
	if pa == nil {
		return nil, status.Error(codes.NotFound, "physical asset not found")
	}

	return &resourcev1.GetPhysicalAssetResponse{
		PhysicalAsset: toProtoPhysicalAsset(pa),
	}, nil
}

// ListPhysicalAssets returns all Physical Assets.
func (s *EquipmentServer) ListPhysicalAssets(ctx context.Context, req *resourcev1.ListPhysicalAssetsRequest) (*resourcev1.ListPhysicalAssetsResponse, error) {
	assets, err := s.repo.ListPhysicalAssets(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list physical assets: %v", err)
	}

	protoAssets := make([]*resourcev1.PhysicalAsset, len(assets))
	for i, pa := range assets {
		protoAssets[i] = toProtoPhysicalAsset(pa)
	}

	return &resourcev1.ListPhysicalAssetsResponse{
		PhysicalAssets: protoAssets,
	}, nil
}

// ============================================================================
// Installation RPCs
// ============================================================================

// InstallAsset installs a Physical Asset at a Work Unit.
func (s *EquipmentServer) InstallAsset(ctx context.Context, req *resourcev1.InstallAssetRequest) (*resourcev1.InstallAssetResponse, error) {
	if req.GetPhysicalAssetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "physical_asset_id is required")
	}
	if req.GetWorkUnitId() == "" {
		return nil, status.Error(codes.InvalidArgument, "work_unit_id is required")
	}

	// Verify Physical Asset exists
	pa, err := s.repo.GetPhysicalAsset(ctx, req.GetPhysicalAssetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to verify physical asset: %v", err)
	}
	if pa == nil {
		return nil, status.Error(codes.NotFound, "physical asset not found")
	}

	// Verify Work Unit exists
	wu, err := s.repo.GetWorkUnit(ctx, req.GetWorkUnitId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to verify work unit: %v", err)
	}
	if wu == nil {
		return nil, status.Error(codes.NotFound, "work unit not found")
	}

	inst, err := s.repo.InstallAsset(ctx, req.GetPhysicalAssetId(), req.GetWorkUnitId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to install asset: %v", err)
	}

	return &resourcev1.InstallAssetResponse{
		Installation: toProtoInstallation(inst),
	}, nil
}

// UninstallAsset removes the currently installed Physical Asset from a Work Unit.
func (s *EquipmentServer) UninstallAsset(ctx context.Context, req *resourcev1.UninstallAssetRequest) (*resourcev1.UninstallAssetResponse, error) {
	if req.GetWorkUnitId() == "" {
		return nil, status.Error(codes.InvalidArgument, "work_unit_id is required")
	}

	inst, err := s.repo.UninstallAsset(ctx, req.GetWorkUnitId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to uninstall asset: %v", err)
	}
	if inst == nil {
		return nil, status.Error(codes.NotFound, "no active installation found")
	}

	return &resourcev1.UninstallAssetResponse{
		Installation: toProtoInstallation(inst),
	}, nil
}

// GetCurrentInstallation returns the active installation for a Work Unit.
func (s *EquipmentServer) GetCurrentInstallation(ctx context.Context, req *resourcev1.GetCurrentInstallationRequest) (*resourcev1.GetCurrentInstallationResponse, error) {
	if req.GetWorkUnitId() == "" {
		return nil, status.Error(codes.InvalidArgument, "work_unit_id is required")
	}

	inst, err := s.repo.GetCurrentInstallation(ctx, req.GetWorkUnitId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get current installation: %v", err)
	}
	if inst == nil {
		return nil, status.Error(codes.NotFound, "no active installation found")
	}

	return &resourcev1.GetCurrentInstallationResponse{
		Installation: toProtoInstallation(inst),
	}, nil
}

// ListInstallations returns the full installation history for a Work Unit.
func (s *EquipmentServer) ListInstallations(ctx context.Context, req *resourcev1.ListInstallationsRequest) (*resourcev1.ListInstallationsResponse, error) {
	if req.GetWorkUnitId() == "" {
		return nil, status.Error(codes.InvalidArgument, "work_unit_id is required")
	}

	installations, err := s.repo.ListInstallations(ctx, req.GetWorkUnitId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list installations: %v", err)
	}

	protoInstalls := make([]*resourcev1.PhysicalAssetInstallation, len(installations))
	for i, inst := range installations {
		protoInstalls[i] = toProtoInstallation(inst)
	}

	return &resourcev1.ListInstallationsResponse{
		Installations: protoInstalls,
	}, nil
}

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

// ============================================================================
// Status Conversion Helpers
// ============================================================================

// toProtoStatus converts a DB status string to a proto WorkUnitStatus enum.
func toProtoStatus(status string) resourcev1.WorkUnitStatus {
	switch status {
	case "available":
		return resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_AVAILABLE
	case "allocated":
		return resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_ALLOCATED
	case "in_production":
		return resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_IN_PRODUCTION
	case "faulted":
		return resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_FAULTED
	default:
		return resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_UNSPECIFIED
	}
}

// fromProtoStatus converts a proto WorkUnitStatus enum to a DB status string.
func fromProtoStatus(status resourcev1.WorkUnitStatus) string {
	switch status {
	case resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_AVAILABLE:
		return "available"
	case resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_ALLOCATED:
		return "allocated"
	case resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_IN_PRODUCTION:
		return "in_production"
	case resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_FAULTED:
		return "faulted"
	default:
		return statusUnknown
	}
}

// toProtoAssetStatus converts a DB status string to a proto PhysicalAssetStatus enum.
func toProtoAssetStatus(status string) resourcev1.PhysicalAssetStatus {
	switch status {
	case "active":
		return resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_ACTIVE
	case "faulted":
		return resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_FAULTED
	case "under_maintenance":
		return resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_UNDER_MAINTENANCE
	case "decommissioned":
		return resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_DECOMMISSIONED
	default:
		return resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_UNSPECIFIED
	}
}

// fromProtoAssetStatus converts a proto PhysicalAssetStatus enum to a DB status string.
func fromProtoAssetStatus(status resourcev1.PhysicalAssetStatus) string {
	switch status {
	case resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_ACTIVE:
		return "active"
	case resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_FAULTED:
		return "faulted"
	case resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_UNDER_MAINTENANCE:
		return "under_maintenance"
	case resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_DECOMMISSIONED:
		return "decommissioned"
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
