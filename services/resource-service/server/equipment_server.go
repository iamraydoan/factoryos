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
