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

// mockShiftAssignmentRepo implements db.ShiftAssignmentRepository for testing.
type mockShiftAssignmentRepo struct {
	assignments map[string]*db.ShiftAssignment
	nextID      int
}

func newMockShiftAssignmentRepo() *mockShiftAssignmentRepo {
	return &mockShiftAssignmentRepo{assignments: make(map[string]*db.ShiftAssignment), nextID: 1}
}

func (m *mockShiftAssignmentRepo) AssignShift(_ context.Context, personID, shiftID, workCenterID string, effectiveFrom time.Time) (*db.ShiftAssignment, error) {
	id := fmt.Sprintf("sa-%d", m.nextID)
	m.nextID++
	now := time.Now()
	sa := &db.ShiftAssignment{
		ID: id, PersonID: personID, ShiftID: shiftID, WorkCenterID: workCenterID,
		EffectiveFrom: effectiveFrom, CreatedAt: now, UpdatedAt: now,
	}
	m.assignments[sa.ID] = sa
	return sa, nil
}

func (m *mockShiftAssignmentRepo) GetShiftAssignment(_ context.Context, id string) (*db.ShiftAssignment, error) {
	sa, ok := m.assignments[id]
	if !ok {
		return nil, nil
	}
	return sa, nil
}

func (m *mockShiftAssignmentRepo) ListShiftAssignments(_ context.Context, personID, shiftID, workCenterID string) ([]*db.ShiftAssignment, error) {
	var result []*db.ShiftAssignment
	for _, sa := range m.assignments {
		matchPerson := personID == "" || sa.PersonID == personID
		matchShift := shiftID == "" || sa.ShiftID == shiftID
		matchWC := workCenterID == "" || sa.WorkCenterID == workCenterID
		if matchPerson && matchShift && matchWC {
			result = append(result, sa)
		}
	}
	return result, nil
}

func (m *mockShiftAssignmentRepo) UnassignShift(_ context.Context, id string) (bool, error) {
	if _, ok := m.assignments[id]; !ok {
		return false, nil
	}
	delete(m.assignments, id)
	return true, nil
}

// newTestShiftAssignmentServer creates a ShiftAssignmentServer with all happy-path mocks.
func newTestShiftAssignmentServer() *ShiftAssignmentServer {
	return NewShiftAssignmentServer(
		newMockPersonRepo(),
		newMockShiftRepo(),
		newMockShiftAssignmentRepo(),
	)
}

// seedShiftAssignment creates an assignment directly in the mock repo.
func seedShiftAssignment(repo *mockShiftAssignmentRepo, personID, shiftID, workCenterID string, effectiveFrom time.Time) {
	id := fmt.Sprintf("sa-%d", repo.nextID)
	repo.nextID++
	now := time.Now()
	sa := &db.ShiftAssignment{
		ID: id, PersonID: personID, ShiftID: shiftID, WorkCenterID: workCenterID,
		EffectiveFrom: effectiveFrom, CreatedAt: now, UpdatedAt: now,
	}
	repo.assignments[sa.ID] = sa
}

// ============================================================================
// AssignShift
// ============================================================================

func TestAssignShift_Success(t *testing.T) {
	persons := newMockPersonRepo()
	shifts := newMockShiftRepo()
	srv := NewShiftAssignmentServer(persons, shifts, newMockShiftAssignmentRepo())

	person, _ := persons.CreatePerson(context.Background(), "pc-1", "EMP-001", "Jane", "Doe", nil)
	shift, _ := shifts.CreateShift(context.Background(), "Day Shift", "06:00", "14:00", nil)

	resp, err := srv.AssignShift(context.Background(), &resourcev1.AssignShiftRequest{
		PersonId:      person.ID,
		ShiftId:       shift.ID,
		WorkCenterId:  "wc-1",
		EffectiveFrom: "2026-08-21T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("AssignShift() = %v, want nil", err)
	}
	if resp.GetAssignment().GetPersonId() != person.ID {
		t.Errorf("PersonId = %q, want %q", resp.GetAssignment().GetPersonId(), person.ID)
	}
	if resp.GetAssignment().GetShiftId() != shift.ID {
		t.Errorf("ShiftId = %q, want %q", resp.GetAssignment().GetShiftId(), shift.ID)
	}
	if resp.GetAssignment().GetEffectiveTo() != nil {
		t.Errorf("EffectiveTo = %v, want nil (open-ended)", resp.GetAssignment().GetEffectiveTo())
	}
}

func TestAssignShift_MissingPersonID(t *testing.T) {
	srv := newTestShiftAssignmentServer()

	_, err := srv.AssignShift(context.Background(), &resourcev1.AssignShiftRequest{
		ShiftId: "sh-1", WorkCenterId: "wc-1", EffectiveFrom: "2026-08-21T00:00:00Z",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("AssignShift() = %v, want InvalidArgument", err)
	}
}

func TestAssignShift_MissingShiftID(t *testing.T) {
	srv := newTestShiftAssignmentServer()

	_, err := srv.AssignShift(context.Background(), &resourcev1.AssignShiftRequest{
		PersonId: "p-1", WorkCenterId: "wc-1", EffectiveFrom: "2026-08-21T00:00:00Z",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("AssignShift() = %v, want InvalidArgument", err)
	}
}

func TestAssignShift_MissingWorkCenterID(t *testing.T) {
	srv := newTestShiftAssignmentServer()

	_, err := srv.AssignShift(context.Background(), &resourcev1.AssignShiftRequest{
		PersonId: "p-1", ShiftId: "sh-1", EffectiveFrom: "2026-08-21T00:00:00Z",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("AssignShift() = %v, want InvalidArgument", err)
	}
}

func TestAssignShift_MissingEffectiveFrom(t *testing.T) {
	srv := newTestShiftAssignmentServer()

	_, err := srv.AssignShift(context.Background(), &resourcev1.AssignShiftRequest{
		PersonId: "p-1", ShiftId: "sh-1", WorkCenterId: "wc-1",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("AssignShift() = %v, want InvalidArgument", err)
	}
}

func TestAssignShift_InvalidEffectiveFrom(t *testing.T) {
	persons := newMockPersonRepo()
	shifts := newMockShiftRepo()
	srv := NewShiftAssignmentServer(persons, shifts, newMockShiftAssignmentRepo())

	person, _ := persons.CreatePerson(context.Background(), "pc-1", "EMP-001", "Jane", "Doe", nil)
	shift, _ := shifts.CreateShift(context.Background(), "Day Shift", "06:00", "14:00", nil)

	_, err := srv.AssignShift(context.Background(), &resourcev1.AssignShiftRequest{
		PersonId: person.ID, ShiftId: shift.ID, WorkCenterId: "wc-1",
		EffectiveFrom: "not-a-date",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("AssignShift() = %v, want InvalidArgument", err)
	}
}

func TestAssignShift_PersonNotFound(t *testing.T) {
	shifts := newMockShiftRepo()
	srv := NewShiftAssignmentServer(newMockPersonRepo(), shifts, newMockShiftAssignmentRepo())

	shift, _ := shifts.CreateShift(context.Background(), "Day Shift", "06:00", "14:00", nil)

	_, err := srv.AssignShift(context.Background(), &resourcev1.AssignShiftRequest{
		PersonId: "non-existent", ShiftId: shift.ID, WorkCenterId: "wc-1",
		EffectiveFrom: "2026-08-21T00:00:00Z",
	})
	if status.Code(err) != codes.NotFound {
		t.Errorf("AssignShift() = %v, want NotFound", err)
	}
}

func TestAssignShift_ShiftNotFound(t *testing.T) {
	persons := newMockPersonRepo()
	srv := NewShiftAssignmentServer(persons, newMockShiftRepo(), newMockShiftAssignmentRepo())

	person, _ := persons.CreatePerson(context.Background(), "pc-1", "EMP-001", "Jane", "Doe", nil)

	_, err := srv.AssignShift(context.Background(), &resourcev1.AssignShiftRequest{
		PersonId: person.ID, ShiftId: "non-existent", WorkCenterId: "wc-1",
		EffectiveFrom: "2026-08-21T00:00:00Z",
	})
	if status.Code(err) != codes.NotFound {
		t.Errorf("AssignShift() = %v, want NotFound", err)
	}
}

// ============================================================================
// GetShiftAssignment
// ============================================================================

func TestGetShiftAssignment_Success(t *testing.T) {
	persons := newMockPersonRepo()
	shifts := newMockShiftRepo()
	assigns := newMockShiftAssignmentRepo()
	srv := NewShiftAssignmentServer(persons, shifts, assigns)

	person, _ := persons.CreatePerson(context.Background(), "pc-1", "EMP-001", "Jane", "Doe", nil)
	shift, _ := shifts.CreateShift(context.Background(), "Day Shift", "06:00", "14:00", nil)

	createResp, _ := srv.AssignShift(context.Background(), &resourcev1.AssignShiftRequest{
		PersonId: person.ID, ShiftId: shift.ID, WorkCenterId: "wc-1",
		EffectiveFrom: "2026-08-21T00:00:00Z",
	})

	getResp, err := srv.GetShiftAssignment(context.Background(), &resourcev1.GetShiftAssignmentRequest{
		Id: createResp.GetAssignment().GetId(),
	})
	if err != nil {
		t.Fatalf("GetShiftAssignment() = %v, want nil", err)
	}
	if getResp.GetAssignment().GetId() != createResp.GetAssignment().GetId() {
		t.Errorf("ID = %s, want %s", getResp.GetAssignment().GetId(), createResp.GetAssignment().GetId())
	}
}

func TestGetShiftAssignment_MissingID(t *testing.T) {
	srv := newTestShiftAssignmentServer()

	_, err := srv.GetShiftAssignment(context.Background(), &resourcev1.GetShiftAssignmentRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("GetShiftAssignment() = %v, want InvalidArgument", err)
	}
}

func TestGetShiftAssignment_NotFound(t *testing.T) {
	srv := newTestShiftAssignmentServer()

	_, err := srv.GetShiftAssignment(context.Background(), &resourcev1.GetShiftAssignmentRequest{Id: "non-existent"})
	if status.Code(err) != codes.NotFound {
		t.Errorf("GetShiftAssignment() = %v, want NotFound", err)
	}
}

// ============================================================================
// ListShiftAssignments
// ============================================================================

func TestListShiftAssignments_All(t *testing.T) {
	assigns := newMockShiftAssignmentRepo()
	srv := NewShiftAssignmentServer(newMockPersonRepo(), newMockShiftRepo(), assigns)

	seedShiftAssignment(assigns, "p-1", "sh-1", "wc-1", time.Now())
	seedShiftAssignment(assigns, "p-2", "sh-1", "wc-1", time.Now())

	resp, err := srv.ListShiftAssignments(context.Background(), &resourcev1.ListShiftAssignmentsRequest{})
	if err != nil {
		t.Fatalf("ListShiftAssignments() = %v, want nil", err)
	}
	if len(resp.GetAssignments()) != 2 {
		t.Errorf("ListShiftAssignments() returned %d, want 2", len(resp.GetAssignments()))
	}
}

func TestListShiftAssignments_FilterByPerson(t *testing.T) {
	assigns := newMockShiftAssignmentRepo()
	srv := NewShiftAssignmentServer(newMockPersonRepo(), newMockShiftRepo(), assigns)

	seedShiftAssignment(assigns, "p-1", "sh-1", "wc-1", time.Now())
	seedShiftAssignment(assigns, "p-2", "sh-1", "wc-1", time.Now())

	resp, err := srv.ListShiftAssignments(context.Background(), &resourcev1.ListShiftAssignmentsRequest{PersonId: "p-1"})
	if err != nil {
		t.Fatalf("ListShiftAssignments(p-1) = %v, want nil", err)
	}
	if len(resp.GetAssignments()) != 1 {
		t.Errorf("ListShiftAssignments(p-1) returned %d, want 1", len(resp.GetAssignments()))
	}
}

func TestListShiftAssignments_FilterByShift(t *testing.T) {
	assigns := newMockShiftAssignmentRepo()
	srv := NewShiftAssignmentServer(newMockPersonRepo(), newMockShiftRepo(), assigns)

	seedShiftAssignment(assigns, "p-1", "sh-1", "wc-1", time.Now())
	seedShiftAssignment(assigns, "p-1", "sh-2", "wc-1", time.Now())

	resp, err := srv.ListShiftAssignments(context.Background(), &resourcev1.ListShiftAssignmentsRequest{ShiftId: "sh-1"})
	if err != nil {
		t.Fatalf("ListShiftAssignments(sh-1) = %v, want nil", err)
	}
	if len(resp.GetAssignments()) != 1 {
		t.Errorf("ListShiftAssignments(sh-1) returned %d, want 1", len(resp.GetAssignments()))
	}
}

func TestListShiftAssignments_FilterByWorkCenter(t *testing.T) {
	assigns := newMockShiftAssignmentRepo()
	srv := NewShiftAssignmentServer(newMockPersonRepo(), newMockShiftRepo(), assigns)

	seedShiftAssignment(assigns, "p-1", "sh-1", "wc-1", time.Now())
	seedShiftAssignment(assigns, "p-1", "sh-1", "wc-2", time.Now())

	resp, err := srv.ListShiftAssignments(context.Background(), &resourcev1.ListShiftAssignmentsRequest{WorkCenterId: "wc-1"})
	if err != nil {
		t.Fatalf("ListShiftAssignments(wc-1) = %v, want nil", err)
	}
	if len(resp.GetAssignments()) != 1 {
		t.Errorf("ListShiftAssignments(wc-1) returned %d, want 1", len(resp.GetAssignments()))
	}
}

func TestListShiftAssignments_FilterByPersonAndShift(t *testing.T) {
	assigns := newMockShiftAssignmentRepo()
	srv := NewShiftAssignmentServer(newMockPersonRepo(), newMockShiftRepo(), assigns)

	seedShiftAssignment(assigns, "p-1", "sh-1", "wc-1", time.Now())
	seedShiftAssignment(assigns, "p-1", "sh-2", "wc-1", time.Now())
	seedShiftAssignment(assigns, "p-2", "sh-1", "wc-1", time.Now())

	resp, err := srv.ListShiftAssignments(context.Background(), &resourcev1.ListShiftAssignmentsRequest{
		PersonId: "p-1", ShiftId: "sh-1",
	})
	if err != nil {
		t.Fatalf("ListShiftAssignments(p-1, sh-1) = %v, want nil", err)
	}
	if len(resp.GetAssignments()) != 1 {
		t.Errorf("ListShiftAssignments(p-1, sh-1) returned %d, want 1", len(resp.GetAssignments()))
	}
}

func TestListShiftAssignments_Empty(t *testing.T) {
	srv := newTestShiftAssignmentServer()

	resp, err := srv.ListShiftAssignments(context.Background(), &resourcev1.ListShiftAssignmentsRequest{})
	if err != nil {
		t.Fatalf("ListShiftAssignments() = %v, want nil", err)
	}
	if len(resp.GetAssignments()) != 0 {
		t.Errorf("ListShiftAssignments() returned %d, want 0", len(resp.GetAssignments()))
	}
}

// ============================================================================
// UnassignShift
// ============================================================================

func TestUnassignShift_Success(t *testing.T) {
	persons := newMockPersonRepo()
	shifts := newMockShiftRepo()
	assigns := newMockShiftAssignmentRepo()
	srv := NewShiftAssignmentServer(persons, shifts, assigns)

	person, _ := persons.CreatePerson(context.Background(), "pc-1", "EMP-001", "Jane", "Doe", nil)
	shift, _ := shifts.CreateShift(context.Background(), "Day Shift", "06:00", "14:00", nil)

	createResp, _ := srv.AssignShift(context.Background(), &resourcev1.AssignShiftRequest{
		PersonId: person.ID, ShiftId: shift.ID, WorkCenterId: "wc-1",
		EffectiveFrom: "2026-08-21T00:00:00Z",
	})

	unassignResp, err := srv.UnassignShift(context.Background(), &resourcev1.UnassignShiftRequest{
		Id: createResp.GetAssignment().GetId(),
	})
	if err != nil {
		t.Fatalf("UnassignShift() = %v, want nil", err)
	}
	if !unassignResp.GetUnassigned() {
		t.Error("Unassigned = false, want true")
	}
}

func TestUnassignShift_NotFound(t *testing.T) {
	srv := newTestShiftAssignmentServer()

	unassignResp, err := srv.UnassignShift(context.Background(), &resourcev1.UnassignShiftRequest{
		Id: "non-existent",
	})
	if err != nil {
		t.Fatalf("UnassignShift() = %v, want nil", err)
	}
	if unassignResp.GetUnassigned() {
		t.Error("Unassigned = true, want false for non-existent ID")
	}
}

func TestUnassignShift_MissingID(t *testing.T) {
	srv := newTestShiftAssignmentServer()

	_, err := srv.UnassignShift(context.Background(), &resourcev1.UnassignShiftRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("UnassignShift() = %v, want InvalidArgument", err)
	}
}

// ============================================================================
// Error-path tests
// ============================================================================

type mockShiftAssignmentRepoErr struct{}

func (m *mockShiftAssignmentRepoErr) AssignShift(_ context.Context, _, _, _ string, _ time.Time) (*db.ShiftAssignment, error) {
	return nil, fmt.Errorf("simulated db error")
}
func (m *mockShiftAssignmentRepoErr) GetShiftAssignment(_ context.Context, _ string) (*db.ShiftAssignment, error) {
	return nil, fmt.Errorf("simulated db error")
}
func (m *mockShiftAssignmentRepoErr) ListShiftAssignments(_ context.Context, _, _, _ string) ([]*db.ShiftAssignment, error) {
	return nil, fmt.Errorf("simulated db error")
}
func (m *mockShiftAssignmentRepoErr) UnassignShift(_ context.Context, _ string) (bool, error) {
	return false, fmt.Errorf("simulated db error")
}

func TestAssignShift_RepoError(t *testing.T) {
	persons := newMockPersonRepo()
	shifts := newMockShiftRepo()
	srv := NewShiftAssignmentServer(persons, shifts, &mockShiftAssignmentRepoErr{})

	person, _ := persons.CreatePerson(context.Background(), "pc-1", "EMP-001", "Jane", "Doe", nil)
	shift, _ := shifts.CreateShift(context.Background(), "Day Shift", "06:00", "14:00", nil)

	_, err := srv.AssignShift(context.Background(), &resourcev1.AssignShiftRequest{
		PersonId: person.ID, ShiftId: shift.ID, WorkCenterId: "wc-1",
		EffectiveFrom: "2026-08-21T00:00:00Z",
	})
	if status.Code(err) != codes.Internal {
		t.Errorf("AssignShift() = %v, want Internal", err)
	}
}

func TestGetShiftAssignment_RepoError(t *testing.T) {
	srv := NewShiftAssignmentServer(newMockPersonRepo(), newMockShiftRepo(), &mockShiftAssignmentRepoErr{})

	_, err := srv.GetShiftAssignment(context.Background(), &resourcev1.GetShiftAssignmentRequest{Id: "sa-1"})
	if status.Code(err) != codes.Internal {
		t.Errorf("GetShiftAssignment() = %v, want Internal", err)
	}
}

func TestListShiftAssignments_RepoError(t *testing.T) {
	srv := NewShiftAssignmentServer(newMockPersonRepo(), newMockShiftRepo(), &mockShiftAssignmentRepoErr{})

	_, err := srv.ListShiftAssignments(context.Background(), &resourcev1.ListShiftAssignmentsRequest{})
	if status.Code(err) != codes.Internal {
		t.Errorf("ListShiftAssignments() = %v, want Internal", err)
	}
}

func TestUnassignShift_RepoError(t *testing.T) {
	srv := NewShiftAssignmentServer(newMockPersonRepo(), newMockShiftRepo(), &mockShiftAssignmentRepoErr{})

	_, err := srv.UnassignShift(context.Background(), &resourcev1.UnassignShiftRequest{Id: "sa-1"})
	if status.Code(err) != codes.Internal {
		t.Errorf("UnassignShift() = %v, want Internal", err)
	}
}
