package server

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	resourcev1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/resource/v1"
	"github.com/iamraydoan/factoryos/services/resource-service/db"
)

// MaterialClassServer implements Material Class gRPC RPCs.
type MaterialClassServer struct {
	resourcev1.UnimplementedEquipmentServiceServer
	repo db.MaterialClassRepository
}

// NewMaterialClassServer creates a new MaterialClassServer.
func NewMaterialClassServer(repo db.MaterialClassRepository) *MaterialClassServer {
	return &MaterialClassServer{repo: repo}
}

// CreateMaterialClass creates a new Material Class.
func (s *MaterialClassServer) CreateMaterialClass(ctx context.Context, req *resourcev1.CreateMaterialClassRequest) (*resourcev1.CreateMaterialClassResponse, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	var desc *string
	if req.GetDescription() != "" {
		d := req.GetDescription()
		desc = &d
	}

	mc, err := s.repo.CreateMaterialClass(ctx, req.GetName(), desc)
	if err != nil {
		log.Printf("[MaterialClassServer][ERROR] CreateMaterialClass: %v", err)
		return nil, status.Error(codes.Internal, "failed to create material class")
	}

	return &resourcev1.CreateMaterialClassResponse{
		MaterialClass: toProtoMaterialClass(mc),
	}, nil
}

// GetMaterialClass retrieves a Material Class by ID.
func (s *MaterialClassServer) GetMaterialClass(ctx context.Context, req *resourcev1.GetMaterialClassRequest) (*resourcev1.GetMaterialClassResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	mc, err := s.repo.GetMaterialClass(ctx, req.GetId())
	if err != nil {
		log.Printf("[MaterialClassServer][ERROR] GetMaterialClass(%s): %v", req.GetId(), err)
		return nil, status.Error(codes.Internal, "failed to get material class")
	}
	if mc == nil {
		return nil, status.Error(codes.NotFound, "material class not found")
	}

	return &resourcev1.GetMaterialClassResponse{
		MaterialClass: toProtoMaterialClass(mc),
	}, nil
}

// ListMaterialClasses returns all Material Classes.
func (s *MaterialClassServer) ListMaterialClasses(ctx context.Context, req *resourcev1.ListMaterialClassesRequest) (*resourcev1.ListMaterialClassesResponse, error) {
	classes, err := s.repo.ListMaterialClasses(ctx)
	if err != nil {
		log.Printf("[MaterialClassServer][ERROR] ListMaterialClasses: %v", err)
		return nil, status.Error(codes.Internal, "failed to list material classes")
	}

	protoClasses := make([]*resourcev1.MaterialClass, len(classes))
	for i, mc := range classes {
		protoClasses[i] = toProtoMaterialClass(mc)
	}

	return &resourcev1.ListMaterialClassesResponse{
		MaterialClasses: protoClasses,
	}, nil
}
