package server

import (
	"context"

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

	// Convert proto enum to DB string (e.g., WORK_UNIT_STATUS_ALLOCATED → "allocated")
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
		// Status changed between Get and Update (race condition)
		return nil, status.Error(codes.Aborted, "status changed concurrently, please retry")
	}

	return &resourcev1.UpdateWorkUnitStatusResponse{
		WorkUnit: toProtoWorkUnit(updated),
	}, nil
}

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

// derefString safely dereferences a *string, returns "" if nil.
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

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
		return "unknown"
	}
}
