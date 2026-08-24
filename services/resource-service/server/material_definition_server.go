package server

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	resourcev1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/resource/v1"
	"github.com/iamraydoan/factoryos/services/resource-service/db"
)

// MaterialDefinitionServer implements Material Definition gRPC RPCs.
type MaterialDefinitionServer struct {
	resourcev1.UnimplementedEquipmentServiceServer
	classes db.MaterialClassRepository
	defs    db.MaterialDefinitionRepository
}

// NewMaterialDefinitionServer creates a new MaterialDefinitionServer.
func NewMaterialDefinitionServer(classes db.MaterialClassRepository, defs db.MaterialDefinitionRepository) *MaterialDefinitionServer {
	return &MaterialDefinitionServer{
		classes: classes,
		defs:    defs,
	}
}

// CreateMaterialDefinition creates a new Material Definition.
func (s *MaterialDefinitionServer) CreateMaterialDefinition(ctx context.Context, req *resourcev1.CreateMaterialDefinitionRequest) (*resourcev1.CreateMaterialDefinitionResponse, error) {
	if req.GetMaterialClassId() == "" {
		return nil, status.Error(codes.InvalidArgument, "material_class_id is required")
	}
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.GetPartNumber() == "" {
		return nil, status.Error(codes.InvalidArgument, "part_number is required")
	}
	if req.GetUnitOfMeasure() == "" {
		return nil, status.Error(codes.InvalidArgument, "unit_of_measure is required")
	}

	// Verify MaterialClass exists
	mc, err := s.classes.GetMaterialClass(ctx, req.GetMaterialClassId())
	if err != nil {
		log.Printf("[MaterialDefinitionServer][ERROR] GetMaterialClass(%s): %v", req.GetMaterialClassId(), err)
		return nil, status.Error(codes.Internal, "failed to verify material class")
	}
	if mc == nil {
		return nil, status.Error(codes.NotFound, "material class not found")
	}

	var spec *string
	if req.GetSpecification() != "" {
		s := req.GetSpecification()
		spec = &s
	}

	md, err := s.defs.CreateMaterialDefinition(ctx, req.GetMaterialClassId(), req.GetName(), req.GetPartNumber(), req.GetUnitOfMeasure(), spec)
	if err != nil {
		log.Printf("[MaterialDefinitionServer][ERROR] CreateMaterialDefinition: %v", err)
		return nil, status.Error(codes.Internal, "failed to create material definition")
	}

	return &resourcev1.CreateMaterialDefinitionResponse{
		MaterialDefinition: toProtoMaterialDefinition(md),
	}, nil
}

// GetMaterialDefinition retrieves a Material Definition by ID.
func (s *MaterialDefinitionServer) GetMaterialDefinition(ctx context.Context, req *resourcev1.GetMaterialDefinitionRequest) (*resourcev1.GetMaterialDefinitionResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	md, err := s.defs.GetMaterialDefinition(ctx, req.GetId())
	if err != nil {
		log.Printf("[MaterialDefinitionServer][ERROR] GetMaterialDefinition(%s): %v", req.GetId(), err)
		return nil, status.Error(codes.Internal, "failed to get material definition")
	}
	if md == nil {
		return nil, status.Error(codes.NotFound, "material definition not found")
	}

	return &resourcev1.GetMaterialDefinitionResponse{
		MaterialDefinition: toProtoMaterialDefinition(md),
	}, nil
}

// ListMaterialDefinitions returns Material Definitions, optionally filtered by MaterialClassID.
func (s *MaterialDefinitionServer) ListMaterialDefinitions(ctx context.Context, req *resourcev1.ListMaterialDefinitionsRequest) (*resourcev1.ListMaterialDefinitionsResponse, error) {
	defs, err := s.defs.ListMaterialDefinitions(ctx, req.GetMaterialClassId())
	if err != nil {
		log.Printf("[MaterialDefinitionServer][ERROR] ListMaterialDefinitions: %v", err)
		return nil, status.Error(codes.Internal, "failed to list material definitions")
	}

	protoDefs := make([]*resourcev1.MaterialDefinition, len(defs))
	for i, md := range defs {
		protoDefs[i] = toProtoMaterialDefinition(md)
	}

	return &resourcev1.ListMaterialDefinitionsResponse{
		MaterialDefinitions: protoDefs,
	}, nil
}
