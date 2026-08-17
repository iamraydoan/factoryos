package server

import (
	"context"
	"fmt"
	"testing"
	"time"

	resourcev1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/resource/v1"
	"github.com/iamraydoan/factoryos/services/resource-service/db"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mockInstallationRepo implements db.InstallationRepository for testing.
type mockInstallationRepo struct {
	installs map[string]*db.PhysicalAssetInstallation // keyed by work_unit_id
	nextID   int
}

func newMockInstallationRepo() *mockInstallationRepo {
	return &mockInstallationRepo{installs: make(map[string]*db.PhysicalAssetInstallation), nextID: 1}
}

func (m *mockInstallationRepo) InstallAsset(_ context.Context, physicalAssetID, workUnitID string) (*db.PhysicalAssetInstallation, error) {
	if _, exists := m.installs[workUnitID]; exists {
		return nil, fmt.Errorf("work unit already has an asset installed")
	}
	for _, inst := range m.installs {
		if inst.PhysicalAssetID == physicalAssetID {
			return nil, fmt.Errorf("asset is already installed at work unit %s", inst.WorkUnitID)
		}
	}
	now := time.Now()
	inst := &db.PhysicalAssetInstallation{
		ID: fmt.Sprintf("inst-%d", m.nextID), PhysicalAssetID: physicalAssetID,
		WorkUnitID: workUnitID, InstalledAt: now, CreatedAt: now, UpdatedAt: now,
	}
	m.nextID++
	m.installs[workUnitID] = inst
	return inst, nil
}

func (m *mockInstallationRepo) UninstallAsset(_ context.Context, workUnitID string) (*db.PhysicalAssetInstallation, error) {
	inst, ok := m.installs[workUnitID]
	if !ok {
		return nil, nil
	}
	now := time.Now()
	inst.RemovedAt = &now
	inst.UpdatedAt = now
	delete(m.installs, workUnitID)
	return inst, nil
}

func (m *mockInstallationRepo) GetCurrentInstallation(_ context.Context, workUnitID string) (*db.PhysicalAssetInstallation, error) {
	inst, ok := m.installs[workUnitID]
	if !ok {
		return nil, nil
	}
	return inst, nil
}

func (m *mockInstallationRepo) ListInstallations(_ context.Context, workUnitID string) ([]*db.PhysicalAssetInstallation, error) {
	inst, ok := m.installs[workUnitID]
	if !ok {
		return nil, nil
	}
	return []*db.PhysicalAssetInstallation{inst}, nil
}

func TestInstallAsset_Success(t *testing.T) {
	srv := NewInstallationServer(newMockWorkUnitRepo(), newMockPhysicalAssetRepo(), newMockInstallationRepo())

	wu, _ := srv.workUnits.(*mockWorkUnitRepo).CreateWorkUnit(context.Background(), "wc-1", "Station A")
	pa, _ := srv.assets.(*mockPhysicalAssetRepo).CreatePhysicalAsset(context.Background(), &db.PhysicalAsset{
		Name: "Haas VF-2", SerialNumber: "SN-48291",
	})

	resp, err := srv.InstallAsset(context.Background(), &resourcev1.InstallAssetRequest{
		PhysicalAssetId: pa.ID, WorkUnitId: wu.ID,
	})
	if err != nil {
		t.Fatalf("InstallAsset() = %v, want nil", err)
	}
	inst := resp.GetInstallation()
	if inst.GetPhysicalAssetId() != pa.ID {
		t.Errorf("PhysicalAssetId = %s, want %s", inst.GetPhysicalAssetId(), pa.ID)
	}
	if inst.GetWorkUnitId() != wu.ID {
		t.Errorf("WorkUnitId = %s, want %s", inst.GetWorkUnitId(), wu.ID)
	}
	if inst.GetInstalledAt() == nil {
		t.Error("InstalledAt = nil, want non-nil")
	}
}

func TestInstallAsset_MissingPhysicalAssetID(t *testing.T) {
	srv := NewInstallationServer(newMockWorkUnitRepo(), newMockPhysicalAssetRepo(), newMockInstallationRepo())

	_, err := srv.InstallAsset(context.Background(), &resourcev1.InstallAssetRequest{WorkUnitId: "wu-1"})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("InstallAsset() = %v, want InvalidArgument", err)
	}
}

func TestInstallAsset_MissingWorkUnitID(t *testing.T) {
	srv := NewInstallationServer(newMockWorkUnitRepo(), newMockPhysicalAssetRepo(), newMockInstallationRepo())

	_, err := srv.InstallAsset(context.Background(), &resourcev1.InstallAssetRequest{PhysicalAssetId: "pa-1"})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("InstallAsset() = %v, want InvalidArgument", err)
	}
}

func TestInstallAsset_PhysicalAssetNotFound(t *testing.T) {
	srv := NewInstallationServer(newMockWorkUnitRepo(), newMockPhysicalAssetRepo(), newMockInstallationRepo())

	wu, _ := srv.workUnits.(*mockWorkUnitRepo).CreateWorkUnit(context.Background(), "wc-1", "Station A")

	_, err := srv.InstallAsset(context.Background(), &resourcev1.InstallAssetRequest{
		PhysicalAssetId: "non-existent", WorkUnitId: wu.ID,
	})
	if status.Code(err) != codes.NotFound {
		t.Errorf("InstallAsset() = %v, want NotFound", err)
	}
}

func TestInstallAsset_WorkUnitNotFound(t *testing.T) {
	srv := NewInstallationServer(newMockWorkUnitRepo(), newMockPhysicalAssetRepo(), newMockInstallationRepo())

	pa, _ := srv.assets.(*mockPhysicalAssetRepo).CreatePhysicalAsset(context.Background(), &db.PhysicalAsset{
		Name: "Haas VF-2", SerialNumber: "SN-48291",
	})

	_, err := srv.InstallAsset(context.Background(), &resourcev1.InstallAssetRequest{
		PhysicalAssetId: pa.ID, WorkUnitId: "non-existent",
	})
	if status.Code(err) != codes.NotFound {
		t.Errorf("InstallAsset() = %v, want NotFound", err)
	}
}

func TestInstallAsset_AlreadyInstalled(t *testing.T) {
	srv := NewInstallationServer(newMockWorkUnitRepo(), newMockPhysicalAssetRepo(), newMockInstallationRepo())

	wu1, _ := srv.workUnits.(*mockWorkUnitRepo).CreateWorkUnit(context.Background(), "wc-1", "Station A")
	wu2, _ := srv.workUnits.(*mockWorkUnitRepo).CreateWorkUnit(context.Background(), "wc-1", "Station B")
	pa, _ := srv.assets.(*mockPhysicalAssetRepo).CreatePhysicalAsset(context.Background(), &db.PhysicalAsset{
		Name: "Haas VF-2", SerialNumber: "SN-48291",
	})

	_, _ = srv.InstallAsset(context.Background(), &resourcev1.InstallAssetRequest{
		PhysicalAssetId: pa.ID, WorkUnitId: wu1.ID,
	})

	_, err := srv.InstallAsset(context.Background(), &resourcev1.InstallAssetRequest{
		PhysicalAssetId: pa.ID, WorkUnitId: wu2.ID,
	})
	if status.Code(err) != codes.Internal {
		t.Errorf("InstallAsset() = %v, want Internal (already installed)", err)
	}
}

func TestUninstallAsset_Success(t *testing.T) {
	srv := NewInstallationServer(newMockWorkUnitRepo(), newMockPhysicalAssetRepo(), newMockInstallationRepo())

	wu, _ := srv.workUnits.(*mockWorkUnitRepo).CreateWorkUnit(context.Background(), "wc-1", "Station A")
	pa, _ := srv.assets.(*mockPhysicalAssetRepo).CreatePhysicalAsset(context.Background(), &db.PhysicalAsset{
		Name: "Haas VF-2", SerialNumber: "SN-48291",
	})
	_, _ = srv.InstallAsset(context.Background(), &resourcev1.InstallAssetRequest{
		PhysicalAssetId: pa.ID, WorkUnitId: wu.ID,
	})

	resp, err := srv.UninstallAsset(context.Background(), &resourcev1.UninstallAssetRequest{WorkUnitId: wu.ID})
	if err != nil {
		t.Fatalf("UninstallAsset() = %v, want nil", err)
	}
	if resp.GetInstallation().GetRemovedAt() == nil {
		t.Error("RemovedAt = nil, want non-nil")
	}
	if resp.GetInstallation().GetPhysicalAssetId() != pa.ID {
		t.Errorf("PhysicalAssetId = %s, want %s", resp.GetInstallation().GetPhysicalAssetId(), pa.ID)
	}
}

func TestUninstallAsset_MissingWorkUnitID(t *testing.T) {
	srv := NewInstallationServer(newMockWorkUnitRepo(), newMockPhysicalAssetRepo(), newMockInstallationRepo())

	_, err := srv.UninstallAsset(context.Background(), &resourcev1.UninstallAssetRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("UninstallAsset() = %v, want InvalidArgument", err)
	}
}

func TestUninstallAsset_NotFound(t *testing.T) {
	srv := NewInstallationServer(newMockWorkUnitRepo(), newMockPhysicalAssetRepo(), newMockInstallationRepo())

	wu, _ := srv.workUnits.(*mockWorkUnitRepo).CreateWorkUnit(context.Background(), "wc-1", "Station A")

	_, err := srv.UninstallAsset(context.Background(), &resourcev1.UninstallAssetRequest{WorkUnitId: wu.ID})
	if status.Code(err) != codes.NotFound {
		t.Errorf("UninstallAsset() = %v, want NotFound", err)
	}
}

func TestGetCurrentInstallation_Success(t *testing.T) {
	srv := NewInstallationServer(newMockWorkUnitRepo(), newMockPhysicalAssetRepo(), newMockInstallationRepo())

	wu, _ := srv.workUnits.(*mockWorkUnitRepo).CreateWorkUnit(context.Background(), "wc-1", "Station A")
	pa, _ := srv.assets.(*mockPhysicalAssetRepo).CreatePhysicalAsset(context.Background(), &db.PhysicalAsset{
		Name: "Haas VF-2", SerialNumber: "SN-48291",
	})
	_, _ = srv.InstallAsset(context.Background(), &resourcev1.InstallAssetRequest{
		PhysicalAssetId: pa.ID, WorkUnitId: wu.ID,
	})

	resp, err := srv.GetCurrentInstallation(context.Background(), &resourcev1.GetCurrentInstallationRequest{WorkUnitId: wu.ID})
	if err != nil {
		t.Fatalf("GetCurrentInstallation() = %v, want nil", err)
	}
	if resp.GetInstallation().GetPhysicalAssetId() != pa.ID {
		t.Errorf("PhysicalAssetId = %s, want %s", resp.GetInstallation().GetPhysicalAssetId(), pa.ID)
	}
}

func TestGetCurrentInstallation_MissingWorkUnitID(t *testing.T) {
	srv := NewInstallationServer(newMockWorkUnitRepo(), newMockPhysicalAssetRepo(), newMockInstallationRepo())

	_, err := srv.GetCurrentInstallation(context.Background(), &resourcev1.GetCurrentInstallationRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("GetCurrentInstallation() = %v, want InvalidArgument", err)
	}
}

func TestGetCurrentInstallation_NotFound(t *testing.T) {
	srv := NewInstallationServer(newMockWorkUnitRepo(), newMockPhysicalAssetRepo(), newMockInstallationRepo())

	wu, _ := srv.workUnits.(*mockWorkUnitRepo).CreateWorkUnit(context.Background(), "wc-1", "Station A")

	_, err := srv.GetCurrentInstallation(context.Background(), &resourcev1.GetCurrentInstallationRequest{WorkUnitId: wu.ID})
	if status.Code(err) != codes.NotFound {
		t.Errorf("GetCurrentInstallation() = %v, want NotFound", err)
	}
}

func TestListInstallations_Success(t *testing.T) {
	srv := NewInstallationServer(newMockWorkUnitRepo(), newMockPhysicalAssetRepo(), newMockInstallationRepo())

	wu, _ := srv.workUnits.(*mockWorkUnitRepo).CreateWorkUnit(context.Background(), "wc-1", "Station A")
	pa, _ := srv.assets.(*mockPhysicalAssetRepo).CreatePhysicalAsset(context.Background(), &db.PhysicalAsset{
		Name: "Haas VF-2", SerialNumber: "SN-48291",
	})
	_, _ = srv.InstallAsset(context.Background(), &resourcev1.InstallAssetRequest{
		PhysicalAssetId: pa.ID, WorkUnitId: wu.ID,
	})

	resp, err := srv.ListInstallations(context.Background(), &resourcev1.ListInstallationsRequest{WorkUnitId: wu.ID})
	if err != nil {
		t.Fatalf("ListInstallations() = %v, want nil", err)
	}
	if len(resp.GetInstallations()) != 1 {
		t.Errorf("ListInstallations() returned %d, want 1", len(resp.GetInstallations()))
	}
}

func TestListInstallations_MissingWorkUnitID(t *testing.T) {
	srv := NewInstallationServer(newMockWorkUnitRepo(), newMockPhysicalAssetRepo(), newMockInstallationRepo())

	_, err := srv.ListInstallations(context.Background(), &resourcev1.ListInstallationsRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("ListInstallations() = %v, want InvalidArgument", err)
	}
}

func TestListInstallations_Empty(t *testing.T) {
	srv := NewInstallationServer(newMockWorkUnitRepo(), newMockPhysicalAssetRepo(), newMockInstallationRepo())

	wu, _ := srv.workUnits.(*mockWorkUnitRepo).CreateWorkUnit(context.Background(), "wc-1", "Station A")

	resp, err := srv.ListInstallations(context.Background(), &resourcev1.ListInstallationsRequest{WorkUnitId: wu.ID})
	if err != nil {
		t.Fatalf("ListInstallations() = %v, want nil", err)
	}
	if len(resp.GetInstallations()) != 0 {
		t.Errorf("ListInstallations() returned %d, want 0", len(resp.GetInstallations()))
	}
}

// ============================================================================
// Error-path tests: mock repos that always return errors
// ============================================================================

// mockInstallationRepoErr implements db.InstallationRepository and returns errors on all methods.
type mockInstallationRepoErr struct{}

func (m *mockInstallationRepoErr) InstallAsset(_ context.Context, _, _ string) (*db.PhysicalAssetInstallation, error) {
	return nil, fmt.Errorf("simulated db error")
}

func (m *mockInstallationRepoErr) UninstallAsset(_ context.Context, _ string) (*db.PhysicalAssetInstallation, error) {
	return nil, fmt.Errorf("simulated db error")
}

func (m *mockInstallationRepoErr) GetCurrentInstallation(_ context.Context, _ string) (*db.PhysicalAssetInstallation, error) {
	return nil, fmt.Errorf("simulated db error")
}

func (m *mockInstallationRepoErr) ListInstallations(_ context.Context, _ string) ([]*db.PhysicalAssetInstallation, error) {
	return nil, fmt.Errorf("simulated db error")
}

func TestInstallAsset_AssetRepoError(t *testing.T) {
	srv := NewInstallationServer(newMockWorkUnitRepo(), &mockPhysicalAssetRepoErr{}, newMockInstallationRepo())

	wu, _ := srv.workUnits.(*mockWorkUnitRepo).CreateWorkUnit(context.Background(), "wc-1", "Station A")

	_, err := srv.InstallAsset(context.Background(), &resourcev1.InstallAssetRequest{
		PhysicalAssetId: "pa-1", WorkUnitId: wu.ID,
	})
	if status.Code(err) != codes.Internal {
		t.Errorf("InstallAsset() = %v, want Internal", err)
	}
}

func TestInstallAsset_WorkUnitRepoError(t *testing.T) {
	srv := NewInstallationServer(&mockWorkUnitRepoErr{}, newMockPhysicalAssetRepo(), newMockInstallationRepo())

	pa, _ := srv.assets.(*mockPhysicalAssetRepo).CreatePhysicalAsset(context.Background(), &db.PhysicalAsset{
		Name: "Haas VF-2", SerialNumber: "SN-48291",
	})

	_, err := srv.InstallAsset(context.Background(), &resourcev1.InstallAssetRequest{
		PhysicalAssetId: pa.ID, WorkUnitId: "wu-1",
	})
	if status.Code(err) != codes.Internal {
		t.Errorf("InstallAsset() = %v, want Internal", err)
	}
}

func TestInstallAsset_RepoError(t *testing.T) {
	srv := NewInstallationServer(newMockWorkUnitRepo(), newMockPhysicalAssetRepo(), &mockInstallationRepoErr{})

	wu, _ := srv.workUnits.(*mockWorkUnitRepo).CreateWorkUnit(context.Background(), "wc-1", "Station A")
	pa, _ := srv.assets.(*mockPhysicalAssetRepo).CreatePhysicalAsset(context.Background(), &db.PhysicalAsset{
		Name: "Haas VF-2", SerialNumber: "SN-48291",
	})

	_, err := srv.InstallAsset(context.Background(), &resourcev1.InstallAssetRequest{
		PhysicalAssetId: pa.ID, WorkUnitId: wu.ID,
	})
	if status.Code(err) != codes.Internal {
		t.Errorf("InstallAsset() = %v, want Internal", err)
	}
}

func TestUninstallAsset_RepoError(t *testing.T) {
	srv := NewInstallationServer(newMockWorkUnitRepo(), newMockPhysicalAssetRepo(), &mockInstallationRepoErr{})

	_, err := srv.UninstallAsset(context.Background(), &resourcev1.UninstallAssetRequest{WorkUnitId: "wu-1"})
	if status.Code(err) != codes.Internal {
		t.Errorf("UninstallAsset() = %v, want Internal", err)
	}
}

func TestGetCurrentInstallation_RepoError(t *testing.T) {
	srv := NewInstallationServer(newMockWorkUnitRepo(), newMockPhysicalAssetRepo(), &mockInstallationRepoErr{})

	_, err := srv.GetCurrentInstallation(context.Background(), &resourcev1.GetCurrentInstallationRequest{WorkUnitId: "wu-1"})
	if status.Code(err) != codes.Internal {
		t.Errorf("GetCurrentInstallation() = %v, want Internal", err)
	}
}

func TestListInstallations_RepoError(t *testing.T) {
	srv := NewInstallationServer(newMockWorkUnitRepo(), newMockPhysicalAssetRepo(), &mockInstallationRepoErr{})

	_, err := srv.ListInstallations(context.Background(), &resourcev1.ListInstallationsRequest{WorkUnitId: "wu-1"})
	if status.Code(err) != codes.Internal {
		t.Errorf("ListInstallations() = %v, want Internal", err)
	}
}
