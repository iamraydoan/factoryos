package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	resourcev1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/resource/v1"
	"github.com/iamraydoan/factoryos/services/resource-service/db"
)

// EquipmentClassServer implements Equipment Class gRPC RPCs.
type EquipmentClassServer struct {
	resourcev1.UnimplementedEquipmentServiceServer
	repo db.EquipmentClassRepository
}

// NewEquipmentClassServer creates a new EquipmentClassServer.
func NewEquipmentClassServer(repo db.EquipmentClassRepository) *EquipmentClassServer {
	return &EquipmentClassServer{repo: repo}
}

// CreateEquipmentClass creates a new Equipment Class.
func (s *EquipmentClassServer) CreateEquipmentClass(ctx context.Context, req *resourcev1.CreateEquipmentClassRequest) (*resourcev1.CreateEquipmentClassResponse, error) {
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
func (s *EquipmentClassServer) GetEquipmentClass(ctx context.Context, req *resourcev1.GetEquipmentClassRequest) (*resourcev1.GetEquipmentClassResponse, error) {
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
func (s *EquipmentClassServer) ListEquipmentClasses(ctx context.Context, req *resourcev1.ListEquipmentClassesRequest) (*resourcev1.ListEquipmentClassesResponse, error) {
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
