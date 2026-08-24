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

// mockMaterialDefinitionRepo implements db.MaterialDefinitionRepository for testing.
type mockMaterialDefinitionRepo struct {
	defs   map[string]*db.MaterialDefinition
	nextID int
}

func newMockMaterialDefinitionRepo() *mockMaterialDefinitionRepo {
	return &mockMaterialDefinitionRepo{defs: make(map[string]*db.MaterialDefinition), nextID: 1}
}

func (m *mockMaterialDefinitionRepo) CreateMaterialDefinition(_ context.Context, materialClassID, name, partNumber, unitOfMeasure string, specification *string) (*db.MaterialDefinition, error) {
	id := fmt.Sprintf("md-%d", m.nextID)
	m.nextID++
	now := time.Now()
	md := &db.MaterialDefinition{
		ID: id, MaterialClassID: materialClassID, Name: name,
		PartNumber: partNumber, UnitOfMeasure: unitOfMeasure,
		Specification: specification, CreatedAt: now, UpdatedAt: now,
	}
	m.defs[md.ID] = md
	return md, nil
}

func (m *mockMaterialDefinitionRepo) GetMaterialDefinition(_ context.Context, id string) (*db.MaterialDefinition, error) {
	md, ok := m.defs[id]
	if !ok {
		return nil, nil
	}
	return md, nil
}

func (m *mockMaterialDefinitionRepo) ListMaterialDefinitions(_ context.Context, materialClassID string) ([]*db.MaterialDefinition, error) {
	var result []*db.MaterialDefinition
	for _, md := range m.defs {
		if materialClassID == "" || md.MaterialClassID == materialClassID {
			result = append(result, md)
		}
	}
	return result, nil
}

// newTestMaterialDefinitionServer creates a MaterialDefinitionServer with all happy-path mocks.
func newTestMaterialDefinitionServer() *MaterialDefinitionServer {
	return NewMaterialDefinitionServer(
		newMockMaterialClassRepo(),
		newMockMaterialDefinitionRepo(),
	)
}

// ============================================================================
// CreateMaterialDefinition
// ============================================================================

func TestCreateMaterialDefinition_Success(t *testing.T) {
	classes := newMockMaterialClassRepo()
	defs := newMockMaterialDefinitionRepo()
	srv := NewMaterialDefinitionServer(classes, defs)

	class, _ := classes.CreateMaterialClass(context.Background(), "Raw Material", nil)

	spec := `{"grade": "6061-T6", "thickness_mm": 3}`
	resp, err := srv.CreateMaterialDefinition(context.Background(), &resourcev1.CreateMaterialDefinitionRequest{
		MaterialClassId: class.ID,
		Name:            "Aluminum Sheet 3mm",
		PartNumber:      "MAT-AL-3MM",
		UnitOfMeasure:   "kg",
		Specification:   spec,
	})
	if err != nil {
		t.Fatalf("CreateMaterialDefinition() = %v, want nil", err)
	}
	if resp.GetMaterialDefinition().GetName() != "Aluminum Sheet 3mm" {
		t.Errorf("Name = %q, want 'Aluminum Sheet 3mm'", resp.GetMaterialDefinition().GetName())
	}
	if resp.GetMaterialDefinition().GetPartNumber() != "MAT-AL-3MM" {
		t.Errorf("PartNumber = %q, want 'MAT-AL-3MM'", resp.GetMaterialDefinition().GetPartNumber())
	}
	if resp.GetMaterialDefinition().GetSpecification() != spec {
		t.Errorf("Specification = %q, want %q", resp.GetMaterialDefinition().GetSpecification(), spec)
	}
}

func TestCreateMaterialDefinition_NoSpecification(t *testing.T) {
	classes := newMockMaterialClassRepo()
	defs := newMockMaterialDefinitionRepo()
	srv := NewMaterialDefinitionServer(classes, defs)

	class, _ := classes.CreateMaterialClass(context.Background(), "Finished Good", nil)

	resp, err := srv.CreateMaterialDefinition(context.Background(), &resourcev1.CreateMaterialDefinitionRequest{
		MaterialClassId: class.ID,
		Name:            "Engine Block Assembly",
		PartNumber:      "ENG-BLK-001",
		UnitOfMeasure:   "pcs",
	})
	if err != nil {
		t.Fatalf("CreateMaterialDefinition() = %v, want nil", err)
	}
	if resp.GetMaterialDefinition().GetSpecification() != "" {
		t.Errorf("Specification = %q, want ''", resp.GetMaterialDefinition().GetSpecification())
	}
}

func TestCreateMaterialDefinition_MissingMaterialClassID(t *testing.T) {
	srv := newTestMaterialDefinitionServer()

	_, err := srv.CreateMaterialDefinition(context.Background(), &resourcev1.CreateMaterialDefinitionRequest{
		Name: "Aluminum Sheet", PartNumber: "MAT-AL-001", UnitOfMeasure: "kg",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreateMaterialDefinition() = %v, want InvalidArgument", err)
	}
}

func TestCreateMaterialDefinition_MissingName(t *testing.T) {
	srv := newTestMaterialDefinitionServer()

	_, err := srv.CreateMaterialDefinition(context.Background(), &resourcev1.CreateMaterialDefinitionRequest{
		MaterialClassId: "mc-1", PartNumber: "MAT-AL-001", UnitOfMeasure: "kg",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreateMaterialDefinition() = %v, want InvalidArgument", err)
	}
}

func TestCreateMaterialDefinition_MissingPartNumber(t *testing.T) {
	srv := newTestMaterialDefinitionServer()

	_, err := srv.CreateMaterialDefinition(context.Background(), &resourcev1.CreateMaterialDefinitionRequest{
		MaterialClassId: "mc-1", Name: "Aluminum Sheet", UnitOfMeasure: "kg",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreateMaterialDefinition() = %v, want InvalidArgument", err)
	}
}

func TestCreateMaterialDefinition_MissingUnitOfMeasure(t *testing.T) {
	srv := newTestMaterialDefinitionServer()

	_, err := srv.CreateMaterialDefinition(context.Background(), &resourcev1.CreateMaterialDefinitionRequest{
		MaterialClassId: "mc-1", Name: "Aluminum Sheet", PartNumber: "MAT-AL-001",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreateMaterialDefinition() = %v, want InvalidArgument", err)
	}
}

func TestCreateMaterialDefinition_MaterialClassNotFound(t *testing.T) {
	srv := NewMaterialDefinitionServer(newMockMaterialClassRepo(), newMockMaterialDefinitionRepo())

	_, err := srv.CreateMaterialDefinition(context.Background(), &resourcev1.CreateMaterialDefinitionRequest{
		MaterialClassId: "non-existent", Name: "Aluminum Sheet",
		PartNumber: "MAT-AL-001", UnitOfMeasure: "kg",
	})
	if status.Code(err) != codes.NotFound {
		t.Errorf("CreateMaterialDefinition() = %v, want NotFound", err)
	}
}

// ============================================================================
// GetMaterialDefinition
// ============================================================================

func TestGetMaterialDefinition_Success(t *testing.T) {
	classes := newMockMaterialClassRepo()
	defs := newMockMaterialDefinitionRepo()
	srv := NewMaterialDefinitionServer(classes, defs)

	class, _ := classes.CreateMaterialClass(context.Background(), "Raw Material", nil)

	createResp, _ := srv.CreateMaterialDefinition(context.Background(), &resourcev1.CreateMaterialDefinitionRequest{
		MaterialClassId: class.ID, Name: "Aluminum Sheet 3mm",
		PartNumber: "MAT-AL-3MM", UnitOfMeasure: "kg",
	})

	getResp, err := srv.GetMaterialDefinition(context.Background(), &resourcev1.GetMaterialDefinitionRequest{
		Id: createResp.GetMaterialDefinition().GetId(),
	})
	if err != nil {
		t.Fatalf("GetMaterialDefinition() = %v, want nil", err)
	}
	if getResp.GetMaterialDefinition().GetId() != createResp.GetMaterialDefinition().GetId() {
		t.Errorf("ID = %s, want %s", getResp.GetMaterialDefinition().GetId(), createResp.GetMaterialDefinition().GetId())
	}
}

func TestGetMaterialDefinition_MissingID(t *testing.T) {
	srv := newTestMaterialDefinitionServer()

	_, err := srv.GetMaterialDefinition(context.Background(), &resourcev1.GetMaterialDefinitionRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("GetMaterialDefinition() = %v, want InvalidArgument", err)
	}
}

func TestGetMaterialDefinition_NotFound(t *testing.T) {
	srv := newTestMaterialDefinitionServer()

	_, err := srv.GetMaterialDefinition(context.Background(), &resourcev1.GetMaterialDefinitionRequest{Id: "non-existent"})
	if status.Code(err) != codes.NotFound {
		t.Errorf("GetMaterialDefinition() = %v, want NotFound", err)
	}
}

// ============================================================================
// ListMaterialDefinitions
// ============================================================================

func TestListMaterialDefinitions_Success(t *testing.T) {
	classes := newMockMaterialClassRepo()
	defs := newMockMaterialDefinitionRepo()
	srv := NewMaterialDefinitionServer(classes, defs)

	class, _ := classes.CreateMaterialClass(context.Background(), "Raw Material", nil)

	_, _ = srv.CreateMaterialDefinition(context.Background(), &resourcev1.CreateMaterialDefinitionRequest{
		MaterialClassId: class.ID, Name: "Aluminum Sheet 3mm",
		PartNumber: "MAT-AL-3MM", UnitOfMeasure: "kg",
	})
	_, _ = srv.CreateMaterialDefinition(context.Background(), &resourcev1.CreateMaterialDefinitionRequest{
		MaterialClassId: class.ID, Name: "Steel Rod 10mm",
		PartNumber: "MAT-ST-10MM", UnitOfMeasure: "kg",
	})

	resp, err := srv.ListMaterialDefinitions(context.Background(), &resourcev1.ListMaterialDefinitionsRequest{})
	if err != nil {
		t.Fatalf("ListMaterialDefinitions() = %v, want nil", err)
	}
	if len(resp.GetMaterialDefinitions()) != 2 {
		t.Errorf("ListMaterialDefinitions() returned %d, want 2", len(resp.GetMaterialDefinitions()))
	}
}

func TestListMaterialDefinitions_FilterByClass(t *testing.T) {
	classes := newMockMaterialClassRepo()
	defs := newMockMaterialDefinitionRepo()
	srv := NewMaterialDefinitionServer(classes, defs)

	raw, _ := classes.CreateMaterialClass(context.Background(), "Raw Material", nil)
	finished, _ := classes.CreateMaterialClass(context.Background(), "Finished Good", nil)

	_, _ = srv.CreateMaterialDefinition(context.Background(), &resourcev1.CreateMaterialDefinitionRequest{
		MaterialClassId: raw.ID, Name: "Aluminum Sheet",
		PartNumber: "MAT-AL-001", UnitOfMeasure: "kg",
	})
	_, _ = srv.CreateMaterialDefinition(context.Background(), &resourcev1.CreateMaterialDefinitionRequest{
		MaterialClassId: finished.ID, Name: "Engine Block",
		PartNumber: "ENG-BLK-001", UnitOfMeasure: "pcs",
	})

	resp, err := srv.ListMaterialDefinitions(context.Background(), &resourcev1.ListMaterialDefinitionsRequest{
		MaterialClassId: raw.ID,
	})
	if err != nil {
		t.Fatalf("ListMaterialDefinitions(raw) = %v, want nil", err)
	}
	if len(resp.GetMaterialDefinitions()) != 1 {
		t.Errorf("ListMaterialDefinitions(raw) returned %d, want 1", len(resp.GetMaterialDefinitions()))
	}
}

func TestListMaterialDefinitions_Empty(t *testing.T) {
	srv := newTestMaterialDefinitionServer()

	resp, err := srv.ListMaterialDefinitions(context.Background(), &resourcev1.ListMaterialDefinitionsRequest{})
	if err != nil {
		t.Fatalf("ListMaterialDefinitions() = %v, want nil", err)
	}
	if len(resp.GetMaterialDefinitions()) != 0 {
		t.Errorf("ListMaterialDefinitions() returned %d, want 0", len(resp.GetMaterialDefinitions()))
	}
}

// ============================================================================
// Error-path tests
// ============================================================================

type mockMaterialDefinitionRepoErr struct{}

func (m *mockMaterialDefinitionRepoErr) CreateMaterialDefinition(_ context.Context, _, _, _, _ string, _ *string) (*db.MaterialDefinition, error) {
	return nil, fmt.Errorf("simulated db error")
}
func (m *mockMaterialDefinitionRepoErr) GetMaterialDefinition(_ context.Context, _ string) (*db.MaterialDefinition, error) {
	return nil, fmt.Errorf("simulated db error")
}
func (m *mockMaterialDefinitionRepoErr) ListMaterialDefinitions(_ context.Context, _ string) ([]*db.MaterialDefinition, error) {
	return nil, fmt.Errorf("simulated db error")
}

func TestCreateMaterialDefinition_RepoError(t *testing.T) {
	classes := newMockMaterialClassRepo()
	srv := NewMaterialDefinitionServer(classes, &mockMaterialDefinitionRepoErr{})

	class, _ := classes.CreateMaterialClass(context.Background(), "Raw Material", nil)

	_, err := srv.CreateMaterialDefinition(context.Background(), &resourcev1.CreateMaterialDefinitionRequest{
		MaterialClassId: class.ID, Name: "Aluminum Sheet",
		PartNumber: "MAT-AL-001", UnitOfMeasure: "kg",
	})
	if status.Code(err) != codes.Internal {
		t.Errorf("CreateMaterialDefinition() = %v, want Internal", err)
	}
}

func TestGetMaterialDefinition_RepoError(t *testing.T) {
	srv := NewMaterialDefinitionServer(newMockMaterialClassRepo(), &mockMaterialDefinitionRepoErr{})

	_, err := srv.GetMaterialDefinition(context.Background(), &resourcev1.GetMaterialDefinitionRequest{Id: "md-1"})
	if status.Code(err) != codes.Internal {
		t.Errorf("GetMaterialDefinition() = %v, want Internal", err)
	}
}

func TestListMaterialDefinitions_RepoError(t *testing.T) {
	srv := NewMaterialDefinitionServer(newMockMaterialClassRepo(), &mockMaterialDefinitionRepoErr{})

	_, err := srv.ListMaterialDefinitions(context.Background(), &resourcev1.ListMaterialDefinitionsRequest{})
	if status.Code(err) != codes.Internal {
		t.Errorf("ListMaterialDefinitions() = %v, want Internal", err)
	}
}
