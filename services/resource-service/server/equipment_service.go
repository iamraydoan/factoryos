// Package server implements the gRPC handlers for the Resource Service.
//
// equipment_service.go provides a combined EquipmentServiceServer adapter
// that delegates to the focused per-domain server structs. This is needed
// because the proto defines a single EquipmentService containing all RPCs,
// while the implementation is split into WorkUnitServer, EquipmentClassServer,
// CapabilityServer, PhysicalAssetServer, and InstallationServer.
package server

import (
	"context"

	resourcev1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/resource/v1"
	"github.com/iamraydoan/factoryos/services/resource-service/db"
)

// EquipmentService implements resourcev1.EquipmentServiceServer by delegating
// to focused per-domain server structs. Each domain server owns only the
// repository interfaces it needs.
type EquipmentService struct {
	resourcev1.UnimplementedEquipmentServiceServer

	WorkUnits     *WorkUnitServer
	Classes       *EquipmentClassServer
	Capabilities  *CapabilityServer
	Assets        *PhysicalAssetServer
	Installations *InstallationServer
}

// NewEquipmentService creates a combined EquipmentService from a single
// PostgresEquipmentRepository that implements all domain interfaces.
func NewEquipmentService(repo *db.PostgresEquipmentRepository) *EquipmentService {
	return &EquipmentService{
		WorkUnits:     NewWorkUnitServer(repo),
		Classes:       NewEquipmentClassServer(repo),
		Capabilities:  NewCapabilityServer(repo, repo, repo),
		Assets:        NewPhysicalAssetServer(repo),
		Installations: NewInstallationServer(repo, repo, repo),
	}
}

// --- Work Unit delegation ---

func (s *EquipmentService) CreateWorkUnit(ctx context.Context, req *resourcev1.CreateWorkUnitRequest) (*resourcev1.CreateWorkUnitResponse, error) {
	return s.WorkUnits.CreateWorkUnit(ctx, req)
}

func (s *EquipmentService) GetWorkUnit(ctx context.Context, req *resourcev1.GetWorkUnitRequest) (*resourcev1.GetWorkUnitResponse, error) {
	return s.WorkUnits.GetWorkUnit(ctx, req)
}

func (s *EquipmentService) ListWorkUnits(ctx context.Context, req *resourcev1.ListWorkUnitsRequest) (*resourcev1.ListWorkUnitsResponse, error) {
	return s.WorkUnits.ListWorkUnits(ctx, req)
}

func (s *EquipmentService) UpdateWorkUnitStatus(ctx context.Context, req *resourcev1.UpdateWorkUnitStatusRequest) (*resourcev1.UpdateWorkUnitStatusResponse, error) {
	return s.WorkUnits.UpdateWorkUnitStatus(ctx, req)
}

// --- Equipment Class delegation ---

func (s *EquipmentService) CreateEquipmentClass(ctx context.Context, req *resourcev1.CreateEquipmentClassRequest) (*resourcev1.CreateEquipmentClassResponse, error) {
	return s.Classes.CreateEquipmentClass(ctx, req)
}

func (s *EquipmentService) GetEquipmentClass(ctx context.Context, req *resourcev1.GetEquipmentClassRequest) (*resourcev1.GetEquipmentClassResponse, error) {
	return s.Classes.GetEquipmentClass(ctx, req)
}

func (s *EquipmentService) ListEquipmentClasses(ctx context.Context, req *resourcev1.ListEquipmentClassesRequest) (*resourcev1.ListEquipmentClassesResponse, error) {
	return s.Classes.ListEquipmentClasses(ctx, req)
}

// --- Capability delegation ---

func (s *EquipmentService) AssignCapability(ctx context.Context, req *resourcev1.AssignCapabilityRequest) (*resourcev1.AssignCapabilityResponse, error) {
	return s.Capabilities.AssignCapability(ctx, req)
}

func (s *EquipmentService) ListWorkUnitCapabilities(ctx context.Context, req *resourcev1.ListWorkUnitCapabilitiesRequest) (*resourcev1.ListWorkUnitCapabilitiesResponse, error) {
	return s.Capabilities.ListWorkUnitCapabilities(ctx, req)
}

func (s *EquipmentService) RemoveCapability(ctx context.Context, req *resourcev1.RemoveCapabilityRequest) (*resourcev1.RemoveCapabilityResponse, error) {
	return s.Capabilities.RemoveCapability(ctx, req)
}

// --- Physical Asset delegation ---

func (s *EquipmentService) CreatePhysicalAsset(ctx context.Context, req *resourcev1.CreatePhysicalAssetRequest) (*resourcev1.CreatePhysicalAssetResponse, error) {
	return s.Assets.CreatePhysicalAsset(ctx, req)
}

func (s *EquipmentService) GetPhysicalAsset(ctx context.Context, req *resourcev1.GetPhysicalAssetRequest) (*resourcev1.GetPhysicalAssetResponse, error) {
	return s.Assets.GetPhysicalAsset(ctx, req)
}

func (s *EquipmentService) ListPhysicalAssets(ctx context.Context, req *resourcev1.ListPhysicalAssetsRequest) (*resourcev1.ListPhysicalAssetsResponse, error) {
	return s.Assets.ListPhysicalAssets(ctx, req)
}

// --- Installation delegation ---

func (s *EquipmentService) InstallAsset(ctx context.Context, req *resourcev1.InstallAssetRequest) (*resourcev1.InstallAssetResponse, error) {
	return s.Installations.InstallAsset(ctx, req)
}

func (s *EquipmentService) UninstallAsset(ctx context.Context, req *resourcev1.UninstallAssetRequest) (*resourcev1.UninstallAssetResponse, error) {
	return s.Installations.UninstallAsset(ctx, req)
}

func (s *EquipmentService) GetCurrentInstallation(ctx context.Context, req *resourcev1.GetCurrentInstallationRequest) (*resourcev1.GetCurrentInstallationResponse, error) {
	return s.Installations.GetCurrentInstallation(ctx, req)
}

func (s *EquipmentService) ListInstallations(ctx context.Context, req *resourcev1.ListInstallationsRequest) (*resourcev1.ListInstallationsResponse, error) {
	return s.Installations.ListInstallations(ctx, req)
}
