package server

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	resourcev1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/resource/v1"
	"github.com/iamraydoan/factoryos/services/resource-service/db"
)

// InstallationServer implements Installation gRPC RPCs.
type InstallationServer struct {
	resourcev1.UnimplementedEquipmentServiceServer
	workUnits db.WorkUnitRepository
	assets    db.PhysicalAssetRepository
	installs  db.InstallationRepository
}

// NewInstallationServer creates a new InstallationServer.
func NewInstallationServer(workUnits db.WorkUnitRepository, assets db.PhysicalAssetRepository, installs db.InstallationRepository) *InstallationServer {
	return &InstallationServer{workUnits: workUnits, assets: assets, installs: installs}
}

// InstallAsset installs a Physical Asset at a Work Unit.
func (s *InstallationServer) InstallAsset(ctx context.Context, req *resourcev1.InstallAssetRequest) (*resourcev1.InstallAssetResponse, error) {
	if req.GetPhysicalAssetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "physical_asset_id is required")
	}
	if req.GetWorkUnitId() == "" {
		return nil, status.Error(codes.InvalidArgument, "work_unit_id is required")
	}

	// Verify Physical Asset exists
	pa, err := s.assets.GetPhysicalAsset(ctx, req.GetPhysicalAssetId())
	if err != nil {
		log.Printf("[InstallationServer][ERROR] GetPhysicalAsset(%s): %v", req.GetPhysicalAssetId(), err)
		return nil, status.Error(codes.Internal, "failed to verify physical asset")
	}
	if pa == nil {
		return nil, status.Error(codes.NotFound, "physical asset not found")
	}

	// Verify Work Unit exists
	wu, err := s.workUnits.GetWorkUnit(ctx, req.GetWorkUnitId())
	if err != nil {
		log.Printf("[InstallationServer][ERROR] GetWorkUnit(%s): %v", req.GetWorkUnitId(), err)
		return nil, status.Error(codes.Internal, "failed to verify work unit")
	}
	if wu == nil {
		return nil, status.Error(codes.NotFound, "work unit not found")
	}

	inst, err := s.installs.InstallAsset(ctx, req.GetPhysicalAssetId(), req.GetWorkUnitId())
	if err != nil {
		log.Printf("[InstallationServer][ERROR] InstallAsset: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to install asset: %v", err)
	}

	return &resourcev1.InstallAssetResponse{
		Installation: toProtoInstallation(inst),
	}, nil
}

// UninstallAsset removes the currently installed Physical Asset from a Work Unit.
func (s *InstallationServer) UninstallAsset(ctx context.Context, req *resourcev1.UninstallAssetRequest) (*resourcev1.UninstallAssetResponse, error) {
	if req.GetWorkUnitId() == "" {
		return nil, status.Error(codes.InvalidArgument, "work_unit_id is required")
	}

	inst, err := s.installs.UninstallAsset(ctx, req.GetWorkUnitId())
	if err != nil {
		log.Printf("[InstallationServer][ERROR] UninstallAsset(%s): %v", req.GetWorkUnitId(), err)
		return nil, status.Error(codes.Internal, "failed to uninstall asset")
	}
	if inst == nil {
		return nil, status.Error(codes.NotFound, "no active installation found")
	}

	return &resourcev1.UninstallAssetResponse{
		Installation: toProtoInstallation(inst),
	}, nil
}

// GetCurrentInstallation returns the active installation for a Work Unit.
func (s *InstallationServer) GetCurrentInstallation(ctx context.Context, req *resourcev1.GetCurrentInstallationRequest) (*resourcev1.GetCurrentInstallationResponse, error) {
	if req.GetWorkUnitId() == "" {
		return nil, status.Error(codes.InvalidArgument, "work_unit_id is required")
	}

	inst, err := s.installs.GetCurrentInstallation(ctx, req.GetWorkUnitId())
	if err != nil {
		log.Printf("[InstallationServer][ERROR] GetCurrentInstallation(%s): %v", req.GetWorkUnitId(), err)
		return nil, status.Error(codes.Internal, "failed to get current installation")
	}
	if inst == nil {
		return nil, status.Error(codes.NotFound, "no active installation found")
	}

	return &resourcev1.GetCurrentInstallationResponse{
		Installation: toProtoInstallation(inst),
	}, nil
}

// ListInstallations returns the full installation history for a Work Unit.
func (s *InstallationServer) ListInstallations(ctx context.Context, req *resourcev1.ListInstallationsRequest) (*resourcev1.ListInstallationsResponse, error) {
	if req.GetWorkUnitId() == "" {
		return nil, status.Error(codes.InvalidArgument, "work_unit_id is required")
	}

	installations, err := s.installs.ListInstallations(ctx, req.GetWorkUnitId())
	if err != nil {
		log.Printf("[InstallationServer][ERROR] ListInstallations(%s): %v", req.GetWorkUnitId(), err)
		return nil, status.Error(codes.Internal, "failed to list installations")
	}

	protoInstalls := make([]*resourcev1.PhysicalAssetInstallation, len(installations))
	for i, inst := range installations {
		protoInstalls[i] = toProtoInstallation(inst)
	}

	return &resourcev1.ListInstallationsResponse{
		Installations: protoInstalls,
	}, nil
}
