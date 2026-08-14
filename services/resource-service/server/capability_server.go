package server

import (
	"context"
	"encoding/json"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	resourcev1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/resource/v1"
	"github.com/iamraydoan/factoryos/services/resource-service/db"
)

// CapabilityServer implements Capability Assignment gRPC RPCs.
type CapabilityServer struct {
	resourcev1.UnimplementedEquipmentServiceServer
	workUnits db.WorkUnitRepository
	classes   db.EquipmentClassRepository
	caps      db.CapabilityRepository
}

// NewCapabilityServer creates a new CapabilityServer.
func NewCapabilityServer(workUnits db.WorkUnitRepository, classes db.EquipmentClassRepository, caps db.CapabilityRepository) *CapabilityServer {
	return &CapabilityServer{workUnits: workUnits, classes: classes, caps: caps}
}

// AssignCapability links a Work Unit to an Equipment Class.
func (s *CapabilityServer) AssignCapability(ctx context.Context, req *resourcev1.AssignCapabilityRequest) (*resourcev1.AssignCapabilityResponse, error) {
	if req.GetWorkUnitId() == "" {
		return nil, status.Error(codes.InvalidArgument, "work_unit_id is required")
	}
	if req.GetEquipmentClassId() == "" {
		return nil, status.Error(codes.InvalidArgument, "equipment_class_id is required")
	}

	// Verify Work Unit exists
	wu, err := s.workUnits.GetWorkUnit(ctx, req.GetWorkUnitId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to verify work unit: %v", err)
	}
	if wu == nil {
		return nil, status.Error(codes.NotFound, "work unit not found")
	}

	// Verify Equipment Class exists
	ec, err := s.classes.GetEquipmentClass(ctx, req.GetEquipmentClassId())
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

	cap, err := s.caps.AssignCapability(ctx, req.GetWorkUnitId(), req.GetEquipmentClassId(), props)
	if err != nil {
		log.Printf("[CapabilityServer][ERROR] AssignCapability: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to assign capability: %v", err)
	}

	return &resourcev1.AssignCapabilityResponse{
		Capability: toProtoCapability(cap),
	}, nil
}

// ListWorkUnitCapabilities returns all capabilities for a Work Unit.
func (s *CapabilityServer) ListWorkUnitCapabilities(ctx context.Context, req *resourcev1.ListWorkUnitCapabilitiesRequest) (*resourcev1.ListWorkUnitCapabilitiesResponse, error) {
	if req.GetWorkUnitId() == "" {
		return nil, status.Error(codes.InvalidArgument, "work_unit_id is required")
	}

	caps, err := s.caps.ListWorkUnitCapabilities(ctx, req.GetWorkUnitId())
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
func (s *CapabilityServer) RemoveCapability(ctx context.Context, req *resourcev1.RemoveCapabilityRequest) (*resourcev1.RemoveCapabilityResponse, error) {
	if req.GetWorkUnitId() == "" {
		return nil, status.Error(codes.InvalidArgument, "work_unit_id is required")
	}
	if req.GetEquipmentClassId() == "" {
		return nil, status.Error(codes.InvalidArgument, "equipment_class_id is required")
	}

	removed, err := s.caps.RemoveCapability(ctx, req.GetWorkUnitId(), req.GetEquipmentClassId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to remove capability: %v", err)
	}

	return &resourcev1.RemoveCapabilityResponse{
		Removed: removed,
	}, nil
}
