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

// mockShiftRepo implements db.ShiftRepository for testing.
type mockShiftRepo struct {
	shifts map[string]*db.Shift
	nextID int
}

func newMockShiftRepo() *mockShiftRepo {
	return &mockShiftRepo{shifts: make(map[string]*db.Shift), nextID: 1}
}

func (m *mockShiftRepo) CreateShift(_ context.Context, name, startTime, endTime string, description *string) (*db.Shift, error) {
	id := fmt.Sprintf("sh-%d", m.nextID)
	m.nextID++
	now := time.Now()
	s := &db.Shift{
		ID: id, Name: name, StartTime: startTime, EndTime: endTime,
		Description: description, CreatedAt: now, UpdatedAt: now,
	}
	m.shifts[s.ID] = s
	return s, nil
}

func (m *mockShiftRepo) GetShift(_ context.Context, id string) (*db.Shift, error) {
	s, ok := m.shifts[id]
	if !ok {
		return nil, nil
	}
	return s, nil
}

func (m *mockShiftRepo) ListShifts(_ context.Context) ([]*db.Shift, error) {
	var result []*db.Shift
	for _, s := range m.shifts {
		result = append(result, s)
	}
	return result, nil
}

// ============================================================================
// CreateShift
// ============================================================================

func TestCreateShift_Success(t *testing.T) {
	srv := NewShiftServer(newMockShiftRepo())

	resp, err := srv.CreateShift(context.Background(), &resourcev1.CreateShiftRequest{
		Name: "Day Shift", StartTime: "06:00", EndTime: "14:00",
	})
	if err != nil {
		t.Fatalf("CreateShift() = %v, want nil", err)
	}
	if resp.GetShift().GetName() != "Day Shift" {
		t.Errorf("Name = %q, want 'Day Shift'", resp.GetShift().GetName())
	}
	if resp.GetShift().GetStartTime() != "06:00" {
		t.Errorf("StartTime = %q, want '06:00'", resp.GetShift().GetStartTime())
	}
	if resp.GetShift().GetEndTime() != "14:00" {
		t.Errorf("EndTime = %q, want '14:00'", resp.GetShift().GetEndTime())
	}
}

func TestCreateShift_WithDescription(t *testing.T) {
	srv := NewShiftServer(newMockShiftRepo())

	resp, err := srv.CreateShift(context.Background(), &resourcev1.CreateShiftRequest{
		Name: "Night Shift", StartTime: "22:00", EndTime: "06:00",
		Description: "Overnight shift, crosses midnight",
	})
	if err != nil {
		t.Fatalf("CreateShift() = %v, want nil", err)
	}
	if resp.GetShift().GetDescription() != "Overnight shift, crosses midnight" {
		t.Errorf("Description = %q, want 'Overnight shift, crosses midnight'", resp.GetShift().GetDescription())
	}
}

func TestCreateShift_MissingName(t *testing.T) {
	srv := NewShiftServer(newMockShiftRepo())

	_, err := srv.CreateShift(context.Background(), &resourcev1.CreateShiftRequest{
		StartTime: "06:00", EndTime: "14:00",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreateShift() = %v, want InvalidArgument", err)
	}
}

func TestCreateShift_MissingStartTime(t *testing.T) {
	srv := NewShiftServer(newMockShiftRepo())

	_, err := srv.CreateShift(context.Background(), &resourcev1.CreateShiftRequest{
		Name: "Day Shift", EndTime: "14:00",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreateShift() = %v, want InvalidArgument", err)
	}
}

func TestCreateShift_MissingEndTime(t *testing.T) {
	srv := NewShiftServer(newMockShiftRepo())

	_, err := srv.CreateShift(context.Background(), &resourcev1.CreateShiftRequest{
		Name: "Day Shift", StartTime: "06:00",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreateShift() = %v, want InvalidArgument", err)
	}
}

func TestCreateShift_InvalidStartTime(t *testing.T) {
	srv := NewShiftServer(newMockShiftRepo())

	_, err := srv.CreateShift(context.Background(), &resourcev1.CreateShiftRequest{
		Name: "Day Shift", StartTime: "not-a-time", EndTime: "14:00",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreateShift() = %v, want InvalidArgument", err)
	}
}

func TestCreateShift_InvalidEndTime(t *testing.T) {
	srv := NewShiftServer(newMockShiftRepo())

	_, err := srv.CreateShift(context.Background(), &resourcev1.CreateShiftRequest{
		Name: "Day Shift", StartTime: "06:00", EndTime: "25:00",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreateShift() = %v, want InvalidArgument", err)
	}
}

// ============================================================================
// GetShift
// ============================================================================

func TestGetShift_Success(t *testing.T) {
	srv := NewShiftServer(newMockShiftRepo())

	createResp, _ := srv.CreateShift(context.Background(), &resourcev1.CreateShiftRequest{
		Name: "Day Shift", StartTime: "06:00", EndTime: "14:00",
	})

	getResp, err := srv.GetShift(context.Background(), &resourcev1.GetShiftRequest{
		Id: createResp.GetShift().GetId(),
	})
	if err != nil {
		t.Fatalf("GetShift() = %v, want nil", err)
	}
	if getResp.GetShift().GetId() != createResp.GetShift().GetId() {
		t.Errorf("ID = %s, want %s", getResp.GetShift().GetId(), createResp.GetShift().GetId())
	}
}

func TestGetShift_MissingID(t *testing.T) {
	srv := NewShiftServer(newMockShiftRepo())

	_, err := srv.GetShift(context.Background(), &resourcev1.GetShiftRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("GetShift() = %v, want InvalidArgument", err)
	}
}

func TestGetShift_NotFound(t *testing.T) {
	srv := NewShiftServer(newMockShiftRepo())

	_, err := srv.GetShift(context.Background(), &resourcev1.GetShiftRequest{Id: "non-existent"})
	if status.Code(err) != codes.NotFound {
		t.Errorf("GetShift() = %v, want NotFound", err)
	}
}

// ============================================================================
// ListShifts
// ============================================================================

func TestListShifts_Success(t *testing.T) {
	srv := NewShiftServer(newMockShiftRepo())

	_, _ = srv.CreateShift(context.Background(), &resourcev1.CreateShiftRequest{
		Name: "Day Shift", StartTime: "06:00", EndTime: "14:00",
	})
	_, _ = srv.CreateShift(context.Background(), &resourcev1.CreateShiftRequest{
		Name: "Night Shift", StartTime: "22:00", EndTime: "06:00",
	})

	resp, err := srv.ListShifts(context.Background(), &resourcev1.ListShiftsRequest{})
	if err != nil {
		t.Fatalf("ListShifts() = %v, want nil", err)
	}
	if len(resp.GetShifts()) != 2 {
		t.Errorf("ListShifts() returned %d, want 2", len(resp.GetShifts()))
	}
}

func TestListShifts_Empty(t *testing.T) {
	srv := NewShiftServer(newMockShiftRepo())

	resp, err := srv.ListShifts(context.Background(), &resourcev1.ListShiftsRequest{})
	if err != nil {
		t.Fatalf("ListShifts() = %v, want nil", err)
	}
	if len(resp.GetShifts()) != 0 {
		t.Errorf("ListShifts() returned %d, want 0", len(resp.GetShifts()))
	}
}

// ============================================================================
// Error-path tests
// ============================================================================

type mockShiftRepoErr struct{}

func (m *mockShiftRepoErr) CreateShift(_ context.Context, _, _, _ string, _ *string) (*db.Shift, error) {
	return nil, fmt.Errorf("simulated db error")
}
func (m *mockShiftRepoErr) GetShift(_ context.Context, _ string) (*db.Shift, error) {
	return nil, fmt.Errorf("simulated db error")
}
func (m *mockShiftRepoErr) ListShifts(_ context.Context) ([]*db.Shift, error) {
	return nil, fmt.Errorf("simulated db error")
}

func TestCreateShift_RepoError(t *testing.T) {
	srv := NewShiftServer(&mockShiftRepoErr{})
	_, err := srv.CreateShift(context.Background(), &resourcev1.CreateShiftRequest{
		Name: "Day Shift", StartTime: "06:00", EndTime: "14:00",
	})
	if status.Code(err) != codes.Internal {
		t.Errorf("CreateShift() = %v, want Internal", err)
	}
}

func TestGetShift_RepoError(t *testing.T) {
	srv := NewShiftServer(&mockShiftRepoErr{})
	_, err := srv.GetShift(context.Background(), &resourcev1.GetShiftRequest{Id: "sh-1"})
	if status.Code(err) != codes.Internal {
		t.Errorf("GetShift() = %v, want Internal", err)
	}
}

func TestListShifts_RepoError(t *testing.T) {
	srv := NewShiftServer(&mockShiftRepoErr{})
	_, err := srv.ListShifts(context.Background(), &resourcev1.ListShiftsRequest{})
	if status.Code(err) != codes.Internal {
		t.Errorf("ListShifts() = %v, want Internal", err)
	}
}
