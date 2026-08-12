package server

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/iamraydoan/factoryos/services/resource-service/db"
	resourcev1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/resource/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mockRepo implements db.EquipmentRepository for unit testing.
type mockRepo struct {
	units  map[string]*db.WorkUnit
	nextID int
}

func newMockRepo() *mockRepo {
	return &mockRepo{units: make(map[string]*db.WorkUnit), nextID: 1}
}

func (m *mockRepo) CreateWorkUnit(_ context.Context, workCenterID, name string) (*db.WorkUnit, error) {
	id := fmt.Sprintf("test-uuid-%d", m.nextID)
	m.nextID++
	unit := &db.WorkUnit{
		ID:           id,
		WorkCenterID: workCenterID,
		Name:         name,
		Status:       "available",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	m.units[unit.ID] = unit
	return unit, nil
}

func (m *mockRepo) GetWorkUnit(_ context.Context, id string) (*db.WorkUnit, error) {
	unit, ok := m.units[id]
	if !ok {
		return nil, nil
	}
	return unit, nil
}

func (m *mockRepo) ListWorkUnits(_ context.Context, workCenterID string) ([]*db.WorkUnit, error) {
	var result []*db.WorkUnit
	for _, u := range m.units {
		if u.WorkCenterID == workCenterID {
			result = append(result, u)
		}
	}
	return result, nil
}

func (m *mockRepo) UpdateWorkUnitStatus(_ context.Context, id, expectedStatus, newStatus string) (*db.WorkUnit, error) {
	unit, ok := m.units[id]
	if !ok {
		return nil, nil
	}
	if unit.Status != expectedStatus {
		return nil, nil // Status changed concurrently
	}
	unit.Status = newStatus
	unit.UpdatedAt = time.Now()
	return unit, nil
}

// --- Tests ---

func TestCreateWorkUnit_Success(t *testing.T) {
	repo := newMockRepo()
	srv := NewEquipmentServer(repo)

	req := &resourcev1.CreateWorkUnitRequest{
		WorkCenterId: "wc-123",
		Name:         "Test Unit",
	}

	resp, err := srv.CreateWorkUnit(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateWorkUnit() = %v, want nil", err)
	}
	if resp.GetWorkUnit() == nil {
		t.Fatal("CreateWorkUnit() returned nil WorkUnit")
	}
	if resp.GetWorkUnit().GetName() != "Test Unit" {
		t.Errorf("WorkUnit.Name = %s, want 'Test Unit'", resp.GetWorkUnit().GetName())
	}
}

func TestCreateWorkUnit_MissingWorkCenterID(t *testing.T) {
	repo := newMockRepo()
	srv := NewEquipmentServer(repo)

	req := &resourcev1.CreateWorkUnitRequest{
		Name: "Test Unit",
	}

	_, err := srv.CreateWorkUnit(context.Background(), req)
	if err == nil {
		t.Fatal("CreateWorkUnit() = nil, want error")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreateWorkUnit() = %v, want InvalidArgument", err)
	}
}

func TestCreateWorkUnit_MissingName(t *testing.T) {
	repo := newMockRepo()
	srv := NewEquipmentServer(repo)

	req := &resourcev1.CreateWorkUnitRequest{
		WorkCenterId: "wc-123",
	}

	_, err := srv.CreateWorkUnit(context.Background(), req)
	if err == nil {
		t.Fatal("CreateWorkUnit() = nil, want error")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreateWorkUnit() = %v, want InvalidArgument", err)
	}
}

func TestCreateWorkUnit_NameTooLong(t *testing.T) {
	repo := newMockRepo()
	srv := NewEquipmentServer(repo)

	// Create a name longer than 255 characters
	longName := ""
	for i := 0; i < 256; i++ {
		longName += "a"
	}

	req := &resourcev1.CreateWorkUnitRequest{
		WorkCenterId: "wc-123",
		Name:         longName,
	}

	_, err := srv.CreateWorkUnit(context.Background(), req)
	if err == nil {
		t.Fatal("CreateWorkUnit() = nil, want error")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreateWorkUnit() = %v, want InvalidArgument", err)
	}
}

func TestGetWorkUnit_Success(t *testing.T) {
	repo := newMockRepo()
	srv := NewEquipmentServer(repo)

	// Create a unit first
	createReq := &resourcev1.CreateWorkUnitRequest{
		WorkCenterId: "wc-123",
		Name:         "Test Unit",
	}
	createResp, err := srv.CreateWorkUnit(context.Background(), createReq)
	if err != nil {
		t.Fatalf("CreateWorkUnit() = %v", err)
	}

	// Get the unit
	getReq := &resourcev1.GetWorkUnitRequest{
		Id: createResp.GetWorkUnit().GetId(),
	}

	getResp, err := srv.GetWorkUnit(context.Background(), getReq)
	if err != nil {
		t.Fatalf("GetWorkUnit() = %v, want nil", err)
	}
	if getResp.GetWorkUnit().GetId() != createResp.GetWorkUnit().GetId() {
		t.Errorf("GetWorkUnit().ID = %s, want %s", getResp.GetWorkUnit().GetId(), createResp.GetWorkUnit().GetId())
	}
}

func TestGetWorkUnit_MissingID(t *testing.T) {
	repo := newMockRepo()
	srv := NewEquipmentServer(repo)

	req := &resourcev1.GetWorkUnitRequest{}

	_, err := srv.GetWorkUnit(context.Background(), req)
	if err == nil {
		t.Fatal("GetWorkUnit() = nil, want error")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("GetWorkUnit() = %v, want InvalidArgument", err)
	}
}

func TestGetWorkUnit_NotFound(t *testing.T) {
	repo := newMockRepo()
	srv := NewEquipmentServer(repo)

	req := &resourcev1.GetWorkUnitRequest{
		Id: "non-existent-id",
	}

	_, err := srv.GetWorkUnit(context.Background(), req)
	if err == nil {
		t.Fatal("GetWorkUnit() = nil, want error")
	}
	if status.Code(err) != codes.NotFound {
		t.Errorf("GetWorkUnit() = %v, want NotFound", err)
	}
}

func TestListWorkUnits_Success(t *testing.T) {
	repo := newMockRepo()
	srv := NewEquipmentServer(repo)

	// Create 2 units in same work center
	for i := 1; i <= 2; i++ {
		req := &resourcev1.CreateWorkUnitRequest{
			WorkCenterId: "wc-123",
			Name:         "Unit " + string(rune('A'+i-1)),
		}
		_, err := srv.CreateWorkUnit(context.Background(), req)
		if err != nil {
			t.Fatalf("CreateWorkUnit(%d) = %v", i, err)
		}
	}

	// List
	listReq := &resourcev1.ListWorkUnitsRequest{
		WorkCenterId: "wc-123",
	}

	resp, err := srv.ListWorkUnits(context.Background(), listReq)
	if err != nil {
		t.Fatalf("ListWorkUnits() = %v, want nil", err)
	}
	if len(resp.GetWorkUnits()) != 2 {
		t.Errorf("ListWorkUnits() returned %d units, want 2", len(resp.GetWorkUnits()))
	}
}

func TestListWorkUnits_MissingWorkCenterID(t *testing.T) {
	repo := newMockRepo()
	srv := NewEquipmentServer(repo)

	req := &resourcev1.ListWorkUnitsRequest{}

	_, err := srv.ListWorkUnits(context.Background(), req)
	if err == nil {
		t.Fatal("ListWorkUnits() = nil, want error")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("ListWorkUnits() = %v, want InvalidArgument", err)
	}
}

func TestUpdateWorkUnitStatus_Success(t *testing.T) {
	repo := newMockRepo()
	srv := NewEquipmentServer(repo)

	// Create a unit
	createReq := &resourcev1.CreateWorkUnitRequest{
		WorkCenterId: "wc-123",
		Name:         "Test Unit",
	}
	createResp, err := srv.CreateWorkUnit(context.Background(), createReq)
	if err != nil {
		t.Fatalf("CreateWorkUnit() = %v", err)
	}

	// Update status: available → allocated
	updateReq := &resourcev1.UpdateWorkUnitStatusRequest{
		Id:     createResp.GetWorkUnit().GetId(),
		Status: resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_ALLOCATED,
	}

	updateResp, err := srv.UpdateWorkUnitStatus(context.Background(), updateReq)
	if err != nil {
		t.Fatalf("UpdateWorkUnitStatus() = %v, want nil", err)
	}
	if updateResp.GetWorkUnit().GetStatus() != resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_ALLOCATED {
		t.Errorf("Status = %v, want ALLOCATED", updateResp.GetWorkUnit().GetStatus())
	}

	// Update status: allocated → in_production
	updateReq2 := &resourcev1.UpdateWorkUnitStatusRequest{
		Id:     createResp.GetWorkUnit().GetId(),
		Status: resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_IN_PRODUCTION,
	}

	updateResp2, err := srv.UpdateWorkUnitStatus(context.Background(), updateReq2)
	if err != nil {
		t.Fatalf("UpdateWorkUnitStatus() = %v, want nil", err)
	}
	if updateResp2.GetWorkUnit().GetStatus() != resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_IN_PRODUCTION {
		t.Errorf("Status = %v, want IN_PRODUCTION", updateResp2.GetWorkUnit().GetStatus())
	}
}

func TestUpdateWorkUnitStatus_MissingID(t *testing.T) {
	repo := newMockRepo()
	srv := NewEquipmentServer(repo)

	req := &resourcev1.UpdateWorkUnitStatusRequest{
		Status: resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_ALLOCATED,
	}

	_, err := srv.UpdateWorkUnitStatus(context.Background(), req)
	if err == nil {
		t.Fatal("UpdateWorkUnitStatus() = nil, want error")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("UpdateWorkUnitStatus() = %v, want InvalidArgument", err)
	}
}

func TestUpdateWorkUnitStatus_NotFound(t *testing.T) {
	repo := newMockRepo()
	srv := NewEquipmentServer(repo)

	req := &resourcev1.UpdateWorkUnitStatusRequest{
		Id:     "non-existent-id",
		Status: resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_ALLOCATED,
	}

	_, err := srv.UpdateWorkUnitStatus(context.Background(), req)
	if err == nil {
		t.Fatal("UpdateWorkUnitStatus() = nil, want error")
	}
	if status.Code(err) != codes.NotFound {
		t.Errorf("UpdateWorkUnitStatus() = %v, want NotFound", err)
	}
}

func TestDerefString(t *testing.T) {
	tests := []struct {
		name string
		s    *string
		want string
	}{
		{"nil pointer", nil, ""},
		{"empty string", strPtr(""), ""},
		{"valid string", strPtr("test-asset"), "test-asset"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := derefString(tt.s)
			if got != tt.want {
				t.Errorf("derefString(%v) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}

func TestToProtoStatus(t *testing.T) {
	tests := []struct {
		input string
		want  resourcev1.WorkUnitStatus
	}{
		{"available", resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_AVAILABLE},
		{"allocated", resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_ALLOCATED},
		{"in_production", resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_IN_PRODUCTION},
		{"faulted", resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_FAULTED},
		{"unknown", resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toProtoStatus(tt.input)
			if got != tt.want {
				t.Errorf("toProtoStatus(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFromProtoStatus(t *testing.T) {
	tests := []struct {
		input resourcev1.WorkUnitStatus
		want  string
	}{
		{resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_AVAILABLE, "available"},
		{resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_ALLOCATED, "allocated"},
		{resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_IN_PRODUCTION, "in_production"},
		{resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_FAULTED, "faulted"},
		{resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_UNSPECIFIED, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := fromProtoStatus(tt.input)
			if got != tt.want {
				t.Errorf("fromProtoStatus(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// strPtr is a helper to create a *string from a string value.
func strPtr(s string) *string {
	return &s
}
