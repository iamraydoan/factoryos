package server

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	resourcev1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/resource/v1"
	"github.com/iamraydoan/factoryos/services/resource-service/db"
)

// PhysicalAssetServer implements Physical Asset gRPC RPCs.
type PhysicalAssetServer struct {
	resourcev1.UnimplementedEquipmentServiceServer
	repo db.PhysicalAssetRepository
}

// NewPhysicalAssetServer creates a new PhysicalAssetServer.
func NewPhysicalAssetServer(repo db.PhysicalAssetRepository) *PhysicalAssetServer {
	return &PhysicalAssetServer{repo: repo}
}

// CreatePhysicalAsset creates a new Physical Asset.
func (s *PhysicalAssetServer) CreatePhysicalAsset(ctx context.Context, req *resourcev1.CreatePhysicalAssetRequest) (*resourcev1.CreatePhysicalAssetResponse, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.GetSerialNumber() == "" {
		return nil, status.Error(codes.InvalidArgument, "serial_number is required")
	}

	asset := &db.PhysicalAsset{
		Name:         req.GetName(),
		SerialNumber: req.GetSerialNumber(),
		Manufacturer: req.GetManufacturer(),
		Model:        req.GetModel(),
		AssetType:    req.GetAssetType(),
		Status:       "active",
	}

	pa, err := s.repo.CreatePhysicalAsset(ctx, asset)
	if err != nil {
		log.Printf("[PhysicalAssetServer][ERROR] CreatePhysicalAsset: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create physical asset: %v", err)
	}

	return &resourcev1.CreatePhysicalAssetResponse{
		PhysicalAsset: toProtoPhysicalAsset(pa),
	}, nil
}

// GetPhysicalAsset retrieves a Physical Asset by ID.
func (s *PhysicalAssetServer) GetPhysicalAsset(ctx context.Context, req *resourcev1.GetPhysicalAssetRequest) (*resourcev1.GetPhysicalAssetResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	pa, err := s.repo.GetPhysicalAsset(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get physical asset: %v", err)
	}
	if pa == nil {
		return nil, status.Error(codes.NotFound, "physical asset not found")
	}

	return &resourcev1.GetPhysicalAssetResponse{
		PhysicalAsset: toProtoPhysicalAsset(pa),
	}, nil
}

// ListPhysicalAssets returns all Physical Assets.
func (s *PhysicalAssetServer) ListPhysicalAssets(ctx context.Context, req *resourcev1.ListPhysicalAssetsRequest) (*resourcev1.ListPhysicalAssetsResponse, error) {
	assets, err := s.repo.ListPhysicalAssets(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list physical assets: %v", err)
	}

	protoAssets := make([]*resourcev1.PhysicalAsset, len(assets))
	for i, pa := range assets {
		protoAssets[i] = toProtoPhysicalAsset(pa)
	}

	return &resourcev1.ListPhysicalAssetsResponse{
		PhysicalAssets: protoAssets,
	}, nil
}
