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

// mockWorkUnitRepo implements db.WorkUnitRepository for testing.
type mockWorkUnitRepo struct {
	units  map[string]*db.WorkUnit
	nextID int
}

func newMockWorkUnitRepo() *mockWorkUnitRepo {
	return &mockWorkUnitRepo{units: make(map[string]*db.WorkUnit), nextID: 1}
}

func (m *mockWorkUnitRepo) CreateWorkUnit(_ context.Context, workCenterID, name string) (*db.WorkUnit, error) {
	id := fmt.Sprintf("wu-%d", m.nextID)
	m.nextID++
	wu := &db.WorkUnit{
		ID: id, WorkCenterID: workCenterID, Name: name,
		Status: StatusAvailable, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	m.units[wu.ID] = wu
	return wu, nil
}

func (m *mockWorkUnitRepo) GetWorkUnit(_ context.Context, id string) (*db.WorkUnit, error) {
	wu, ok := m.units[id]
	if !ok {
		return nil, nil
	}
	return wu, nil
}

func (m *mockWorkUnitRepo) ListWorkUnits(_ context.Context, workCenterID string) ([]*db.WorkUnit, error) {
	var result []*db.WorkUnit
	for _, u := range m.units {
		if u.WorkCenterID == workCenterID {
			result = append(result, u)
		}
	}
	return result, nil
}

func (m *mockWorkUnitRepo) UpdateWorkUnitStatus(_ context.Context, id, expectedStatus, newStatus string) (*db.WorkUnit, error) {
	wu, ok := m.units[id]
	if !ok {
		return nil, nil
	}
	if wu.Status != expectedStatus {
		return nil, nil
	}
	wu.Status = newStatus
	wu.UpdatedAt = time.Now()
	return wu, nil
}

func TestCreateWorkUnit_Success(t *testing.T) {
	srv := NewWorkUnitServer(newMockWorkUnitRepo())

	resp, err := srv.CreateWorkUnit(context.Background(), &resourcev1.CreateWorkUnitRequest{
		WorkCenterId: "wc-123", Name: "Test Unit",
	})
	if err != nil {
		t.Fatalf("CreateWorkUnit() = %v, want nil", err)
	}
	if resp.GetWorkUnit().GetName() != "Test Unit" {
		t.Errorf("Name = %s, want 'Test Unit'", resp.GetWorkUnit().GetName())
	}
}

func TestCreateWorkUnit_MissingWorkCenterID(t *testing.T) {
	srv := NewWorkUnitServer(newMockWorkUnitRepo())

	_, err := srv.CreateWorkUnit(context.Background(), &resourcev1.CreateWorkUnitRequest{Name: "Test Unit"})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreateWorkUnit() = %v, want InvalidArgument", err)
	}
}

func TestCreateWorkUnit_MissingName(t *testing.T) {
	srv := NewWorkUnitServer(newMockWorkUnitRepo())

	_, err := srv.CreateWorkUnit(context.Background(), &resourcev1.CreateWorkUnitRequest{WorkCenterId: "wc-123"})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreateWorkUnit() = %v, want InvalidArgument", err)
	}
}

func TestCreateWorkUnit_NameTooLong(t *testing.T) {
	srv := NewWorkUnitServer(newMockWorkUnitRepo())

	longName := ""
	for i := 0; i < 256; i++ {
		longName += "a"
	}

	_, err := srv.CreateWorkUnit(context.Background(), &resourcev1.CreateWorkUnitRequest{
		WorkCenterId: "wc-123", Name: longName,
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreateWorkUnit() = %v, want InvalidArgument", err)
	}
}

func TestGetWorkUnit_Success(t *testing.T) {
	srv := NewWorkUnitServer(newMockWorkUnitRepo())

	createResp, _ := srv.CreateWorkUnit(context.Background(), &resourcev1.CreateWorkUnitRequest{
		WorkCenterId: "wc-123", Name: "Test Unit",
	})

	getResp, err := srv.GetWorkUnit(context.Background(), &resourcev1.GetWorkUnitRequest{
		Id: createResp.GetWorkUnit().GetId(),
	})
	if err != nil {
		t.Fatalf("GetWorkUnit() = %v, want nil", err)
	}
	if getResp.GetWorkUnit().GetId() != createResp.GetWorkUnit().GetId() {
		t.Errorf("ID = %s, want %s", getResp.GetWorkUnit().GetId(), createResp.GetWorkUnit().GetId())
	}
}

func TestGetWorkUnit_MissingID(t *testing.T) {
	srv := NewWorkUnitServer(newMockWorkUnitRepo())

	_, err := srv.GetWorkUnit(context.Background(), &resourcev1.GetWorkUnitRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("GetWorkUnit() = %v, want InvalidArgument", err)
	}
}

func TestGetWorkUnit_NotFound(t *testing.T) {
	srv := NewWorkUnitServer(newMockWorkUnitRepo())

	_, err := srv.GetWorkUnit(context.Background(), &resourcev1.GetWorkUnitRequest{Id: "non-existent"})
	if status.Code(err) != codes.NotFound {
		t.Errorf("GetWorkUnit() = %v, want NotFound", err)
	}
}

func TestListWorkUnits_Success(t *testing.T) {
	srv := NewWorkUnitServer(newMockWorkUnitRepo())

	for i := 1; i <= 2; i++ {
		_, _ = srv.CreateWorkUnit(context.Background(), &resourcev1.CreateWorkUnitRequest{
			WorkCenterId: "wc-123", Name: fmt.Sprintf("Unit %d", i),
		})
	}

	resp, err := srv.ListWorkUnits(context.Background(), &resourcev1.ListWorkUnitsRequest{WorkCenterId: "wc-123"})
	if err != nil {
		t.Fatalf("ListWorkUnits() = %v, want nil", err)
	}
	if len(resp.GetWorkUnits()) != 2 {
		t.Errorf("ListWorkUnits() returned %d units, want 2", len(resp.GetWorkUnits()))
	}
}

func TestListWorkUnits_MissingWorkCenterID(t *testing.T) {
	srv := NewWorkUnitServer(newMockWorkUnitRepo())

	_, err := srv.ListWorkUnits(context.Background(), &resourcev1.ListWorkUnitsRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("ListWorkUnits() = %v, want InvalidArgument", err)
	}
}

func TestUpdateWorkUnitStatus_Success(t *testing.T) {
	srv := NewWorkUnitServer(newMockWorkUnitRepo())

	createResp, _ := srv.CreateWorkUnit(context.Background(), &resourcev1.CreateWorkUnitRequest{
		WorkCenterId: "wc-123", Name: "Test Unit",
	})

	updateResp, err := srv.UpdateWorkUnitStatus(context.Background(), &resourcev1.UpdateWorkUnitStatusRequest{
		Id:     createResp.GetWorkUnit().GetId(),
		Status: resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_ALLOCATED,
	})
	if err != nil {
		t.Fatalf("UpdateWorkUnitStatus() = %v, want nil", err)
	}
	if updateResp.GetWorkUnit().GetStatus() != resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_ALLOCATED {
		t.Errorf("Status = %v, want ALLOCATED", updateResp.GetWorkUnit().GetStatus())
	}
}

func TestUpdateWorkUnitStatus_MissingID(t *testing.T) {
	srv := NewWorkUnitServer(newMockWorkUnitRepo())

	_, err := srv.UpdateWorkUnitStatus(context.Background(), &resourcev1.UpdateWorkUnitStatusRequest{
		Status: resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_ALLOCATED,
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("UpdateWorkUnitStatus() = %v, want InvalidArgument", err)
	}
}

func TestUpdateWorkUnitStatus_NotFound(t *testing.T) {
	srv := NewWorkUnitServer(newMockWorkUnitRepo())

	_, err := srv.UpdateWorkUnitStatus(context.Background(), &resourcev1.UpdateWorkUnitStatusRequest{
		Id: "non-existent", Status: resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_ALLOCATED,
	})
	if status.Code(err) != codes.NotFound {
		t.Errorf("UpdateWorkUnitStatus() = %v, want NotFound", err)
	}
}
