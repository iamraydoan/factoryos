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

// mockMaterialClassRepo implements db.MaterialClassRepository for testing.
type mockMaterialClassRepo struct {
	classes map[string]*db.MaterialClass
	nextID  int
}

func newMockMaterialClassRepo() *mockMaterialClassRepo {
	return &mockMaterialClassRepo{classes: make(map[string]*db.MaterialClass), nextID: 1}
}

func (m *mockMaterialClassRepo) CreateMaterialClass(_ context.Context, name string, description *string) (*db.MaterialClass, error) {
	id := fmt.Sprintf("mc-%d", m.nextID)
	m.nextID++
	now := time.Now()
	mc := &db.MaterialClass{
		ID: id, Name: name, Description: description,
		CreatedAt: now, UpdatedAt: now,
	}
	m.classes[mc.ID] = mc
	return mc, nil
}

func (m *mockMaterialClassRepo) GetMaterialClass(_ context.Context, id string) (*db.MaterialClass, error) {
	mc, ok := m.classes[id]
	if !ok {
		return nil, nil
	}
	return mc, nil
}

func (m *mockMaterialClassRepo) ListMaterialClasses(_ context.Context) ([]*db.MaterialClass, error) {
	var result []*db.MaterialClass
	for _, mc := range m.classes {
		result = append(result, mc)
	}
	return result, nil
}

// ============================================================================
// CreateMaterialClass
// ============================================================================

func TestCreateMaterialClass_Success(t *testing.T) {
	srv := NewMaterialClassServer(newMockMaterialClassRepo())

	resp, err := srv.CreateMaterialClass(context.Background(), &resourcev1.CreateMaterialClassRequest{
		Name: "Raw Material", Description: "Base materials for production",
	})
	if err != nil {
		t.Fatalf("CreateMaterialClass() = %v, want nil", err)
	}
	if resp.GetMaterialClass().GetName() != "Raw Material" {
		t.Errorf("Name = %q, want 'Raw Material'", resp.GetMaterialClass().GetName())
	}
	if resp.GetMaterialClass().GetDescription() != "Base materials for production" {
		t.Errorf("Description = %q, want 'Base materials for production'", resp.GetMaterialClass().GetDescription())
	}
}

func TestCreateMaterialClass_MissingName(t *testing.T) {
	srv := NewMaterialClassServer(newMockMaterialClassRepo())

	_, err := srv.CreateMaterialClass(context.Background(), &resourcev1.CreateMaterialClassRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreateMaterialClass() = %v, want InvalidArgument", err)
	}
}

func TestCreateMaterialClass_NoDescription(t *testing.T) {
	srv := NewMaterialClassServer(newMockMaterialClassRepo())

	resp, err := srv.CreateMaterialClass(context.Background(), &resourcev1.CreateMaterialClassRequest{
		Name: "Finished Good",
	})
	if err != nil {
		t.Fatalf("CreateMaterialClass() = %v, want nil", err)
	}
	if resp.GetMaterialClass().GetName() != "Finished Good" {
		t.Errorf("Name = %q, want 'Finished Good'", resp.GetMaterialClass().GetName())
	}
}

// ============================================================================
// GetMaterialClass
// ============================================================================

func TestGetMaterialClass_Success(t *testing.T) {
	srv := NewMaterialClassServer(newMockMaterialClassRepo())

	createResp, _ := srv.CreateMaterialClass(context.Background(), &resourcev1.CreateMaterialClassRequest{
		Name: "Sub-Assembly",
	})

	getResp, err := srv.GetMaterialClass(context.Background(), &resourcev1.GetMaterialClassRequest{
		Id: createResp.GetMaterialClass().GetId(),
	})
	if err != nil {
		t.Fatalf("GetMaterialClass() = %v, want nil", err)
	}
	if getResp.GetMaterialClass().GetId() != createResp.GetMaterialClass().GetId() {
		t.Errorf("ID = %s, want %s", getResp.GetMaterialClass().GetId(), createResp.GetMaterialClass().GetId())
	}
}

func TestGetMaterialClass_MissingID(t *testing.T) {
	srv := NewMaterialClassServer(newMockMaterialClassRepo())

	_, err := srv.GetMaterialClass(context.Background(), &resourcev1.GetMaterialClassRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("GetMaterialClass() = %v, want InvalidArgument", err)
	}
}

func TestGetMaterialClass_NotFound(t *testing.T) {
	srv := NewMaterialClassServer(newMockMaterialClassRepo())

	_, err := srv.GetMaterialClass(context.Background(), &resourcev1.GetMaterialClassRequest{Id: "non-existent"})
	if status.Code(err) != codes.NotFound {
		t.Errorf("GetMaterialClass() = %v, want NotFound", err)
	}
}

// ============================================================================
// ListMaterialClasses
// ============================================================================

func TestListMaterialClasses_Success(t *testing.T) {
	srv := NewMaterialClassServer(newMockMaterialClassRepo())

	_, _ = srv.CreateMaterialClass(context.Background(), &resourcev1.CreateMaterialClassRequest{Name: "Raw Material"})
	_, _ = srv.CreateMaterialClass(context.Background(), &resourcev1.CreateMaterialClassRequest{Name: "Finished Good"})

	resp, err := srv.ListMaterialClasses(context.Background(), &resourcev1.ListMaterialClassesRequest{})
	if err != nil {
		t.Fatalf("ListMaterialClasses() = %v, want nil", err)
	}
	if len(resp.GetMaterialClasses()) != 2 {
		t.Errorf("ListMaterialClasses() returned %d, want 2", len(resp.GetMaterialClasses()))
	}
}

func TestListMaterialClasses_Empty(t *testing.T) {
	srv := NewMaterialClassServer(newMockMaterialClassRepo())

	resp, err := srv.ListMaterialClasses(context.Background(), &resourcev1.ListMaterialClassesRequest{})
	if err != nil {
		t.Fatalf("ListMaterialClasses() = %v, want nil", err)
	}
	if len(resp.GetMaterialClasses()) != 0 {
		t.Errorf("ListMaterialClasses() returned %d, want 0", len(resp.GetMaterialClasses()))
	}
}

// ============================================================================
// Error-path tests
// ============================================================================

type mockMaterialClassRepoErr struct{}

func (m *mockMaterialClassRepoErr) CreateMaterialClass(_ context.Context, _ string, _ *string) (*db.MaterialClass, error) {
	return nil, fmt.Errorf("simulated db error")
}
func (m *mockMaterialClassRepoErr) GetMaterialClass(_ context.Context, _ string) (*db.MaterialClass, error) {
	return nil, fmt.Errorf("simulated db error")
}
func (m *mockMaterialClassRepoErr) ListMaterialClasses(_ context.Context) ([]*db.MaterialClass, error) {
	return nil, fmt.Errorf("simulated db error")
}

func TestCreateMaterialClass_RepoError(t *testing.T) {
	srv := NewMaterialClassServer(&mockMaterialClassRepoErr{})
	_, err := srv.CreateMaterialClass(context.Background(), &resourcev1.CreateMaterialClassRequest{Name: "Raw Material"})
	if status.Code(err) != codes.Internal {
		t.Errorf("CreateMaterialClass() = %v, want Internal", err)
	}
}

func TestGetMaterialClass_RepoError(t *testing.T) {
	srv := NewMaterialClassServer(&mockMaterialClassRepoErr{})
	_, err := srv.GetMaterialClass(context.Background(), &resourcev1.GetMaterialClassRequest{Id: "mc-1"})
	if status.Code(err) != codes.Internal {
		t.Errorf("GetMaterialClass() = %v, want Internal", err)
	}
}

func TestListMaterialClasses_RepoError(t *testing.T) {
	srv := NewMaterialClassServer(&mockMaterialClassRepoErr{})
	_, err := srv.ListMaterialClasses(context.Background(), &resourcev1.ListMaterialClassesRequest{})
	if status.Code(err) != codes.Internal {
		t.Errorf("ListMaterialClasses() = %v, want Internal", err)
	}
}
