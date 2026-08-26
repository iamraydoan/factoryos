package server

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	resourcev1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/resource/v1"
	"github.com/iamraydoan/factoryos/services/resource-service/db"
)

// ProductRoutingSpecServer implements Product Routing Spec gRPC RPCs.
type ProductRoutingSpecServer struct {
	resourcev1.UnimplementedEquipmentServiceServer
	materials    db.MaterialDefinitionRepository
	routingSpecs db.ProductRoutingSpecRepository
}

// NewProductRoutingSpecServer creates a new ProductRoutingSpecServer.
func NewProductRoutingSpecServer(materials db.MaterialDefinitionRepository, routingSpecs db.ProductRoutingSpecRepository) *ProductRoutingSpecServer {
	return &ProductRoutingSpecServer{
		materials:    materials,
		routingSpecs: routingSpecs,
	}
}

// CreateRoutingSpec creates a new Product Routing Spec.
func (s *ProductRoutingSpecServer) CreateRoutingSpec(ctx context.Context, req *resourcev1.CreateRoutingSpecRequest) (*resourcev1.CreateRoutingSpecResponse, error) {
	if req.GetMaterialDefinitionId() == "" {
		return nil, status.Error(codes.InvalidArgument, "material_definition_id is required")
	}
	if req.GetVersion() == "" {
		return nil, status.Error(codes.InvalidArgument, "version is required")
	}

	// Verify MaterialDefinition exists
	md, err := s.materials.GetMaterialDefinition(ctx, req.GetMaterialDefinitionId())
	if err != nil {
		log.Printf("[ProductRoutingSpecServer][ERROR] GetMaterialDefinition(%s): %v", req.GetMaterialDefinitionId(), err)
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

	spec, err := s.routingSpecs.CreateRoutingSpec(ctx, req.GetMaterialDefinitionId(), req.GetVersion(), desc)
	if err != nil {
		log.Printf("[ProductRoutingSpecServer][ERROR] CreateRoutingSpec: %v", err)
		return nil, status.Error(codes.Internal, "failed to create routing spec")
	}

	return &resourcev1.CreateRoutingSpecResponse{
		RoutingSpec: toProtoRoutingSpec(spec),
	}, nil
}

// GetRoutingSpec retrieves a Product Routing Spec by ID.
func (s *ProductRoutingSpecServer) GetRoutingSpec(ctx context.Context, req *resourcev1.GetRoutingSpecRequest) (*resourcev1.GetRoutingSpecResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	spec, err := s.routingSpecs.GetRoutingSpec(ctx, req.GetId())
	if err != nil {
		log.Printf("[ProductRoutingSpecServer][ERROR] GetRoutingSpec(%s): %v", req.GetId(), err)
		return nil, status.Error(codes.Internal, "failed to get routing spec")
	}
	if spec == nil {
		return nil, status.Error(codes.NotFound, "routing spec not found")
	}

	return &resourcev1.GetRoutingSpecResponse{
		RoutingSpec: toProtoRoutingSpec(spec),
	}, nil
}

// ListRoutingSpecs returns Product Routing Specs, optionally filtered by MaterialDefinitionID.
func (s *ProductRoutingSpecServer) ListRoutingSpecs(ctx context.Context, req *resourcev1.ListRoutingSpecsRequest) (*resourcev1.ListRoutingSpecsResponse, error) {
	specs, err := s.routingSpecs.ListRoutingSpecs(ctx, req.GetMaterialDefinitionId())
	if err != nil {
		log.Printf("[ProductRoutingSpecServer][ERROR] ListRoutingSpecs: %v", err)
		return nil, status.Error(codes.Internal, "failed to list routing specs")
	}

	protoSpecs := make([]*resourcev1.ProductRoutingSpec, len(specs))
	for i, spec := range specs {
		protoSpecs[i] = toProtoRoutingSpec(spec)
	}

	return &resourcev1.ListRoutingSpecsResponse{
		RoutingSpecs: protoSpecs,
	}, nil
}
