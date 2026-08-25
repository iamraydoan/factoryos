package server

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	resourcev1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/resource/v1"
	"github.com/iamraydoan/factoryos/services/resource-service/db"
)

// BOMComponentServer implements BOM Component gRPC RPCs.
type BOMComponentServer struct {
	resourcev1.UnimplementedEquipmentServiceServer
	boms       db.BOMRepository
	materials  db.MaterialDefinitionRepository
	components db.BOMComponentRepository
}

// NewBOMComponentServer creates a new BOMComponentServer.
func NewBOMComponentServer(boms db.BOMRepository, materials db.MaterialDefinitionRepository, components db.BOMComponentRepository) *BOMComponentServer {
	return &BOMComponentServer{
		boms:       boms,
		materials:  materials,
		components: components,
	}
}

// AddBOMComponent adds a component to a Bill of Materials.
func (s *BOMComponentServer) AddBOMComponent(ctx context.Context, req *resourcev1.AddBOMComponentRequest) (*resourcev1.AddBOMComponentResponse, error) {
	if req.GetBomId() == "" {
		return nil, status.Error(codes.InvalidArgument, "bom_id is required")
	}
	if req.GetMaterialDefinitionId() == "" {
		return nil, status.Error(codes.InvalidArgument, "material_definition_id is required")
	}
	if req.GetQuantity() == "" {
		return nil, status.Error(codes.InvalidArgument, "quantity is required")
	}
	if req.GetUnitOfMeasure() == "" {
		return nil, status.Error(codes.InvalidArgument, "unit_of_measure is required")
	}

	// Verify BOM exists
	bom, err := s.boms.GetBOM(ctx, req.GetBomId())
	if err != nil {
		log.Printf("[BOMComponentServer][ERROR] GetBOM(%s): %v", req.GetBomId(), err)
		return nil, status.Error(codes.Internal, "failed to verify BOM")
	}
	if bom == nil {
		return nil, status.Error(codes.NotFound, "BOM not found")
	}

	// Verify child MaterialDefinition exists
	md, err := s.materials.GetMaterialDefinition(ctx, req.GetMaterialDefinitionId())
	if err != nil {
		log.Printf("[BOMComponentServer][ERROR] GetMaterialDefinition(%s): %v", req.GetMaterialDefinitionId(), err)
		return nil, status.Error(codes.Internal, "failed to verify material definition")
	}
	if md == nil {
		return nil, status.Error(codes.NotFound, "material definition not found")
	}

	comp, err := s.components.AddBOMComponent(ctx, req.GetBomId(), req.GetMaterialDefinitionId(), req.GetQuantity(), req.GetUnitOfMeasure())
	if err != nil {
		log.Printf("[BOMComponentServer][ERROR] AddBOMComponent: %v", err)
		return nil, status.Error(codes.Internal, "failed to add BOM component")
	}

	return &resourcev1.AddBOMComponentResponse{
		Component: toProtoBOMComponent(comp),
	}, nil
}

// ListBOMComponents returns all components for a given BOM.
func (s *BOMComponentServer) ListBOMComponents(ctx context.Context, req *resourcev1.ListBOMComponentsRequest) (*resourcev1.ListBOMComponentsResponse, error) {
	if req.GetBomId() == "" {
		return nil, status.Error(codes.InvalidArgument, "bom_id is required")
	}

	components, err := s.components.ListBOMComponents(ctx, req.GetBomId())
	if err != nil {
		log.Printf("[BOMComponentServer][ERROR] ListBOMComponents(%s): %v", req.GetBomId(), err)
		return nil, status.Error(codes.Internal, "failed to list BOM components")
	}

	protoComps := make([]*resourcev1.BOMComponent, len(components))
	for i, comp := range components {
		protoComps[i] = toProtoBOMComponent(comp)
	}

	return &resourcev1.ListBOMComponentsResponse{
		Components: protoComps,
	}, nil
}
