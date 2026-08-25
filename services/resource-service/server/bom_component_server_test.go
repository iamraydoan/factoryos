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

// mockBOMComponentRepo implements db.BOMComponentRepository for testing.
type mockBOMComponentRepo struct {
	components map[string]*db.BOMComponent
	nextID     int
}

func newMockBOMComponentRepo() *mockBOMComponentRepo {
	return &mockBOMComponentRepo{components: make(map[string]*db.BOMComponent), nextID: 1}
}

func (m *mockBOMComponentRepo) AddBOMComponent(_ context.Context, bomID, materialDefinitionID, quantity, unitOfMeasure string) (*db.BOMComponent, error) {
	id := fmt.Sprintf("comp-%d", m.nextID)
	m.nextID++
	now := time.Now()
	comp := &db.BOMComponent{
		ID: id, BOMID: bomID, MaterialDefinitionID: materialDefinitionID,
		Quantity: quantity, UnitOfMeasure: unitOfMeasure,
		CreatedAt: now, UpdatedAt: now,
	}
	m.components[comp.ID] = comp
	return comp, nil
}

func (m *mockBOMComponentRepo) ListBOMComponents(_ context.Context, bomID string) ([]*db.BOMComponent, error) {
	var result []*db.BOMComponent
	for _, comp := range m.components {
		if comp.BOMID == bomID {
			result = append(result, comp)
		}
	}
	return result, nil
}

// newTestBOMComponentServer creates a BOMComponentServer with all happy-path mocks.
func newTestBOMComponentServer() *BOMComponentServer {
	return NewBOMComponentServer(
		newMockBOMRepo(),
		newMockMaterialDefinitionRepo(),
		newMockBOMComponentRepo(),
	)
}

// ============================================================================
// AddBOMComponent
// ============================================================================

func TestAddBOMComponent_Success(t *testing.T) {
	classes := newMockMaterialClassRepo()
	materials := newMockMaterialDefinitionRepo()
	boms := newMockBOMRepo()
	components := newMockBOMComponentRepo()
	srv := NewBOMComponentServer(boms, materials, components)

	class, _ := classes.CreateMaterialClass(context.Background(), "Finished Good", nil)
	parent, _ := materials.CreateMaterialDefinition(context.Background(), class.ID, "Engine Block", "ENG-001", "pcs", nil)
	child, _ := materials.CreateMaterialDefinition(context.Background(), class.ID, "Piston Set", "PST-001", "pcs", nil)
	bom, _ := boms.CreateBOM(context.Background(), parent.ID, "v1.0", nil)

	resp, err := srv.AddBOMComponent(context.Background(), &resourcev1.AddBOMComponentRequest{
		BomId:                bom.ID,
		MaterialDefinitionId: child.ID,
		Quantity:             "4",
		UnitOfMeasure:        "pcs",
	})
	if err != nil {
		t.Fatalf("AddBOMComponent() = %v, want nil", err)
	}
	if resp.GetComponent().GetQuantity() != "4" {
		t.Errorf("Quantity = %q, want '4'", resp.GetComponent().GetQuantity())
	}
	if resp.GetComponent().GetMaterialDefinitionId() != child.ID {
		t.Errorf("MaterialDefinitionId = %s, want %s", resp.GetComponent().GetMaterialDefinitionId(), child.ID)
	}
}

func TestAddBOMComponent_MissingBOMID(t *testing.T) {
	srv := newTestBOMComponentServer()

	_, err := srv.AddBOMComponent(context.Background(), &resourcev1.AddBOMComponentRequest{
		MaterialDefinitionId: "md-1", Quantity: "4", UnitOfMeasure: "pcs",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("AddBOMComponent() = %v, want InvalidArgument", err)
	}
}

func TestAddBOMComponent_MissingMaterialDefinitionID(t *testing.T) {
	srv := newTestBOMComponentServer()

	_, err := srv.AddBOMComponent(context.Background(), &resourcev1.AddBOMComponentRequest{
		BomId: "bom-1", Quantity: "4", UnitOfMeasure: "pcs",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("AddBOMComponent() = %v, want InvalidArgument", err)
	}
}

func TestAddBOMComponent_MissingQuantity(t *testing.T) {
	srv := newTestBOMComponentServer()

	_, err := srv.AddBOMComponent(context.Background(), &resourcev1.AddBOMComponentRequest{
		BomId: "bom-1", MaterialDefinitionId: "md-1", UnitOfMeasure: "pcs",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("AddBOMComponent() = %v, want InvalidArgument", err)
	}
}

func TestAddBOMComponent_MissingUnitOfMeasure(t *testing.T) {
	srv := newTestBOMComponentServer()

	_, err := srv.AddBOMComponent(context.Background(), &resourcev1.AddBOMComponentRequest{
		BomId: "bom-1", MaterialDefinitionId: "md-1", Quantity: "4",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("AddBOMComponent() = %v, want InvalidArgument", err)
	}
}

func TestAddBOMComponent_BOMNotFound(t *testing.T) {
	srv := NewBOMComponentServer(newMockBOMRepo(), newMockMaterialDefinitionRepo(), newMockBOMComponentRepo())

	_, err := srv.AddBOMComponent(context.Background(), &resourcev1.AddBOMComponentRequest{
		BomId: "non-existent", MaterialDefinitionId: "md-1",
		Quantity: "4", UnitOfMeasure: "pcs",
	})
	if status.Code(err) != codes.NotFound {
		t.Errorf("AddBOMComponent() = %v, want NotFound", err)
	}
}

func TestAddBOMComponent_MaterialDefinitionNotFound(t *testing.T) {
	boms := newMockBOMRepo()
	srv := NewBOMComponentServer(boms, newMockMaterialDefinitionRepo(), newMockBOMComponentRepo())

	bom, _ := boms.CreateBOM(context.Background(), "md-1", "v1.0", nil)

	_, err := srv.AddBOMComponent(context.Background(), &resourcev1.AddBOMComponentRequest{
		BomId: bom.ID, MaterialDefinitionId: "non-existent",
		Quantity: "4", UnitOfMeasure: "pcs",
	})
	if status.Code(err) != codes.NotFound {
		t.Errorf("AddBOMComponent() = %v, want NotFound", err)
	}
}

// ============================================================================
// ListBOMComponents
// ============================================================================

func TestListBOMComponents_Success(t *testing.T) {
	classes := newMockMaterialClassRepo()
	materials := newMockMaterialDefinitionRepo()
	boms := newMockBOMRepo()
	components := newMockBOMComponentRepo()
	srv := NewBOMComponentServer(boms, materials, components)

	class, _ := classes.CreateMaterialClass(context.Background(), "Finished Good", nil)
	parent, _ := materials.CreateMaterialDefinition(context.Background(), class.ID, "Engine Block", "ENG-001", "pcs", nil)
	child1, _ := materials.CreateMaterialDefinition(context.Background(), class.ID, "Piston Set", "PST-001", "pcs", nil)
	child2, _ := materials.CreateMaterialDefinition(context.Background(), class.ID, "Gasket Kit", "GSK-001", "pcs", nil)
	bom, _ := boms.CreateBOM(context.Background(), parent.ID, "v1.0", nil)

	_, _ = srv.AddBOMComponent(context.Background(), &resourcev1.AddBOMComponentRequest{
		BomId: bom.ID, MaterialDefinitionId: child1.ID, Quantity: "4", UnitOfMeasure: "pcs",
	})
	_, _ = srv.AddBOMComponent(context.Background(), &resourcev1.AddBOMComponentRequest{
		BomId: bom.ID, MaterialDefinitionId: child2.ID, Quantity: "1", UnitOfMeasure: "pcs",
	})

	resp, err := srv.ListBOMComponents(context.Background(), &resourcev1.ListBOMComponentsRequest{
		BomId: bom.ID,
	})
	if err != nil {
		t.Fatalf("ListBOMComponents() = %v, want nil", err)
	}
	if len(resp.GetComponents()) != 2 {
		t.Errorf("ListBOMComponents() returned %d, want 2", len(resp.GetComponents()))
	}
}

func TestListBOMComponents_MissingBOMID(t *testing.T) {
	srv := newTestBOMComponentServer()

	_, err := srv.ListBOMComponents(context.Background(), &resourcev1.ListBOMComponentsRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("ListBOMComponents() = %v, want InvalidArgument", err)
	}
}

func TestListBOMComponents_Empty(t *testing.T) {
	boms := newMockBOMRepo()
	srv := NewBOMComponentServer(boms, newMockMaterialDefinitionRepo(), newMockBOMComponentRepo())

	bom, _ := boms.CreateBOM(context.Background(), "md-1", "v1.0", nil)

	resp, err := srv.ListBOMComponents(context.Background(), &resourcev1.ListBOMComponentsRequest{
		BomId: bom.ID,
	})
	if err != nil {
		t.Fatalf("ListBOMComponents() = %v, want nil", err)
	}
	if len(resp.GetComponents()) != 0 {
		t.Errorf("ListBOMComponents() returned %d, want 0", len(resp.GetComponents()))
	}
}

// ============================================================================
// Error-path tests
// ============================================================================

type mockBOMComponentRepoErr struct{}

func (m *mockBOMComponentRepoErr) AddBOMComponent(_ context.Context, _, _, _, _ string) (*db.BOMComponent, error) {
	return nil, fmt.Errorf("simulated db error")
}
func (m *mockBOMComponentRepoErr) ListBOMComponents(_ context.Context, _ string) ([]*db.BOMComponent, error) {
	return nil, fmt.Errorf("simulated db error")
}

func TestAddBOMComponent_RepoError(t *testing.T) {
	classes := newMockMaterialClassRepo()
	materials := newMockMaterialDefinitionRepo()
	boms := newMockBOMRepo()
	srv := NewBOMComponentServer(boms, materials, &mockBOMComponentRepoErr{})

	class, _ := classes.CreateMaterialClass(context.Background(), "Finished Good", nil)
	parent, _ := materials.CreateMaterialDefinition(context.Background(), class.ID, "Engine Block", "ENG-001", "pcs", nil)
	child, _ := materials.CreateMaterialDefinition(context.Background(), class.ID, "Piston Set", "PST-001", "pcs", nil)
	bom, _ := boms.CreateBOM(context.Background(), parent.ID, "v1.0", nil)

	_, err := srv.AddBOMComponent(context.Background(), &resourcev1.AddBOMComponentRequest{
		BomId: bom.ID, MaterialDefinitionId: child.ID,
		Quantity: "4", UnitOfMeasure: "pcs",
	})
	if status.Code(err) != codes.Internal {
		t.Errorf("AddBOMComponent() = %v, want Internal", err)
	}
}

func TestListBOMComponents_RepoError(t *testing.T) {
	boms := newMockBOMRepo()
	srv := NewBOMComponentServer(boms, newMockMaterialDefinitionRepo(), &mockBOMComponentRepoErr{})

	bom, _ := boms.CreateBOM(context.Background(), "md-1", "v1.0", nil)

	_, err := srv.ListBOMComponents(context.Background(), &resourcev1.ListBOMComponentsRequest{BomId: bom.ID})
	if status.Code(err) != codes.Internal {
		t.Errorf("ListBOMComponents() = %v, want Internal", err)
	}
}
