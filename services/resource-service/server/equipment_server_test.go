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

// ============================================================================
// Mock Repository
// ============================================================================

// mockRepo implements db.EquipmentRepository for unit testing.
type mockRepo struct {
	units        map[string]*db.WorkUnit
	classes      map[string]*db.EquipmentClass
	capabilities map[string]*db.WorkUnitCapability
	nextID       int
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		units:        make(map[string]*db.WorkUnit),
		classes:      make(map[string]*db.EquipmentClass),
		capabilities: make(map[string]*db.WorkUnitCapability),
		nextID:       1,
	}
}

// --- Work Unit mocks ---

func (m *mockRepo) CreateWorkUnit(_ context.Context, workCenterID, name string) (*db.WorkUnit, error) {
	id := fmt.Sprintf("wu-%d", m.nextID)
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
		return nil, nil
	}
	unit.Status = newStatus
	unit.UpdatedAt = time.Now()
	return unit, nil
}

// --- Equipment Class mocks ---

func (m *mockRepo) CreateEquipmentClass(_ context.Context, name string, description *string) (*db.EquipmentClass, error) {
	id := fmt.Sprintf("ec-%d", m.nextID)
	m.nextID++
	ec := &db.EquipmentClass{
		ID:          id,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.classes[ec.ID] = ec
	return ec, nil
}

func (m *mockRepo) GetEquipmentClass(_ context.Context, id string) (*db.EquipmentClass, error) {
	ec, ok := m.classes[id]
	if !ok {
		return nil, nil
	}
	return ec, nil
}

func (m *mockRepo) ListEquipmentClasses(_ context.Context) ([]*db.EquipmentClass, error) {
	var result []*db.EquipmentClass
	for _, ec := range m.classes {
		result = append(result, ec)
	}
	return result, nil
}

// --- Capability mocks ---

func (m *mockRepo) AssignCapability(_ context.Context, workUnitID, equipmentClassID string, properties map[string]interface{}) (*db.WorkUnitCapability, error) {
	key := workUnitID + ":" + equipmentClassID
	cap := &db.WorkUnitCapability{
		ID:               fmt.Sprintf("cap-%d", m.nextID),
		WorkUnitID:       workUnitID,
		EquipmentClassID: equipmentClassID,
		Properties:       properties,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	m.nextID++
	m.capabilities[key] = cap
	return cap, nil
}

func (m *mockRepo) ListWorkUnitCapabilities(_ context.Context, workUnitID string) ([]*db.WorkUnitCapability, error) {
	var result []*db.WorkUnitCapability
	for _, cap := range m.capabilities {
		if cap.WorkUnitID == workUnitID {
			result = append(result, cap)
		}
	}
	return result, nil
}

func (m *mockRepo) RemoveCapability(_ context.Context, workUnitID, equipmentClassID string) (bool, error) {
	key := workUnitID + ":" + equipmentClassID
	if _, ok := m.capabilities[key]; !ok {
		return false, nil
	}
	delete(m.capabilities, key)
	return true, nil
}

// ============================================================================
// Helper
// ============================================================================

func strPtr(s string) *string { return &s }

// setupTest creates a mock repo and server, and optionally pre-creates a work unit and equipment class.
func setupTest() (*EquipmentServer, *mockRepo) {
	repo := newMockRepo()
	srv := NewEquipmentServer(repo)
	return srv, repo
}

// ============================================================================
// Work Unit Tests
// ============================================================================

func TestCreateWorkUnit_Success(t *testing.T) {
	srv, _ := setupTest()

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
	srv, _ := setupTest()

	req := &resourcev1.CreateWorkUnitRequest{Name: "Test Unit"}

	_, err := srv.CreateWorkUnit(context.Background(), req)
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreateWorkUnit() = %v, want InvalidArgument", err)
	}
}

func TestCreateWorkUnit_MissingName(t *testing.T) {
	srv, _ := setupTest()

	req := &resourcev1.CreateWorkUnitRequest{WorkCenterId: "wc-123"}

	_, err := srv.CreateWorkUnit(context.Background(), req)
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreateWorkUnit() = %v, want InvalidArgument", err)
	}
}

func TestCreateWorkUnit_NameTooLong(t *testing.T) {
	srv, _ := setupTest()

	longName := ""
	for i := 0; i < 256; i++ {
		longName += "a"
	}

	req := &resourcev1.CreateWorkUnitRequest{
		WorkCenterId: "wc-123",
		Name:         longName,
	}

	_, err := srv.CreateWorkUnit(context.Background(), req)
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreateWorkUnit() = %v, want InvalidArgument", err)
	}
}

func TestGetWorkUnit_Success(t *testing.T) {
	srv, _ := setupTest()

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
		t.Errorf("GetWorkUnit().ID = %s, want %s", getResp.GetWorkUnit().GetId(), createResp.GetWorkUnit().GetId())
	}
}

func TestGetWorkUnit_MissingID(t *testing.T) {
	srv, _ := setupTest()

	_, err := srv.GetWorkUnit(context.Background(), &resourcev1.GetWorkUnitRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("GetWorkUnit() = %v, want InvalidArgument", err)
	}
}

func TestGetWorkUnit_NotFound(t *testing.T) {
	srv, _ := setupTest()

	_, err := srv.GetWorkUnit(context.Background(), &resourcev1.GetWorkUnitRequest{Id: "non-existent"})
	if status.Code(err) != codes.NotFound {
		t.Errorf("GetWorkUnit() = %v, want NotFound", err)
	}
}

func TestListWorkUnits_Success(t *testing.T) {
	srv, _ := setupTest()

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
	srv, _ := setupTest()

	_, err := srv.ListWorkUnits(context.Background(), &resourcev1.ListWorkUnitsRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("ListWorkUnits() = %v, want InvalidArgument", err)
	}
}

func TestUpdateWorkUnitStatus_Success(t *testing.T) {
	srv, _ := setupTest()

	createResp, _ := srv.CreateWorkUnit(context.Background(), &resourcev1.CreateWorkUnitRequest{
		WorkCenterId: "wc-123", Name: "Test Unit",
	})

	// available → allocated
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

	// allocated → in_production
	updateResp2, err := srv.UpdateWorkUnitStatus(context.Background(), &resourcev1.UpdateWorkUnitStatusRequest{
		Id:     createResp.GetWorkUnit().GetId(),
		Status: resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_IN_PRODUCTION,
	})
	if err != nil {
		t.Fatalf("UpdateWorkUnitStatus() = %v, want nil", err)
	}
	if updateResp2.GetWorkUnit().GetStatus() != resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_IN_PRODUCTION {
		t.Errorf("Status = %v, want IN_PRODUCTION", updateResp2.GetWorkUnit().GetStatus())
	}
}

func TestUpdateWorkUnitStatus_MissingID(t *testing.T) {
	srv, _ := setupTest()

	_, err := srv.UpdateWorkUnitStatus(context.Background(), &resourcev1.UpdateWorkUnitStatusRequest{
		Status: resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_ALLOCATED,
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("UpdateWorkUnitStatus() = %v, want InvalidArgument", err)
	}
}

func TestUpdateWorkUnitStatus_NotFound(t *testing.T) {
	srv, _ := setupTest()

	_, err := srv.UpdateWorkUnitStatus(context.Background(), &resourcev1.UpdateWorkUnitStatusRequest{
		Id:     "non-existent",
		Status: resourcev1.WorkUnitStatus_WORK_UNIT_STATUS_ALLOCATED,
	})
	if status.Code(err) != codes.NotFound {
		t.Errorf("UpdateWorkUnitStatus() = %v, want NotFound", err)
	}
}

// ============================================================================
// Equipment Class Tests
// ============================================================================

func TestCreateEquipmentClass_Success(t *testing.T) {
	srv, _ := setupTest()

	resp, err := srv.CreateEquipmentClass(context.Background(), &resourcev1.CreateEquipmentClassRequest{
		Name:        "CNC Lathe ≥ 5-axis",
		Description: "Multi-axis CNC lathe",
	})
	if err != nil {
		t.Fatalf("CreateEquipmentClass() = %v, want nil", err)
	}
	if resp.GetEquipmentClass().GetName() != "CNC Lathe ≥ 5-axis" {
		t.Errorf("Name = %q, want 'CNC Lathe ≥ 5-axis'", resp.GetEquipmentClass().GetName())
	}
	if resp.GetEquipmentClass().GetDescription() != "Multi-axis CNC lathe" {
		t.Errorf("Description = %q, want 'Multi-axis CNC lathe'", resp.GetEquipmentClass().GetDescription())
	}
}

func TestCreateEquipmentClass_MissingName(t *testing.T) {
	srv, _ := setupTest()

	_, err := srv.CreateEquipmentClass(context.Background(), &resourcev1.CreateEquipmentClassRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreateEquipmentClass() = %v, want InvalidArgument", err)
	}
}

func TestGetEquipmentClass_Success(t *testing.T) {
	srv, _ := setupTest()

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
	srv, _ := setupTest()

	_, err := srv.GetEquipmentClass(context.Background(), &resourcev1.GetEquipmentClassRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("GetEquipmentClass() = %v, want InvalidArgument", err)
	}
}

func TestGetEquipmentClass_NotFound(t *testing.T) {
	srv, _ := setupTest()

	_, err := srv.GetEquipmentClass(context.Background(), &resourcev1.GetEquipmentClassRequest{Id: "non-existent"})
	if status.Code(err) != codes.NotFound {
		t.Errorf("GetEquipmentClass() = %v, want NotFound", err)
	}
}

func TestListEquipmentClasses_Success(t *testing.T) {
	srv, _ := setupTest()

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
	srv, _ := setupTest()

	resp, err := srv.ListEquipmentClasses(context.Background(), &resourcev1.ListEquipmentClassesRequest{})
	if err != nil {
		t.Fatalf("ListEquipmentClasses() = %v, want nil", err)
	}
	if len(resp.GetEquipmentClasses()) != 0 {
		t.Errorf("ListEquipmentClasses() returned %d, want 0", len(resp.GetEquipmentClasses()))
	}
}

// ============================================================================
// Capability Assignment Tests
// ============================================================================

func TestAssignCapability_Success(t *testing.T) {
	srv, repo := setupTest()

	// Create work unit and equipment class directly in mock
	wu, _ := repo.CreateWorkUnit(context.Background(), "wc-1", "Station A")
	ec, _ := repo.CreateEquipmentClass(context.Background(), "CNC Lathe", nil)

	resp, err := srv.AssignCapability(context.Background(), &resourcev1.AssignCapabilityRequest{
		WorkUnitId:       wu.ID,
		EquipmentClassId: ec.ID,
		PropertiesJson:   `{"max_speed_rpm": 5000}`,
	})
	if err != nil {
		t.Fatalf("AssignCapability() = %v, want nil", err)
	}
	if resp.GetCapability().GetWorkUnitId() != wu.ID {
		t.Errorf("WorkUnitId = %s, want %s", resp.GetCapability().GetWorkUnitId(), wu.ID)
	}
	if resp.GetCapability().GetEquipmentClassId() != ec.ID {
		t.Errorf("EquipmentClassId = %s, want %s", resp.GetCapability().GetEquipmentClassId(), ec.ID)
	}
}

func TestAssignCapability_MissingWorkUnitID(t *testing.T) {
	srv, _ := setupTest()

	_, err := srv.AssignCapability(context.Background(), &resourcev1.AssignCapabilityRequest{
		EquipmentClassId: "ec-1",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("AssignCapability() = %v, want InvalidArgument", err)
	}
}

func TestAssignCapability_MissingEquipmentClassID(t *testing.T) {
	srv, _ := setupTest()

	_, err := srv.AssignCapability(context.Background(), &resourcev1.AssignCapabilityRequest{
		WorkUnitId: "wu-1",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("AssignCapability() = %v, want InvalidArgument", err)
	}
}

func TestAssignCapability_WorkUnitNotFound(t *testing.T) {
	srv, repo := setupTest()

	ec, _ := repo.CreateEquipmentClass(context.Background(), "CNC Lathe", nil)

	_, err := srv.AssignCapability(context.Background(), &resourcev1.AssignCapabilityRequest{
		WorkUnitId:       "non-existent",
		EquipmentClassId: ec.ID,
	})
	if status.Code(err) != codes.NotFound {
		t.Errorf("AssignCapability() = %v, want NotFound", err)
	}
}

func TestAssignCapability_EquipmentClassNotFound(t *testing.T) {
	srv, repo := setupTest()

	wu, _ := repo.CreateWorkUnit(context.Background(), "wc-1", "Station A")

	_, err := srv.AssignCapability(context.Background(), &resourcev1.AssignCapabilityRequest{
		WorkUnitId:       wu.ID,
		EquipmentClassId: "non-existent",
	})
	if status.Code(err) != codes.NotFound {
		t.Errorf("AssignCapability() = %v, want NotFound", err)
	}
}

func TestAssignCapability_InvalidPropertiesJSON(t *testing.T) {
	srv, repo := setupTest()

	wu, _ := repo.CreateWorkUnit(context.Background(), "wc-1", "Station A")
	ec, _ := repo.CreateEquipmentClass(context.Background(), "CNC Lathe", nil)

	_, err := srv.AssignCapability(context.Background(), &resourcev1.AssignCapabilityRequest{
		WorkUnitId:       wu.ID,
		EquipmentClassId: ec.ID,
		PropertiesJson:   "not-json",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("AssignCapability() = %v, want InvalidArgument", err)
	}
}

func TestListWorkUnitCapabilities_Success(t *testing.T) {
	srv, repo := setupTest()

	wu, _ := repo.CreateWorkUnit(context.Background(), "wc-1", "Station A")
	ec1, _ := repo.CreateEquipmentClass(context.Background(), "CNC Lathe", nil)
	ec2, _ := repo.CreateEquipmentClass(context.Background(), "Milling", nil)

	_, _ = srv.AssignCapability(context.Background(), &resourcev1.AssignCapabilityRequest{
		WorkUnitId: wu.ID, EquipmentClassId: ec1.ID,
	})
	_, _ = srv.AssignCapability(context.Background(), &resourcev1.AssignCapabilityRequest{
		WorkUnitId: wu.ID, EquipmentClassId: ec2.ID,
	})

	resp, err := srv.ListWorkUnitCapabilities(context.Background(), &resourcev1.ListWorkUnitCapabilitiesRequest{
		WorkUnitId: wu.ID,
	})
	if err != nil {
		t.Fatalf("ListWorkUnitCapabilities() = %v, want nil", err)
	}
	if len(resp.GetCapabilities()) != 2 {
		t.Errorf("ListWorkUnitCapabilities() returned %d, want 2", len(resp.GetCapabilities()))
	}
}

func TestListWorkUnitCapabilities_MissingWorkUnitID(t *testing.T) {
	srv, _ := setupTest()

	_, err := srv.ListWorkUnitCapabilities(context.Background(), &resourcev1.ListWorkUnitCapabilitiesRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("ListWorkUnitCapabilities() = %v, want InvalidArgument", err)
	}
}

func TestListWorkUnitCapabilities_Empty(t *testing.T) {
	srv, repo := setupTest()

	wu, _ := repo.CreateWorkUnit(context.Background(), "wc-1", "Station A")

	resp, err := srv.ListWorkUnitCapabilities(context.Background(), &resourcev1.ListWorkUnitCapabilitiesRequest{
		WorkUnitId: wu.ID,
	})
	if err != nil {
		t.Fatalf("ListWorkUnitCapabilities() = %v, want nil", err)
	}
	if len(resp.GetCapabilities()) != 0 {
		t.Errorf("ListWorkUnitCapabilities() returned %d, want 0", len(resp.GetCapabilities()))
	}
}

func TestRemoveCapability_Success(t *testing.T) {
	srv, repo := setupTest()

	wu, _ := repo.CreateWorkUnit(context.Background(), "wc-1", "Station A")
	ec, _ := repo.CreateEquipmentClass(context.Background(), "CNC Lathe", nil)

	_, _ = srv.AssignCapability(context.Background(), &resourcev1.AssignCapabilityRequest{
		WorkUnitId: wu.ID, EquipmentClassId: ec.ID,
	})

	resp, err := srv.RemoveCapability(context.Background(), &resourcev1.RemoveCapabilityRequest{
		WorkUnitId: wu.ID, EquipmentClassId: ec.ID,
	})
	if err != nil {
		t.Fatalf("RemoveCapability() = %v, want nil", err)
	}
	if !resp.GetRemoved() {
		t.Error("RemoveCapability() = false, want true")
	}
}

func TestRemoveCapability_MissingWorkUnitID(t *testing.T) {
	srv, _ := setupTest()

	_, err := srv.RemoveCapability(context.Background(), &resourcev1.RemoveCapabilityRequest{
		EquipmentClassId: "ec-1",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("RemoveCapability() = %v, want InvalidArgument", err)
	}
}

func TestRemoveCapability_MissingEquipmentClassID(t *testing.T) {
	srv, _ := setupTest()

	_, err := srv.RemoveCapability(context.Background(), &resourcev1.RemoveCapabilityRequest{
		WorkUnitId: "wu-1",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("RemoveCapability() = %v, want InvalidArgument", err)
	}
}

func TestRemoveCapability_NotFound(t *testing.T) {
	srv, _ := setupTest()

	resp, err := srv.RemoveCapability(context.Background(), &resourcev1.RemoveCapabilityRequest{
		WorkUnitId: "wu-1", EquipmentClassId: "ec-1",
	})
	if err != nil {
		t.Fatalf("RemoveCapability() = %v, want nil", err)
	}
	if resp.GetRemoved() {
		t.Error("RemoveCapability() = true, want false")
	}
}

// ============================================================================
// Helper Function Tests
// ============================================================================

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

func TestToProtoEquipmentClass_Nil(t *testing.T) {
	got := toProtoEquipmentClass(nil)
	if got != nil {
		t.Errorf("toProtoEquipmentClass(nil) = %v, want nil", got)
	}
}

func TestToProtoCapability_Nil(t *testing.T) {
	got := toProtoCapability(nil)
	if got != nil {
		t.Errorf("toProtoCapability(nil) = %v, want nil", got)
	}
}

func TestToProtoWorkUnit_Nil(t *testing.T) {
	got := toProtoWorkUnit(nil)
	if got != nil {
		t.Errorf("toProtoWorkUnit(nil) = %v, want nil", got)
	}
}
