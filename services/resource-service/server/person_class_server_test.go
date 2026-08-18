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

// mockPersonClassRepo implements db.PersonClassRepository for testing.
type mockPersonClassRepo struct {
	classes map[string]*db.PersonClass
	nextID  int
}

func newMockPersonClassRepo() *mockPersonClassRepo {
	return &mockPersonClassRepo{classes: make(map[string]*db.PersonClass), nextID: 1}
}

func (m *mockPersonClassRepo) CreatePersonClass(_ context.Context, name string, description *string) (*db.PersonClass, error) {
	id := fmt.Sprintf("pc-%d", m.nextID)
	m.nextID++
	pc := &db.PersonClass{
		ID: id, Name: name, Description: description,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	m.classes[pc.ID] = pc
	return pc, nil
}

func (m *mockPersonClassRepo) GetPersonClass(_ context.Context, id string) (*db.PersonClass, error) {
	pc, ok := m.classes[id]
	if !ok {
		return nil, nil
	}
	return pc, nil
}

func (m *mockPersonClassRepo) ListPersonClasses(_ context.Context) ([]*db.PersonClass, error) {
	var result []*db.PersonClass
	for _, pc := range m.classes {
		result = append(result, pc)
	}
	return result, nil
}

// ============================================================================
// Create
// ============================================================================

func TestCreatePersonClass_Success(t *testing.T) {
	srv := NewPersonClassServer(newMockPersonClassRepo())

	resp, err := srv.CreatePersonClass(context.Background(), &resourcev1.CreatePersonClassRequest{
		Name: "Operator", Description: "Shop floor operator",
	})
	if err != nil {
		t.Fatalf("CreatePersonClass() = %v, want nil", err)
	}
	if resp.GetPersonClass().GetName() != "Operator" {
		t.Errorf("Name = %q, want 'Operator'", resp.GetPersonClass().GetName())
	}
	if resp.GetPersonClass().GetDescription() != "Shop floor operator" {
		t.Errorf("Description = %q, want 'Shop floor operator'", resp.GetPersonClass().GetDescription())
	}
}

func TestCreatePersonClass_NoDescription(t *testing.T) {
	srv := NewPersonClassServer(newMockPersonClassRepo())

	resp, err := srv.CreatePersonClass(context.Background(), &resourcev1.CreatePersonClassRequest{
		Name: "Technician",
	})
	if err != nil {
		t.Fatalf("CreatePersonClass() = %v, want nil", err)
	}
	if resp.GetPersonClass().GetName() != "Technician" {
		t.Errorf("Name = %q, want 'Technician'", resp.GetPersonClass().GetName())
	}
}

func TestCreatePersonClass_MissingName(t *testing.T) {
	srv := NewPersonClassServer(newMockPersonClassRepo())

	_, err := srv.CreatePersonClass(context.Background(), &resourcev1.CreatePersonClassRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreatePersonClass() = %v, want InvalidArgument", err)
	}
}

// ============================================================================
// Get
// ============================================================================

func TestGetPersonClass_Success(t *testing.T) {
	srv := NewPersonClassServer(newMockPersonClassRepo())

	createResp, _ := srv.CreatePersonClass(context.Background(), &resourcev1.CreatePersonClassRequest{
		Name: "Technician",
	})

	getResp, err := srv.GetPersonClass(context.Background(), &resourcev1.GetPersonClassRequest{
		Id: createResp.GetPersonClass().GetId(),
	})
	if err != nil {
		t.Fatalf("GetPersonClass() = %v, want nil", err)
	}
	if getResp.GetPersonClass().GetId() != createResp.GetPersonClass().GetId() {
		t.Errorf("ID = %s, want %s", getResp.GetPersonClass().GetId(), createResp.GetPersonClass().GetId())
	}
}

func TestGetPersonClass_MissingID(t *testing.T) {
	srv := NewPersonClassServer(newMockPersonClassRepo())

	_, err := srv.GetPersonClass(context.Background(), &resourcev1.GetPersonClassRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("GetPersonClass() = %v, want InvalidArgument", err)
	}
}

func TestGetPersonClass_NotFound(t *testing.T) {
	srv := NewPersonClassServer(newMockPersonClassRepo())

	_, err := srv.GetPersonClass(context.Background(), &resourcev1.GetPersonClassRequest{Id: "non-existent"})
	if status.Code(err) != codes.NotFound {
		t.Errorf("GetPersonClass() = %v, want NotFound", err)
	}
}

// ============================================================================
// List
// ============================================================================

func TestListPersonClasses_Success(t *testing.T) {
	srv := NewPersonClassServer(newMockPersonClassRepo())

	_, _ = srv.CreatePersonClass(context.Background(), &resourcev1.CreatePersonClassRequest{Name: "Operator"})
	_, _ = srv.CreatePersonClass(context.Background(), &resourcev1.CreatePersonClassRequest{Name: "Technician"})

	resp, err := srv.ListPersonClasses(context.Background(), &resourcev1.ListPersonClassesRequest{})
	if err != nil {
		t.Fatalf("ListPersonClasses() = %v, want nil", err)
	}
	if len(resp.GetPersonClasses()) != 2 {
		t.Errorf("ListPersonClasses() returned %d, want 2", len(resp.GetPersonClasses()))
	}
}

func TestListPersonClasses_Empty(t *testing.T) {
	srv := NewPersonClassServer(newMockPersonClassRepo())

	resp, err := srv.ListPersonClasses(context.Background(), &resourcev1.ListPersonClassesRequest{})
	if err != nil {
		t.Fatalf("ListPersonClasses() = %v, want nil", err)
	}
	if len(resp.GetPersonClasses()) != 0 {
		t.Errorf("ListPersonClasses() returned %d, want 0", len(resp.GetPersonClasses()))
	}
}

// ============================================================================
// Error-path tests
// ============================================================================

// mockPersonClassRepoErr returns errors on all methods.
type mockPersonClassRepoErr struct{}

func (m *mockPersonClassRepoErr) CreatePersonClass(_ context.Context, _ string, _ *string) (*db.PersonClass, error) {
	return nil, fmt.Errorf("simulated db error")
}

func (m *mockPersonClassRepoErr) GetPersonClass(_ context.Context, _ string) (*db.PersonClass, error) {
	return nil, fmt.Errorf("simulated db error")
}

func (m *mockPersonClassRepoErr) ListPersonClasses(_ context.Context) ([]*db.PersonClass, error) {
	return nil, fmt.Errorf("simulated db error")
}

func TestCreatePersonClass_RepoError(t *testing.T) {
	srv := NewPersonClassServer(&mockPersonClassRepoErr{})

	_, err := srv.CreatePersonClass(context.Background(), &resourcev1.CreatePersonClassRequest{
		Name: "Operator",
	})
	if status.Code(err) != codes.Internal {
		t.Errorf("CreatePersonClass() = %v, want Internal", err)
	}
}

func TestGetPersonClass_RepoError(t *testing.T) {
	srv := NewPersonClassServer(&mockPersonClassRepoErr{})

	_, err := srv.GetPersonClass(context.Background(), &resourcev1.GetPersonClassRequest{Id: "pc-1"})
	if status.Code(err) != codes.Internal {
		t.Errorf("GetPersonClass() = %v, want Internal", err)
	}
}

func TestListPersonClasses_RepoError(t *testing.T) {
	srv := NewPersonClassServer(&mockPersonClassRepoErr{})

	_, err := srv.ListPersonClasses(context.Background(), &resourcev1.ListPersonClassesRequest{})
	if status.Code(err) != codes.Internal {
		t.Errorf("ListPersonClasses() = %v, want Internal", err)
	}
}
