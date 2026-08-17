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

// mockEquipmentClassRepo implements db.EquipmentClassRepository for testing.
type mockEquipmentClassRepo struct {
	classes map[string]*db.EquipmentClass
	nextID  int
}

func newMockEquipmentClassRepo() *mockEquipmentClassRepo {
	return &mockEquipmentClassRepo{classes: make(map[string]*db.EquipmentClass), nextID: 1}
}

func (m *mockEquipmentClassRepo) CreateEquipmentClass(_ context.Context, name string, description *string) (*db.EquipmentClass, error) {
	id := fmt.Sprintf("ec-%d", m.nextID)
	m.nextID++
	ec := &db.EquipmentClass{
		ID: id, Name: name, Description: description,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	m.classes[ec.ID] = ec
	return ec, nil
}

func (m *mockEquipmentClassRepo) GetEquipmentClass(_ context.Context, id string) (*db.EquipmentClass, error) {
	ec, ok := m.classes[id]
	if !ok {
		return nil, nil
	}
	return ec, nil
}

func (m *mockEquipmentClassRepo) ListEquipmentClasses(_ context.Context) ([]*db.EquipmentClass, error) {
	var result []*db.EquipmentClass
	for _, ec := range m.classes {
		result = append(result, ec)
	}
	return result, nil
}

func TestCreateEquipmentClass_Success(t *testing.T) {
	srv := NewEquipmentClassServer(newMockEquipmentClassRepo())

	resp, err := srv.CreateEquipmentClass(context.Background(), &resourcev1.CreateEquipmentClassRequest{
		Name: "CNC Lathe ≥ 5-axis", Description: "Multi-axis CNC lathe",
	})
	if err != nil {
		t.Fatalf("CreateEquipmentClass() = %v, want nil", err)
	}
	if resp.GetEquipmentClass().GetName() != "CNC Lathe ≥ 5-axis" {
		t.Errorf("Name = %q, want 'CNC Lathe ≥ 5-axis'", resp.GetEquipmentClass().GetName())
	}
}

func TestCreateEquipmentClass_MissingName(t *testing.T) {
	srv := NewEquipmentClassServer(newMockEquipmentClassRepo())

	_, err := srv.CreateEquipmentClass(context.Background(), &resourcev1.CreateEquipmentClassRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreateEquipmentClass() = %v, want InvalidArgument", err)
	}
}

func TestGetEquipmentClass_Success(t *testing.T) {
	srv := NewEquipmentClassServer(newMockEquipmentClassRepo())

	createResp, _ := srv.CreateEquipmentClass(context.Background(), &resourcev1.CreateEquipmentClassRequest{
		Name: "CNC Lathe",
	})

	getResp, err := srv.GetEquipmentClass(context.Background(), &resourcev1.GetEquipmentClassRequest{
		Id: createResp.GetEquipmentClass().GetId(),
	})
	if err != nil {
		t.Fatalf("GetEquipmentClass() = %v, want nil", err)
	}
	if getResp.GetEquipmentClass().GetId() != createResp.GetEquipmentClass().GetId() {
		t.Errorf("ID = %s, want %s", getResp.GetEquipmentClass().GetId(), createResp.GetEquipmentClass().GetId())
	}
}

func TestGetEquipmentClass_MissingID(t *testing.T) {
	srv := NewEquipmentClassServer(newMockEquipmentClassRepo())

	_, err := srv.GetEquipmentClass(context.Background(), &resourcev1.GetEquipmentClassRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("GetEquipmentClass() = %v, want InvalidArgument", err)
	}
}

func TestGetEquipmentClass_NotFound(t *testing.T) {
	srv := NewEquipmentClassServer(newMockEquipmentClassRepo())

	_, err := srv.GetEquipmentClass(context.Background(), &resourcev1.GetEquipmentClassRequest{Id: "non-existent"})
	if status.Code(err) != codes.NotFound {
		t.Errorf("GetEquipmentClass() = %v, want NotFound", err)
	}
}

func TestListEquipmentClasses_Success(t *testing.T) {
	srv := NewEquipmentClassServer(newMockEquipmentClassRepo())

	_, _ = srv.CreateEquipmentClass(context.Background(), &resourcev1.CreateEquipmentClassRequest{Name: "CNC Lathe"})
	_, _ = srv.CreateEquipmentClass(context.Background(), &resourcev1.CreateEquipmentClassRequest{Name: "Milling Machine"})

	resp, err := srv.ListEquipmentClasses(context.Background(), &resourcev1.ListEquipmentClassesRequest{})
	if err != nil {
		t.Fatalf("ListEquipmentClasses() = %v, want nil", err)
	}
	if len(resp.GetEquipmentClasses()) != 2 {
		t.Errorf("ListEquipmentClasses() returned %d, want 2", len(resp.GetEquipmentClasses()))
	}
}

func TestListEquipmentClasses_Empty(t *testing.T) {
	srv := NewEquipmentClassServer(newMockEquipmentClassRepo())

	resp, err := srv.ListEquipmentClasses(context.Background(), &resourcev1.ListEquipmentClassesRequest{})
	if err != nil {
		t.Fatalf("ListEquipmentClasses() = %v, want nil", err)
	}
	if len(resp.GetEquipmentClasses()) != 0 {
		t.Errorf("ListEquipmentClasses() returned %d, want 0", len(resp.GetEquipmentClasses()))
	}
}

// ============================================================================
// Error-path tests: mock repos that always return errors
// ============================================================================

// mockEquipmentClassRepoErr implements db.EquipmentClassRepository and returns errors on all methods.
type mockEquipmentClassRepoErr struct{}

func (m *mockEquipmentClassRepoErr) CreateEquipmentClass(_ context.Context, _ string, _ *string) (*db.EquipmentClass, error) {
	return nil, fmt.Errorf("simulated db error")
}

func (m *mockEquipmentClassRepoErr) GetEquipmentClass(_ context.Context, _ string) (*db.EquipmentClass, error) {
	return nil, fmt.Errorf("simulated db error")
}

func (m *mockEquipmentClassRepoErr) ListEquipmentClasses(_ context.Context) ([]*db.EquipmentClass, error) {
	return nil, fmt.Errorf("simulated db error")
}

func TestCreateEquipmentClass_RepoError(t *testing.T) {
	srv := NewEquipmentClassServer(&mockEquipmentClassRepoErr{})

	_, err := srv.CreateEquipmentClass(context.Background(), &resourcev1.CreateEquipmentClassRequest{
		Name: "CNC Lathe",
	})
	if status.Code(err) != codes.Internal {
		t.Errorf("CreateEquipmentClass() = %v, want Internal", err)
	}
}

func TestGetEquipmentClass_RepoError(t *testing.T) {
	srv := NewEquipmentClassServer(&mockEquipmentClassRepoErr{})

	_, err := srv.GetEquipmentClass(context.Background(), &resourcev1.GetEquipmentClassRequest{Id: "ec-1"})
	if status.Code(err) != codes.Internal {
		t.Errorf("GetEquipmentClass() = %v, want Internal", err)
	}
}

func TestListEquipmentClasses_RepoError(t *testing.T) {
	srv := NewEquipmentClassServer(&mockEquipmentClassRepoErr{})

	_, err := srv.ListEquipmentClasses(context.Background(), &resourcev1.ListEquipmentClassesRequest{})
	if status.Code(err) != codes.Internal {
		t.Errorf("ListEquipmentClasses() = %v, want Internal", err)
	}
}
