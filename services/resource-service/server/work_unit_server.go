package server

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	resourcev1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/resource/v1"
	"github.com/iamraydoan/factoryos/services/resource-service/db"
)

// WorkUnitServer implements Work Unit gRPC RPCs.
type WorkUnitServer struct {
	resourcev1.UnimplementedEquipmentServiceServer
	repo db.WorkUnitRepository
}

// NewWorkUnitServer creates a new WorkUnitServer.
func NewWorkUnitServer(repo db.WorkUnitRepository) *WorkUnitServer {
	return &WorkUnitServer{repo: repo}
}

// CreateWorkUnit creates a new Work Unit in the specified Work Center.
func (s *WorkUnitServer) CreateWorkUnit(ctx context.Context, req *resourcev1.CreateWorkUnitRequest) (*resourcev1.CreateWorkUnitResponse, error) {
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
		log.Printf("[WorkUnitServer][ERROR] CreateWorkUnit: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create work unit: %v", err)
	}

	return &resourcev1.CreateWorkUnitResponse{
		WorkUnit: toProtoWorkUnit(wu),
	}, nil
}

// GetWorkUnit retrieves a Work Unit by ID.
func (s *WorkUnitServer) GetWorkUnit(ctx context.Context, req *resourcev1.GetWorkUnitRequest) (*resourcev1.GetWorkUnitResponse, error) {
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
func (s *WorkUnitServer) ListWorkUnits(ctx context.Context, req *resourcev1.ListWorkUnitsRequest) (*resourcev1.ListWorkUnitsResponse, error) {
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
func (s *WorkUnitServer) UpdateWorkUnitStatus(ctx context.Context, req *resourcev1.UpdateWorkUnitStatusRequest) (*resourcev1.UpdateWorkUnitStatusResponse, error) {
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
