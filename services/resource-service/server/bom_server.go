package server

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	resourcev1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/resource/v1"
	"github.com/iamraydoan/factoryos/services/resource-service/db"
)

// BOMServer implements Bill of Materials gRPC RPCs.
type BOMServer struct {
	resourcev1.UnimplementedEquipmentServiceServer
	materials db.MaterialDefinitionRepository
	boms      db.BOMRepository
}

// NewBOMServer creates a new BOMServer.
func NewBOMServer(materials db.MaterialDefinitionRepository, boms db.BOMRepository) *BOMServer {
	return &BOMServer{
		materials: materials,
		boms:      boms,
	}
}

// CreateBOM creates a new Bill of Materials.
func (s *BOMServer) CreateBOM(ctx context.Context, req *resourcev1.CreateBOMRequest) (*resourcev1.CreateBOMResponse, error) {
	if req.GetMaterialDefinitionId() == "" {
		return nil, status.Error(codes.InvalidArgument, "material_definition_id is required")
	}
	if req.GetVersion() == "" {
		return nil, status.Error(codes.InvalidArgument, "version is required")
	}

	// Verify MaterialDefinition exists
	md, err := s.materials.GetMaterialDefinition(ctx, req.GetMaterialDefinitionId())
	if err != nil {
		log.Printf("[BOMServer][ERROR] GetMaterialDefinition(%s): %v", req.GetMaterialDefinitionId(), err)
		return nil, status.Error(codes.Internal, "failed to verify material definition")
	}
	if md == nil {
		return nil, status.Error(codes.NotFound, "material definition not found")
	}

	var desc *string
	if req.GetDescription() != "" {
		d := req.GetDescription()
		desc = &d
	}

	bom, err := s.boms.CreateBOM(ctx, req.GetMaterialDefinitionId(), req.GetVersion(), desc)
	if err != nil {
		log.Printf("[BOMServer][ERROR] CreateBOM: %v", err)
		return nil, status.Error(codes.Internal, "failed to create BOM")
	}

	return &resourcev1.CreateBOMResponse{
		Bom: toProtoBOM(bom),
	}, nil
}

// GetBOM retrieves a Bill of Materials by ID.
func (s *BOMServer) GetBOM(ctx context.Context, req *resourcev1.GetBOMRequest) (*resourcev1.GetBOMResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	bom, err := s.boms.GetBOM(ctx, req.GetId())
	if err != nil {
		log.Printf("[BOMServer][ERROR] GetBOM(%s): %v", req.GetId(), err)
		return nil, status.Error(codes.Internal, "failed to get BOM")
	}
	if bom == nil {
		return nil, status.Error(codes.NotFound, "BOM not found")
	}

	return &resourcev1.GetBOMResponse{
		Bom: toProtoBOM(bom),
	}, nil
}

// ListBOMs returns Bills of Materials, optionally filtered by MaterialDefinitionID.
func (s *BOMServer) ListBOMs(ctx context.Context, req *resourcev1.ListBOMsRequest) (*resourcev1.ListBOMsResponse, error) {
	boms, err := s.boms.ListBOMs(ctx, req.GetMaterialDefinitionId())
	if err != nil {
		log.Printf("[BOMServer][ERROR] ListBOMs: %v", err)
		return nil, status.Error(codes.Internal, "failed to list BOMs")
	}

	protoBOMs := make([]*resourcev1.BillOfMaterials, len(boms))
	for i, bom := range boms {
		protoBOMs[i] = toProtoBOM(bom)
	}

	return &resourcev1.ListBOMsResponse{
		Boms: protoBOMs,
	}, nil
}
