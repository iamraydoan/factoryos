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

// ============================================================================
// Happy-path mock
// ============================================================================

// mockBOMRepo implements db.BOMRepository for testing.
type mockBOMRepo struct {
	boms   map[string]*db.BillOfMaterials
	nextID int
}

func newMockBOMRepo() *mockBOMRepo {
	return &mockBOMRepo{boms: make(map[string]*db.BillOfMaterials), nextID: 1}
}

func (m *mockBOMRepo) CreateBOM(_ context.Context, materialDefinitionID, version string, description *string) (*db.BillOfMaterials, error) {
	id := fmt.Sprintf("bom-%d", m.nextID)
	m.nextID++
	now := time.Now()
	bom := &db.BillOfMaterials{
		ID: id, MaterialDefinitionID: materialDefinitionID,
		Version: version, Description: description,
		CreatedAt: now, UpdatedAt: now,
	}
	m.boms[bom.ID] = bom
	return bom, nil
}

func (m *mockBOMRepo) GetBOM(_ context.Context, id string) (*db.BillOfMaterials, error) {
	bom, ok := m.boms[id]
	if !ok {
		return nil, nil
	}
	return bom, nil
}

func (m *mockBOMRepo) ListBOMs(_ context.Context, materialDefinitionID string) ([]*db.BillOfMaterials, error) {
	var result []*db.BillOfMaterials
	for _, bom := range m.boms {
		if materialDefinitionID == "" || bom.MaterialDefinitionID == materialDefinitionID {
			result = append(result, bom)
		}
	}
	return result, nil
}

// newTestBOMServer creates a BOMServer with all happy-path mocks.
func newTestBOMServer() *BOMServer {
	return NewBOMServer(
		newMockMaterialDefinitionRepo(),
		newMockBOMRepo(),
	)
}

// ============================================================================
// CreateBOM
// ============================================================================

func TestCreateBOM_Success(t *testing.T) {
	materials := newMockMaterialDefinitionRepo()
	boms := newMockBOMRepo()
	srv := NewBOMServer(materials, boms)

	class, _ := newMockMaterialClassRepo().CreateMaterialClass(context.Background(), "Finished Good", nil)
	md, _ := materials.CreateMaterialDefinition(context.Background(), class.ID, "Engine Block", "ENG-001", "pcs", nil)

	resp, err := srv.CreateBOM(context.Background(), &resourcev1.CreateBOMRequest{
		MaterialDefinitionId: md.ID,
		Version:              "v1.0",
		Description:          "Initial BOM",
	})
	if err != nil {
		t.Fatalf("CreateBOM() = %v, want nil", err)
	}
	if resp.GetBom().GetVersion() != "v1.0" {
		t.Errorf("Version = %q, want 'v1.0'", resp.GetBom().GetVersion())
	}
	if resp.GetBom().GetDescription() != "Initial BOM" {
		t.Errorf("Description = %q, want 'Initial BOM'", resp.GetBom().GetDescription())
	}
}

func TestCreateBOM_NoDescription(t *testing.T) {
	materials := newMockMaterialDefinitionRepo()
	boms := newMockBOMRepo()
	srv := NewBOMServer(materials, boms)

	class, _ := newMockMaterialClassRepo().CreateMaterialClass(context.Background(), "Finished Good", nil)
	md, _ := materials.CreateMaterialDefinition(context.Background(), class.ID, "Engine Block", "ENG-001", "pcs", nil)

	resp, err := srv.CreateBOM(context.Background(), &resourcev1.CreateBOMRequest{
		MaterialDefinitionId: md.ID,
		Version:              "v1.0",
	})
	if err != nil {
		t.Fatalf("CreateBOM() = %v, want nil", err)
	}
	if resp.GetBom().GetDescription() != "" {
		t.Errorf("Description = %q, want ''", resp.GetBom().GetDescription())
	}
}

func TestCreateBOM_MissingMaterialDefinitionID(t *testing.T) {
	srv := newTestBOMServer()

	_, err := srv.CreateBOM(context.Background(), &resourcev1.CreateBOMRequest{
		Version: "v1.0",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreateBOM() = %v, want InvalidArgument", err)
	}
}

func TestCreateBOM_MissingVersion(t *testing.T) {
	srv := newTestBOMServer()

	_, err := srv.CreateBOM(context.Background(), &resourcev1.CreateBOMRequest{
		MaterialDefinitionId: "md-1",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreateBOM() = %v, want InvalidArgument", err)
	}
}

func TestCreateBOM_MaterialDefinitionNotFound(t *testing.T) {
	srv := NewBOMServer(newMockMaterialDefinitionRepo(), newMockBOMRepo())

	_, err := srv.CreateBOM(context.Background(), &resourcev1.CreateBOMRequest{
		MaterialDefinitionId: "non-existent",
		Version:              "v1.0",
	})
	if status.Code(err) != codes.NotFound {
		t.Errorf("CreateBOM() = %v, want NotFound", err)
	}
}

// ============================================================================
// GetBOM
// ============================================================================

func TestGetBOM_Success(t *testing.T) {
	materials := newMockMaterialDefinitionRepo()
	boms := newMockBOMRepo()
	srv := NewBOMServer(materials, boms)

	class, _ := newMockMaterialClassRepo().CreateMaterialClass(context.Background(), "Finished Good", nil)
	md, _ := materials.CreateMaterialDefinition(context.Background(), class.ID, "Engine Block", "ENG-001", "pcs", nil)

	createResp, _ := srv.CreateBOM(context.Background(), &resourcev1.CreateBOMRequest{
		MaterialDefinitionId: md.ID, Version: "v1.0",
	})

	getResp, err := srv.GetBOM(context.Background(), &resourcev1.GetBOMRequest{
		Id: createResp.GetBom().GetId(),
	})
	if err != nil {
		t.Fatalf("GetBOM() = %v, want nil", err)
	}
	if getResp.GetBom().GetId() != createResp.GetBom().GetId() {
		t.Errorf("ID = %s, want %s", getResp.GetBom().GetId(), createResp.GetBom().GetId())
	}
}

func TestGetBOM_MissingID(t *testing.T) {
	srv := newTestBOMServer()

	_, err := srv.GetBOM(context.Background(), &resourcev1.GetBOMRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("GetBOM() = %v, want InvalidArgument", err)
	}
}

func TestGetBOM_NotFound(t *testing.T) {
	srv := newTestBOMServer()

	_, err := srv.GetBOM(context.Background(), &resourcev1.GetBOMRequest{Id: "non-existent"})
	if status.Code(err) != codes.NotFound {
		t.Errorf("GetBOM() = %v, want NotFound", err)
	}
}

// ============================================================================
// ListBOMs
// ============================================================================

func TestListBOMs_Success(t *testing.T) {
	materials := newMockMaterialDefinitionRepo()
	boms := newMockBOMRepo()
	srv := NewBOMServer(materials, boms)

	class, _ := newMockMaterialClassRepo().CreateMaterialClass(context.Background(), "Finished Good", nil)
	md, _ := materials.CreateMaterialDefinition(context.Background(), class.ID, "Engine Block", "ENG-001", "pcs", nil)

	_, _ = srv.CreateBOM(context.Background(), &resourcev1.CreateBOMRequest{
		MaterialDefinitionId: md.ID, Version: "v1.0",
	})
	_, _ = srv.CreateBOM(context.Background(), &resourcev1.CreateBOMRequest{
		MaterialDefinitionId: md.ID, Version: "v2.0",
	})

	resp, err := srv.ListBOMs(context.Background(), &resourcev1.ListBOMsRequest{})
	if err != nil {
		t.Fatalf("ListBOMs() = %v, want nil", err)
	}
	if len(resp.GetBoms()) != 2 {
		t.Errorf("ListBOMs() returned %d, want 2", len(resp.GetBoms()))
	}
}

func TestListBOMs_FilterByMaterial(t *testing.T) {
	materials := newMockMaterialDefinitionRepo()
	boms := newMockBOMRepo()
	srv := NewBOMServer(materials, boms)

	class, _ := newMockMaterialClassRepo().CreateMaterialClass(context.Background(), "Finished Good", nil)
	md1, _ := materials.CreateMaterialDefinition(context.Background(), class.ID, "Engine Block", "ENG-001", "pcs", nil)
	md2, _ := materials.CreateMaterialDefinition(context.Background(), class.ID, "Gearbox", "GBX-001", "pcs", nil)

	_, _ = srv.CreateBOM(context.Background(), &resourcev1.CreateBOMRequest{
		MaterialDefinitionId: md1.ID, Version: "v1.0",
	})
	_, _ = srv.CreateBOM(context.Background(), &resourcev1.CreateBOMRequest{
		MaterialDefinitionId: md2.ID, Version: "v1.0",
	})

	resp, err := srv.ListBOMs(context.Background(), &resourcev1.ListBOMsRequest{
		MaterialDefinitionId: md1.ID,
	})
	if err != nil {
		t.Fatalf("ListBOMs(filter) = %v, want nil", err)
	}
	if len(resp.GetBoms()) != 1 {
		t.Errorf("ListBOMs(filter) returned %d, want 1", len(resp.GetBoms()))
	}
}

func TestListBOMs_Empty(t *testing.T) {
	srv := newTestBOMServer()

	resp, err := srv.ListBOMs(context.Background(), &resourcev1.ListBOMsRequest{})
	if err != nil {
		t.Fatalf("ListBOMs() = %v, want nil", err)
	}
	if len(resp.GetBoms()) != 0 {
		t.Errorf("ListBOMs() returned %d, want 0", len(resp.GetBoms()))
	}
}

// ============================================================================
// Error-path tests
// ============================================================================

type mockBOMRepoErr struct{}

func (m *mockBOMRepoErr) CreateBOM(_ context.Context, _, _ string, _ *string) (*db.BillOfMaterials, error) {
	return nil, fmt.Errorf("simulated db error")
}
func (m *mockBOMRepoErr) GetBOM(_ context.Context, _ string) (*db.BillOfMaterials, error) {
	return nil, fmt.Errorf("simulated db error")
}
func (m *mockBOMRepoErr) ListBOMs(_ context.Context, _ string) ([]*db.BillOfMaterials, error) {
	return nil, fmt.Errorf("simulated db error")
}

func TestCreateBOM_RepoError(t *testing.T) {
	materials := newMockMaterialDefinitionRepo()
	srv := NewBOMServer(materials, &mockBOMRepoErr{})

	class, _ := newMockMaterialClassRepo().CreateMaterialClass(context.Background(), "Finished Good", nil)
	md, _ := materials.CreateMaterialDefinition(context.Background(), class.ID, "Engine Block", "ENG-001", "pcs", nil)

	_, err := srv.CreateBOM(context.Background(), &resourcev1.CreateBOMRequest{
		MaterialDefinitionId: md.ID, Version: "v1.0",
	})
	if status.Code(err) != codes.Internal {
		t.Errorf("CreateBOM() = %v, want Internal", err)
	}
}

func TestGetBOM_RepoError(t *testing.T) {
	srv := NewBOMServer(newMockMaterialDefinitionRepo(), &mockBOMRepoErr{})

	_, err := srv.GetBOM(context.Background(), &resourcev1.GetBOMRequest{Id: "bom-1"})
	if status.Code(err) != codes.Internal {
		t.Errorf("GetBOM() = %v, want Internal", err)
	}
}

func TestListBOMs_RepoError(t *testing.T) {
	srv := NewBOMServer(newMockMaterialDefinitionRepo(), &mockBOMRepoErr{})

	_, err := srv.ListBOMs(context.Background(), &resourcev1.ListBOMsRequest{})
	if status.Code(err) != codes.Internal {
		t.Errorf("ListBOMs() = %v, want Internal", err)
	}
}
