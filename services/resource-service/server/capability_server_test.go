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

// mockCapabilityRepo implements db.CapabilityRepository for testing.
type mockCapabilityRepo struct {
	caps   map[string]*db.WorkUnitCapability
	nextID int
}

func newMockCapabilityRepo() *mockCapabilityRepo {
	return &mockCapabilityRepo{caps: make(map[string]*db.WorkUnitCapability), nextID: 1}
}

func (m *mockCapabilityRepo) AssignCapability(_ context.Context, workUnitID, equipmentClassID string, properties map[string]interface{}) (*db.WorkUnitCapability, error) {
	key := workUnitID + ":" + equipmentClassID
	cap := &db.WorkUnitCapability{
		ID: fmt.Sprintf("cap-%d", m.nextID), WorkUnitID: workUnitID,
		EquipmentClassID: equipmentClassID, Properties: properties,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	m.nextID++
	m.caps[key] = cap
	return cap, nil
}

func (m *mockCapabilityRepo) ListWorkUnitCapabilities(_ context.Context, workUnitID string) ([]*db.WorkUnitCapability, error) {
	var result []*db.WorkUnitCapability
	for _, c := range m.caps {
		if c.WorkUnitID == workUnitID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *mockCapabilityRepo) RemoveCapability(_ context.Context, workUnitID, equipmentClassID string) (bool, error) {
	key := workUnitID + ":" + equipmentClassID
	if _, ok := m.caps[key]; !ok {
		return false, nil
	}
	delete(m.caps, key)
	return true, nil
}

func TestAssignCapability_Success(t *testing.T) {
	repo := newMockCapabilityRepo()
	srv := NewCapabilityServer(newMockWorkUnitRepo(), newMockEquipmentClassRepo(), repo)

	wu, _ := srv.workUnits.(*mockWorkUnitRepo).CreateWorkUnit(context.Background(), "wc-1", "Station A")
	ec, _ := srv.classes.(*mockEquipmentClassRepo).CreateEquipmentClass(context.Background(), "CNC Lathe", nil)

	resp, err := srv.AssignCapability(context.Background(), &resourcev1.AssignCapabilityRequest{
		WorkUnitId: wu.ID, EquipmentClassId: ec.ID, PropertiesJson: `{"max_speed_rpm": 5000}`,
	})
	if err != nil {
		t.Fatalf("AssignCapability() = %v, want nil", err)
	}
	if resp.GetCapability().GetWorkUnitId() != wu.ID {
		t.Errorf("WorkUnitId = %s, want %s", resp.GetCapability().GetWorkUnitId(), wu.ID)
	}
}

func TestAssignCapability_MissingWorkUnitID(t *testing.T) {
	srv := NewCapabilityServer(newMockWorkUnitRepo(), newMockEquipmentClassRepo(), newMockCapabilityRepo())

	_, err := srv.AssignCapability(context.Background(), &resourcev1.AssignCapabilityRequest{
		EquipmentClassId: "ec-1",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("AssignCapability() = %v, want InvalidArgument", err)
	}
}

func TestAssignCapability_MissingEquipmentClassID(t *testing.T) {
	srv := NewCapabilityServer(newMockWorkUnitRepo(), newMockEquipmentClassRepo(), newMockCapabilityRepo())

	_, err := srv.AssignCapability(context.Background(), &resourcev1.AssignCapabilityRequest{
		WorkUnitId: "wu-1",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("AssignCapability() = %v, want InvalidArgument", err)
	}
}

func TestAssignCapability_WorkUnitNotFound(t *testing.T) {
	srv := NewCapabilityServer(newMockWorkUnitRepo(), newMockEquipmentClassRepo(), newMockCapabilityRepo())

	ec, _ := srv.classes.(*mockEquipmentClassRepo).CreateEquipmentClass(context.Background(), "CNC Lathe", nil)

	_, err := srv.AssignCapability(context.Background(), &resourcev1.AssignCapabilityRequest{
		WorkUnitId: "non-existent", EquipmentClassId: ec.ID,
	})
	if status.Code(err) != codes.NotFound {
		t.Errorf("AssignCapability() = %v, want NotFound", err)
	}
}

func TestAssignCapability_EquipmentClassNotFound(t *testing.T) {
	srv := NewCapabilityServer(newMockWorkUnitRepo(), newMockEquipmentClassRepo(), newMockCapabilityRepo())

	wu, _ := srv.workUnits.(*mockWorkUnitRepo).CreateWorkUnit(context.Background(), "wc-1", "Station A")

	_, err := srv.AssignCapability(context.Background(), &resourcev1.AssignCapabilityRequest{
		WorkUnitId: wu.ID, EquipmentClassId: "non-existent",
	})
	if status.Code(err) != codes.NotFound {
		t.Errorf("AssignCapability() = %v, want NotFound", err)
	}
}

func TestAssignCapability_InvalidPropertiesJSON(t *testing.T) {
	srv := NewCapabilityServer(newMockWorkUnitRepo(), newMockEquipmentClassRepo(), newMockCapabilityRepo())

	wu, _ := srv.workUnits.(*mockWorkUnitRepo).CreateWorkUnit(context.Background(), "wc-1", "Station A")
	ec, _ := srv.classes.(*mockEquipmentClassRepo).CreateEquipmentClass(context.Background(), "CNC Lathe", nil)

	_, err := srv.AssignCapability(context.Background(), &resourcev1.AssignCapabilityRequest{
		WorkUnitId: wu.ID, EquipmentClassId: ec.ID, PropertiesJson: "not-json",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("AssignCapability() = %v, want InvalidArgument", err)
	}
}

func TestListWorkUnitCapabilities_Success(t *testing.T) {
	srv := NewCapabilityServer(newMockWorkUnitRepo(), newMockEquipmentClassRepo(), newMockCapabilityRepo())

	wu, _ := srv.workUnits.(*mockWorkUnitRepo).CreateWorkUnit(context.Background(), "wc-1", "Station A")
	ec1, _ := srv.classes.(*mockEquipmentClassRepo).CreateEquipmentClass(context.Background(), "CNC Lathe", nil)
	ec2, _ := srv.classes.(*mockEquipmentClassRepo).CreateEquipmentClass(context.Background(), "Milling", nil)

	_, _ = srv.AssignCapability(context.Background(), &resourcev1.AssignCapabilityRequest{
		WorkUnitId: wu.ID, EquipmentClassId: ec1.ID,
	})
	_, _ = srv.AssignCapability(context.Background(), &resourcev1.AssignCapabilityRequest{
		WorkUnitId: wu.ID, EquipmentClassId: ec2.ID,
	})

	resp, err := srv.ListWorkUnitCapabilities(context.Background(), &resourcev1.ListWorkUnitCapabilitiesRequest{WorkUnitId: wu.ID})
	if err != nil {
		t.Fatalf("ListWorkUnitCapabilities() = %v, want nil", err)
	}
	if len(resp.GetCapabilities()) != 2 {
		t.Errorf("ListWorkUnitCapabilities() returned %d, want 2", len(resp.GetCapabilities()))
	}
}

func TestListWorkUnitCapabilities_MissingWorkUnitID(t *testing.T) {
	srv := NewCapabilityServer(newMockWorkUnitRepo(), newMockEquipmentClassRepo(), newMockCapabilityRepo())

	_, err := srv.ListWorkUnitCapabilities(context.Background(), &resourcev1.ListWorkUnitCapabilitiesRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("ListWorkUnitCapabilities() = %v, want InvalidArgument", err)
	}
}

func TestListWorkUnitCapabilities_Empty(t *testing.T) {
	srv := NewCapabilityServer(newMockWorkUnitRepo(), newMockEquipmentClassRepo(), newMockCapabilityRepo())

	wu, _ := srv.workUnits.(*mockWorkUnitRepo).CreateWorkUnit(context.Background(), "wc-1", "Station A")

	resp, err := srv.ListWorkUnitCapabilities(context.Background(), &resourcev1.ListWorkUnitCapabilitiesRequest{WorkUnitId: wu.ID})
	if err != nil {
		t.Fatalf("ListWorkUnitCapabilities() = %v, want nil", err)
	}
	if len(resp.GetCapabilities()) != 0 {
		t.Errorf("ListWorkUnitCapabilities() returned %d, want 0", len(resp.GetCapabilities()))
	}
}

func TestRemoveCapability_Success(t *testing.T) {
	srv := NewCapabilityServer(newMockWorkUnitRepo(), newMockEquipmentClassRepo(), newMockCapabilityRepo())

	wu, _ := srv.workUnits.(*mockWorkUnitRepo).CreateWorkUnit(context.Background(), "wc-1", "Station A")
	ec, _ := srv.classes.(*mockEquipmentClassRepo).CreateEquipmentClass(context.Background(), "CNC Lathe", nil)
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
	srv := NewCapabilityServer(newMockWorkUnitRepo(), newMockEquipmentClassRepo(), newMockCapabilityRepo())

	_, err := srv.RemoveCapability(context.Background(), &resourcev1.RemoveCapabilityRequest{
		EquipmentClassId: "ec-1",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("RemoveCapability() = %v, want InvalidArgument", err)
	}
}

func TestRemoveCapability_MissingEquipmentClassID(t *testing.T) {
	srv := NewCapabilityServer(newMockWorkUnitRepo(), newMockEquipmentClassRepo(), newMockCapabilityRepo())

	_, err := srv.RemoveCapability(context.Background(), &resourcev1.RemoveCapabilityRequest{
		WorkUnitId: "wu-1",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("RemoveCapability() = %v, want InvalidArgument", err)
	}
}

func TestRemoveCapability_NotFound(t *testing.T) {
	srv := NewCapabilityServer(newMockWorkUnitRepo(), newMockEquipmentClassRepo(), newMockCapabilityRepo())

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
