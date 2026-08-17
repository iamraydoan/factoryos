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

// mockPhysicalAssetRepo implements db.PhysicalAssetRepository for testing.
type mockPhysicalAssetRepo struct {
	assets map[string]*db.PhysicalAsset
	nextID int
}

func newMockPhysicalAssetRepo() *mockPhysicalAssetRepo {
	return &mockPhysicalAssetRepo{assets: make(map[string]*db.PhysicalAsset), nextID: 1}
}

func (m *mockPhysicalAssetRepo) CreatePhysicalAsset(_ context.Context, asset *db.PhysicalAsset) (*db.PhysicalAsset, error) {
	id := fmt.Sprintf("pa-%d", m.nextID)
	m.nextID++
	pa := &db.PhysicalAsset{
		ID: id, Name: asset.Name, SerialNumber: asset.SerialNumber,
		Manufacturer: asset.Manufacturer, Model: asset.Model,
		AssetType: asset.AssetType, Status: "active",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	m.assets[pa.ID] = pa
	return pa, nil
}

func (m *mockPhysicalAssetRepo) GetPhysicalAsset(_ context.Context, id string) (*db.PhysicalAsset, error) {
	pa, ok := m.assets[id]
	if !ok {
		return nil, nil
	}
	return pa, nil
}

func (m *mockPhysicalAssetRepo) ListPhysicalAssets(_ context.Context) ([]*db.PhysicalAsset, error) {
	assets := make([]*db.PhysicalAsset, 0, len(m.assets))
	for _, pa := range m.assets {
		assets = append(assets, pa)
	}
	return assets, nil
}

func TestCreatePhysicalAsset_Success(t *testing.T) {
	srv := NewPhysicalAssetServer(newMockPhysicalAssetRepo())

	resp, err := srv.CreatePhysicalAsset(context.Background(), &resourcev1.CreatePhysicalAssetRequest{
		Name: "Haas VF-2", SerialNumber: "SN-48291",
		Manufacturer: "Haas Automation", Model: "VF-2", AssetType: "CNC Mill",
	})
	if err != nil {
		t.Fatalf("CreatePhysicalAsset() = %v, want nil", err)
	}
	pa := resp.GetPhysicalAsset()
	if pa.GetName() != "Haas VF-2" {
		t.Errorf("Name = %q, want %q", pa.GetName(), "Haas VF-2")
	}
	if pa.GetSerialNumber() != "SN-48291" {
		t.Errorf("SerialNumber = %q, want %q", pa.GetSerialNumber(), "SN-48291")
	}
	if pa.GetStatus() != resourcev1.PhysicalAssetStatus_PHYSICAL_ASSET_STATUS_ACTIVE {
		t.Errorf("Status = %v, want ACTIVE", pa.GetStatus())
	}
}

func TestCreatePhysicalAsset_MissingName(t *testing.T) {
	srv := NewPhysicalAssetServer(newMockPhysicalAssetRepo())

	_, err := srv.CreatePhysicalAsset(context.Background(), &resourcev1.CreatePhysicalAssetRequest{
		SerialNumber: "SN-48291",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreatePhysicalAsset() = %v, want InvalidArgument", err)
	}
}

func TestCreatePhysicalAsset_MissingSerialNumber(t *testing.T) {
	srv := NewPhysicalAssetServer(newMockPhysicalAssetRepo())

	_, err := srv.CreatePhysicalAsset(context.Background(), &resourcev1.CreatePhysicalAssetRequest{
		Name: "Haas VF-2",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreatePhysicalAsset() = %v, want InvalidArgument", err)
	}
}

func TestGetPhysicalAsset_Success(t *testing.T) {
	repo := newMockPhysicalAssetRepo()
	srv := NewPhysicalAssetServer(repo)

	pa, _ := repo.CreatePhysicalAsset(context.Background(), &db.PhysicalAsset{
		Name: "Haas VF-2", SerialNumber: "SN-48291",
	})

	resp, err := srv.GetPhysicalAsset(context.Background(), &resourcev1.GetPhysicalAssetRequest{Id: pa.ID})
	if err != nil {
		t.Fatalf("GetPhysicalAsset() = %v, want nil", err)
	}
	if resp.GetPhysicalAsset().GetId() != pa.ID {
		t.Errorf("ID = %s, want %s", resp.GetPhysicalAsset().GetId(), pa.ID)
	}
}

func TestGetPhysicalAsset_MissingID(t *testing.T) {
	srv := NewPhysicalAssetServer(newMockPhysicalAssetRepo())

	_, err := srv.GetPhysicalAsset(context.Background(), &resourcev1.GetPhysicalAssetRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("GetPhysicalAsset() = %v, want InvalidArgument", err)
	}
}

func TestGetPhysicalAsset_NotFound(t *testing.T) {
	srv := NewPhysicalAssetServer(newMockPhysicalAssetRepo())

	_, err := srv.GetPhysicalAsset(context.Background(), &resourcev1.GetPhysicalAssetRequest{Id: "non-existent"})
	if status.Code(err) != codes.NotFound {
		t.Errorf("GetPhysicalAsset() = %v, want NotFound", err)
	}
}

func TestListPhysicalAssets_Success(t *testing.T) {
	repo := newMockPhysicalAssetRepo()
	srv := NewPhysicalAssetServer(repo)

	repo.CreatePhysicalAsset(context.Background(), &db.PhysicalAsset{
		Name: "Haas VF-2", SerialNumber: "SN-001",
	})
	repo.CreatePhysicalAsset(context.Background(), &db.PhysicalAsset{
		Name: "Fanuc Robot", SerialNumber: "SN-002",
	})

	resp, err := srv.ListPhysicalAssets(context.Background(), &resourcev1.ListPhysicalAssetsRequest{})
	if err != nil {
		t.Fatalf("ListPhysicalAssets() = %v, want nil", err)
	}
	if len(resp.GetPhysicalAssets()) != 2 {
		t.Errorf("ListPhysicalAssets() returned %d, want 2", len(resp.GetPhysicalAssets()))
	}
}

func TestListPhysicalAssets_Empty(t *testing.T) {
	srv := NewPhysicalAssetServer(newMockPhysicalAssetRepo())

	resp, err := srv.ListPhysicalAssets(context.Background(), &resourcev1.ListPhysicalAssetsRequest{})
	if err != nil {
		t.Fatalf("ListPhysicalAssets() = %v, want nil", err)
	}
	if len(resp.GetPhysicalAssets()) != 0 {
		t.Errorf("ListPhysicalAssets() returned %d, want 0", len(resp.GetPhysicalAssets()))
	}
}

// ============================================================================
// Error-path tests: mock repos that always return errors
// ============================================================================

// mockPhysicalAssetRepoErr implements db.PhysicalAssetRepository and returns errors on all methods.
type mockPhysicalAssetRepoErr struct{}

func (m *mockPhysicalAssetRepoErr) CreatePhysicalAsset(_ context.Context, _ *db.PhysicalAsset) (*db.PhysicalAsset, error) {
	return nil, fmt.Errorf("simulated db error")
}

func (m *mockPhysicalAssetRepoErr) GetPhysicalAsset(_ context.Context, _ string) (*db.PhysicalAsset, error) {
	return nil, fmt.Errorf("simulated db error")
}

func (m *mockPhysicalAssetRepoErr) ListPhysicalAssets(_ context.Context) ([]*db.PhysicalAsset, error) {
	return nil, fmt.Errorf("simulated db error")
}

func TestCreatePhysicalAsset_RepoError(t *testing.T) {
	srv := NewPhysicalAssetServer(&mockPhysicalAssetRepoErr{})

	_, err := srv.CreatePhysicalAsset(context.Background(), &resourcev1.CreatePhysicalAssetRequest{
		Name: "Haas VF-2", SerialNumber: "SN-48291",
	})
	if status.Code(err) != codes.Internal {
		t.Errorf("CreatePhysicalAsset() = %v, want Internal", err)
	}
}

func TestGetPhysicalAsset_RepoError(t *testing.T) {
	srv := NewPhysicalAssetServer(&mockPhysicalAssetRepoErr{})

	_, err := srv.GetPhysicalAsset(context.Background(), &resourcev1.GetPhysicalAssetRequest{Id: "pa-1"})
	if status.Code(err) != codes.Internal {
		t.Errorf("GetPhysicalAsset() = %v, want Internal", err)
	}
}

func TestListPhysicalAssets_RepoError(t *testing.T) {
	srv := NewPhysicalAssetServer(&mockPhysicalAssetRepoErr{})

	_, err := srv.ListPhysicalAssets(context.Background(), &resourcev1.ListPhysicalAssetsRequest{})
	if status.Code(err) != codes.Internal {
		t.Errorf("ListPhysicalAssets() = %v, want Internal", err)
	}
}
