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

// mockRoutingSpecRepo implements db.ProductRoutingSpecRepository for testing.
type mockRoutingSpecRepo struct {
	specs  map[string]*db.ProductRoutingSpec
	nextID int
}

func newMockRoutingSpecRepo() *mockRoutingSpecRepo {
	return &mockRoutingSpecRepo{specs: make(map[string]*db.ProductRoutingSpec), nextID: 1}
}

func (m *mockRoutingSpecRepo) CreateRoutingSpec(_ context.Context, materialDefinitionID, version string, description *string) (*db.ProductRoutingSpec, error) {
	id := fmt.Sprintf("rs-%d", m.nextID)
	m.nextID++
	now := time.Now()
	spec := &db.ProductRoutingSpec{
		ID: id, MaterialDefinitionID: materialDefinitionID,
		Version: version, Description: description,
		CreatedAt: now, UpdatedAt: now,
	}
	m.specs[spec.ID] = spec
	return spec, nil
}

func (m *mockRoutingSpecRepo) GetRoutingSpec(_ context.Context, id string) (*db.ProductRoutingSpec, error) {
	spec, ok := m.specs[id]
	if !ok {
		return nil, nil
	}
	return spec, nil
}

func (m *mockRoutingSpecRepo) ListRoutingSpecs(_ context.Context, materialDefinitionID string) ([]*db.ProductRoutingSpec, error) {
	var result []*db.ProductRoutingSpec
	for _, spec := range m.specs {
		if materialDefinitionID == "" || spec.MaterialDefinitionID == materialDefinitionID {
			result = append(result, spec)
		}
	}
	return result, nil
}

// newTestRoutingSpecServer creates a ProductRoutingSpecServer with all happy-path mocks.
func newTestRoutingSpecServer() *ProductRoutingSpecServer {
	return NewProductRoutingSpecServer(
		newMockMaterialDefinitionRepo(),
		newMockRoutingSpecRepo(),
	)
}

// ============================================================================
// CreateRoutingSpec
// ============================================================================

func TestCreateRoutingSpec_Success(t *testing.T) {
	materials := newMockMaterialDefinitionRepo()
	specs := newMockRoutingSpecRepo()
	srv := NewProductRoutingSpecServer(materials, specs)

	class, _ := newMockMaterialClassRepo().CreateMaterialClass(context.Background(), "Finished Good", nil)
	md, _ := materials.CreateMaterialDefinition(context.Background(), class.ID, "Engine Block", "ENG-001", "pcs", nil)

	resp, err := srv.CreateRoutingSpec(context.Background(), &resourcev1.CreateRoutingSpecRequest{
		MaterialDefinitionId: md.ID,
		Version:              "v1.0",
		Description:          "Standard routing",
	})
	if err != nil {
		t.Fatalf("CreateRoutingSpec() = %v, want nil", err)
	}
	if resp.GetRoutingSpec().GetVersion() != "v1.0" {
		t.Errorf("Version = %q, want 'v1.0'", resp.GetRoutingSpec().GetVersion())
	}
	if resp.GetRoutingSpec().GetDescription() != "Standard routing" {
		t.Errorf("Description = %q, want 'Standard routing'", resp.GetRoutingSpec().GetDescription())
	}
}

func TestCreateRoutingSpec_NoDescription(t *testing.T) {
	materials := newMockMaterialDefinitionRepo()
	specs := newMockRoutingSpecRepo()
	srv := NewProductRoutingSpecServer(materials, specs)

	class, _ := newMockMaterialClassRepo().CreateMaterialClass(context.Background(), "Finished Good", nil)
	md, _ := materials.CreateMaterialDefinition(context.Background(), class.ID, "Engine Block", "ENG-001", "pcs", nil)

	resp, err := srv.CreateRoutingSpec(context.Background(), &resourcev1.CreateRoutingSpecRequest{
		MaterialDefinitionId: md.ID,
		Version:              "v1.0",
	})
	if err != nil {
		t.Fatalf("CreateRoutingSpec() = %v, want nil", err)
	}
	if resp.GetRoutingSpec().GetDescription() != "" {
		t.Errorf("Description = %q, want ''", resp.GetRoutingSpec().GetDescription())
	}
}

func TestCreateRoutingSpec_MissingMaterialDefinitionID(t *testing.T) {
	srv := newTestRoutingSpecServer()

	_, err := srv.CreateRoutingSpec(context.Background(), &resourcev1.CreateRoutingSpecRequest{
		Version: "v1.0",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreateRoutingSpec() = %v, want InvalidArgument", err)
	}
}

func TestCreateRoutingSpec_MissingVersion(t *testing.T) {
	srv := newTestRoutingSpecServer()

	_, err := srv.CreateRoutingSpec(context.Background(), &resourcev1.CreateRoutingSpecRequest{
		MaterialDefinitionId: "md-1",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreateRoutingSpec() = %v, want InvalidArgument", err)
	}
}

func TestCreateRoutingSpec_MaterialDefinitionNotFound(t *testing.T) {
	srv := NewProductRoutingSpecServer(newMockMaterialDefinitionRepo(), newMockRoutingSpecRepo())

	_, err := srv.CreateRoutingSpec(context.Background(), &resourcev1.CreateRoutingSpecRequest{
		MaterialDefinitionId: "non-existent",
		Version:              "v1.0",
	})
	if status.Code(err) != codes.NotFound {
		t.Errorf("CreateRoutingSpec() = %v, want NotFound", err)
	}
}

// ============================================================================
// GetRoutingSpec
// ============================================================================

func TestGetRoutingSpec_Success(t *testing.T) {
	materials := newMockMaterialDefinitionRepo()
	specs := newMockRoutingSpecRepo()
	srv := NewProductRoutingSpecServer(materials, specs)

	class, _ := newMockMaterialClassRepo().CreateMaterialClass(context.Background(), "Finished Good", nil)
	md, _ := materials.CreateMaterialDefinition(context.Background(), class.ID, "Engine Block", "ENG-001", "pcs", nil)

	createResp, _ := srv.CreateRoutingSpec(context.Background(), &resourcev1.CreateRoutingSpecRequest{
		MaterialDefinitionId: md.ID, Version: "v1.0",
	})

	getResp, err := srv.GetRoutingSpec(context.Background(), &resourcev1.GetRoutingSpecRequest{
		Id: createResp.GetRoutingSpec().GetId(),
	})
	if err != nil {
		t.Fatalf("GetRoutingSpec() = %v, want nil", err)
	}
	if getResp.GetRoutingSpec().GetId() != createResp.GetRoutingSpec().GetId() {
		t.Errorf("ID = %s, want %s", getResp.GetRoutingSpec().GetId(), createResp.GetRoutingSpec().GetId())
	}
}

func TestGetRoutingSpec_MissingID(t *testing.T) {
	srv := newTestRoutingSpecServer()

	_, err := srv.GetRoutingSpec(context.Background(), &resourcev1.GetRoutingSpecRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("GetRoutingSpec() = %v, want InvalidArgument", err)
	}
}

func TestGetRoutingSpec_NotFound(t *testing.T) {
	srv := newTestRoutingSpecServer()

	_, err := srv.GetRoutingSpec(context.Background(), &resourcev1.GetRoutingSpecRequest{Id: "non-existent"})
	if status.Code(err) != codes.NotFound {
		t.Errorf("GetRoutingSpec() = %v, want NotFound", err)
	}
}

// ============================================================================
// ListRoutingSpecs
// ============================================================================

func TestListRoutingSpecs_Success(t *testing.T) {
	materials := newMockMaterialDefinitionRepo()
	specs := newMockRoutingSpecRepo()
	srv := NewProductRoutingSpecServer(materials, specs)

	class, _ := newMockMaterialClassRepo().CreateMaterialClass(context.Background(), "Finished Good", nil)
	md, _ := materials.CreateMaterialDefinition(context.Background(), class.ID, "Engine Block", "ENG-001", "pcs", nil)

	_, _ = srv.CreateRoutingSpec(context.Background(), &resourcev1.CreateRoutingSpecRequest{
		MaterialDefinitionId: md.ID, Version: "v1.0",
	})
	_, _ = srv.CreateRoutingSpec(context.Background(), &resourcev1.CreateRoutingSpecRequest{
		MaterialDefinitionId: md.ID, Version: "v2.0",
	})

	resp, err := srv.ListRoutingSpecs(context.Background(), &resourcev1.ListRoutingSpecsRequest{})
	if err != nil {
		t.Fatalf("ListRoutingSpecs() = %v, want nil", err)
	}
	if len(resp.GetRoutingSpecs()) != 2 {
		t.Errorf("ListRoutingSpecs() returned %d, want 2", len(resp.GetRoutingSpecs()))
	}
}

func TestListRoutingSpecs_FilterByMaterial(t *testing.T) {
	materials := newMockMaterialDefinitionRepo()
	specs := newMockRoutingSpecRepo()
	srv := NewProductRoutingSpecServer(materials, specs)

	class, _ := newMockMaterialClassRepo().CreateMaterialClass(context.Background(), "Finished Good", nil)
	md1, _ := materials.CreateMaterialDefinition(context.Background(), class.ID, "Engine Block", "ENG-001", "pcs", nil)
	md2, _ := materials.CreateMaterialDefinition(context.Background(), class.ID, "Gearbox", "GBX-001", "pcs", nil)

	_, _ = srv.CreateRoutingSpec(context.Background(), &resourcev1.CreateRoutingSpecRequest{
		MaterialDefinitionId: md1.ID, Version: "v1.0",
	})
	_, _ = srv.CreateRoutingSpec(context.Background(), &resourcev1.CreateRoutingSpecRequest{
		MaterialDefinitionId: md2.ID, Version: "v1.0",
	})

	resp, err := srv.ListRoutingSpecs(context.Background(), &resourcev1.ListRoutingSpecsRequest{
		MaterialDefinitionId: md1.ID,
	})
	if err != nil {
		t.Fatalf("ListRoutingSpecs(filter) = %v, want nil", err)
	}
	if len(resp.GetRoutingSpecs()) != 1 {
		t.Errorf("ListRoutingSpecs(filter) returned %d, want 1", len(resp.GetRoutingSpecs()))
	}
}

func TestListRoutingSpecs_Empty(t *testing.T) {
	srv := newTestRoutingSpecServer()

	resp, err := srv.ListRoutingSpecs(context.Background(), &resourcev1.ListRoutingSpecsRequest{})
	if err != nil {
		t.Fatalf("ListRoutingSpecs() = %v, want nil", err)
	}
	if len(resp.GetRoutingSpecs()) != 0 {
		t.Errorf("ListRoutingSpecs() returned %d, want 0", len(resp.GetRoutingSpecs()))
	}
}

// ============================================================================
// Error-path tests
// ============================================================================

type mockRoutingSpecRepoErr struct{}

func (m *mockRoutingSpecRepoErr) CreateRoutingSpec(_ context.Context, _, _ string, _ *string) (*db.ProductRoutingSpec, error) {
	return nil, fmt.Errorf("simulated db error")
}
func (m *mockRoutingSpecRepoErr) GetRoutingSpec(_ context.Context, _ string) (*db.ProductRoutingSpec, error) {
	return nil, fmt.Errorf("simulated db error")
}
func (m *mockRoutingSpecRepoErr) ListRoutingSpecs(_ context.Context, _ string) ([]*db.ProductRoutingSpec, error) {
	return nil, fmt.Errorf("simulated db error")
}

func TestCreateRoutingSpec_RepoError(t *testing.T) {
	materials := newMockMaterialDefinitionRepo()
	srv := NewProductRoutingSpecServer(materials, &mockRoutingSpecRepoErr{})

	class, _ := newMockMaterialClassRepo().CreateMaterialClass(context.Background(), "Finished Good", nil)
	md, _ := materials.CreateMaterialDefinition(context.Background(), class.ID, "Engine Block", "ENG-001", "pcs", nil)

	_, err := srv.CreateRoutingSpec(context.Background(), &resourcev1.CreateRoutingSpecRequest{
		MaterialDefinitionId: md.ID, Version: "v1.0",
	})
	if status.Code(err) != codes.Internal {
		t.Errorf("CreateRoutingSpec() = %v, want Internal", err)
	}
}

func TestGetRoutingSpec_RepoError(t *testing.T) {
	srv := NewProductRoutingSpecServer(newMockMaterialDefinitionRepo(), &mockRoutingSpecRepoErr{})

	_, err := srv.GetRoutingSpec(context.Background(), &resourcev1.GetRoutingSpecRequest{Id: "rs-1"})
	if status.Code(err) != codes.Internal {
		t.Errorf("GetRoutingSpec() = %v, want Internal", err)
	}
}

func TestListRoutingSpecs_RepoError(t *testing.T) {
	srv := NewProductRoutingSpecServer(newMockMaterialDefinitionRepo(), &mockRoutingSpecRepoErr{})

	_, err := srv.ListRoutingSpecs(context.Background(), &resourcev1.ListRoutingSpecsRequest{})
	if status.Code(err) != codes.Internal {
		t.Errorf("ListRoutingSpecs() = %v, want Internal", err)
	}
}
