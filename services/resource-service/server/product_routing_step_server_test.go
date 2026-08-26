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
// Happy-path mocks
// ============================================================================

// mockWorkCenterRepo implements db.WorkCenterRepository for testing.
type mockWorkCenterRepo struct {
	centers map[string]*db.WorkCenter
}

func newMockWorkCenterRepo() *mockWorkCenterRepo {
	return &mockWorkCenterRepo{centers: make(map[string]*db.WorkCenter)}
}

func (m *mockWorkCenterRepo) GetWorkCenter(_ context.Context, id string) (*db.WorkCenter, error) {
	wc, ok := m.centers[id]
	if !ok {
		return nil, nil
	}
	return wc, nil
}

// mockRoutingStepRepo implements db.ProductRoutingStepRepository for testing.
type mockRoutingStepRepo struct {
	steps  map[string][]*db.ProductRoutingStep
	nextID int
}

func newMockRoutingStepRepo() *mockRoutingStepRepo {
	return &mockRoutingStepRepo{steps: make(map[string][]*db.ProductRoutingStep), nextID: 1}
}

func (m *mockRoutingStepRepo) AddRoutingStep(_ context.Context, routingSpecID, workCenterID string, stepNumber int32, estimatedDuration string, description *string) (*db.ProductRoutingStep, error) {
	id := fmt.Sprintf("step-%d", m.nextID)
	m.nextID++
	now := time.Now()
	step := &db.ProductRoutingStep{
		ID: id, RoutingSpecID: routingSpecID, WorkCenterID: workCenterID,
		StepNumber: stepNumber, EstimatedDuration: estimatedDuration,
		Description: description, CreatedAt: now, UpdatedAt: now,
	}
	m.steps[routingSpecID] = append(m.steps[routingSpecID], step)
	return step, nil
}

func (m *mockRoutingStepRepo) ListRoutingSteps(_ context.Context, routingSpecID string) ([]*db.ProductRoutingStep, error) {
	return m.steps[routingSpecID], nil
}

// newTestRoutingStepServer creates a ProductRoutingStepServer with all happy-path mocks.
func newTestRoutingStepServer() *ProductRoutingStepServer {
	return NewProductRoutingStepServer(
		newMockRoutingSpecRepo(),
		newMockWorkCenterRepo(),
		newMockRoutingStepRepo(),
	)
}

// ============================================================================
// AddRoutingStep
// ============================================================================

func TestAddRoutingStep_Success(t *testing.T) {
	specs := newMockRoutingSpecRepo()
	centers := newMockWorkCenterRepo()
	steps := newMockRoutingStepRepo()
	srv := NewProductRoutingStepServer(specs, centers, steps)

	// Create prerequisite routing spec
	class, _ := newMockMaterialClassRepo().CreateMaterialClass(context.Background(), "Finished Good", nil)
	materials := newMockMaterialDefinitionRepo()
	md, _ := materials.CreateMaterialDefinition(context.Background(), class.ID, "Engine Block", "ENG-001", "pcs", nil)
	spec, _ := specs.CreateRoutingSpec(context.Background(), md.ID, "v1.0", nil)

	// Create a work center manually in the mock
	centers.centers["wc-1"] = &db.WorkCenter{ID: "wc-1", Name: "CNC Cell A"}

	resp, err := srv.AddRoutingStep(context.Background(), &resourcev1.AddRoutingStepRequest{
		RoutingSpecId:     spec.ID,
		WorkCenterId:      "wc-1",
		StepNumber:        1,
		EstimatedDuration: "45m",
		Description:       "CNC machining",
	})
	if err != nil {
		t.Fatalf("AddRoutingStep() = %v, want nil", err)
	}
	if resp.GetStep().GetStepNumber() != 1 {
		t.Errorf("StepNumber = %d, want 1", resp.GetStep().GetStepNumber())
	}
	if resp.GetStep().GetEstimatedDuration() != "45m" {
		t.Errorf("EstimatedDuration = %q, want '45m'", resp.GetStep().GetEstimatedDuration())
	}
	if resp.GetStep().GetDescription() != "CNC machining" {
		t.Errorf("Description = %q, want 'CNC machining'", resp.GetStep().GetDescription())
	}
}

func TestAddRoutingStep_NoDescription(t *testing.T) {
	specs := newMockRoutingSpecRepo()
	centers := newMockWorkCenterRepo()
	steps := newMockRoutingStepRepo()
	srv := NewProductRoutingStepServer(specs, centers, steps)

	class, _ := newMockMaterialClassRepo().CreateMaterialClass(context.Background(), "Finished Good", nil)
	materials := newMockMaterialDefinitionRepo()
	md, _ := materials.CreateMaterialDefinition(context.Background(), class.ID, "Engine Block", "ENG-001", "pcs", nil)
	spec, _ := specs.CreateRoutingSpec(context.Background(), md.ID, "v1.0", nil)

	centers.centers["wc-1"] = &db.WorkCenter{ID: "wc-1", Name: "CNC Cell A"}

	resp, err := srv.AddRoutingStep(context.Background(), &resourcev1.AddRoutingStepRequest{
		RoutingSpecId:     spec.ID,
		WorkCenterId:      "wc-1",
		StepNumber:        1,
		EstimatedDuration: "45m",
	})
	if err != nil {
		t.Fatalf("AddRoutingStep() = %v, want nil", err)
	}
	if resp.GetStep().GetDescription() != "" {
		t.Errorf("Description = %q, want ''", resp.GetStep().GetDescription())
	}
}

func TestAddRoutingStep_MissingRoutingSpecID(t *testing.T) {
	srv := newTestRoutingStepServer()

	_, err := srv.AddRoutingStep(context.Background(), &resourcev1.AddRoutingStepRequest{
		WorkCenterId:      "wc-1",
		StepNumber:        1,
		EstimatedDuration: "45m",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("AddRoutingStep() = %v, want InvalidArgument", err)
	}
}

func TestAddRoutingStep_MissingWorkCenterID(t *testing.T) {
	srv := newTestRoutingStepServer()

	_, err := srv.AddRoutingStep(context.Background(), &resourcev1.AddRoutingStepRequest{
		RoutingSpecId:     "rs-1",
		StepNumber:        1,
		EstimatedDuration: "45m",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("AddRoutingStep() = %v, want InvalidArgument", err)
	}
}

func TestAddRoutingStep_StepNumberZero(t *testing.T) {
	srv := newTestRoutingStepServer()

	_, err := srv.AddRoutingStep(context.Background(), &resourcev1.AddRoutingStepRequest{
		RoutingSpecId:     "rs-1",
		WorkCenterId:      "wc-1",
		StepNumber:        0,
		EstimatedDuration: "45m",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("AddRoutingStep() = %v, want InvalidArgument", err)
	}
}

func TestAddRoutingStep_StepNumberNegative(t *testing.T) {
	srv := newTestRoutingStepServer()

	_, err := srv.AddRoutingStep(context.Background(), &resourcev1.AddRoutingStepRequest{
		RoutingSpecId:     "rs-1",
		WorkCenterId:      "wc-1",
		StepNumber:        -1,
		EstimatedDuration: "45m",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("AddRoutingStep() = %v, want InvalidArgument", err)
	}
}

func TestAddRoutingStep_MissingEstimatedDuration(t *testing.T) {
	srv := newTestRoutingStepServer()

	_, err := srv.AddRoutingStep(context.Background(), &resourcev1.AddRoutingStepRequest{
		RoutingSpecId: "rs-1",
		WorkCenterId:  "wc-1",
		StepNumber:    1,
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("AddRoutingStep() = %v, want InvalidArgument", err)
	}
}

func TestAddRoutingStep_RoutingSpecNotFound(t *testing.T) {
	srv := newTestRoutingStepServer()

	_, err := srv.AddRoutingStep(context.Background(), &resourcev1.AddRoutingStepRequest{
		RoutingSpecId:     "non-existent",
		WorkCenterId:      "wc-1",
		StepNumber:        1,
		EstimatedDuration: "45m",
	})
	if status.Code(err) != codes.NotFound {
		t.Errorf("AddRoutingStep() = %v, want NotFound", err)
	}
}

func TestAddRoutingStep_WorkCenterNotFound(t *testing.T) {
	specs := newMockRoutingSpecRepo()
	centers := newMockWorkCenterRepo()
	steps := newMockRoutingStepRepo()
	srv := NewProductRoutingStepServer(specs, centers, steps)

	class, _ := newMockMaterialClassRepo().CreateMaterialClass(context.Background(), "Finished Good", nil)
	materials := newMockMaterialDefinitionRepo()
	md, _ := materials.CreateMaterialDefinition(context.Background(), class.ID, "Engine Block", "ENG-001", "pcs", nil)
	spec, _ := specs.CreateRoutingSpec(context.Background(), md.ID, "v1.0", nil)

	_, err := srv.AddRoutingStep(context.Background(), &resourcev1.AddRoutingStepRequest{
		RoutingSpecId:     spec.ID,
		WorkCenterId:      "non-existent",
		StepNumber:        1,
		EstimatedDuration: "45m",
	})
	if status.Code(err) != codes.NotFound {
		t.Errorf("AddRoutingStep() = %v, want NotFound", err)
	}
}

// ============================================================================
// ListRoutingSteps
// ============================================================================

func TestListRoutingSteps_Success(t *testing.T) {
	specs := newMockRoutingSpecRepo()
	centers := newMockWorkCenterRepo()
	steps := newMockRoutingStepRepo()
	srv := NewProductRoutingStepServer(specs, centers, steps)

	class, _ := newMockMaterialClassRepo().CreateMaterialClass(context.Background(), "Finished Good", nil)
	materials := newMockMaterialDefinitionRepo()
	md, _ := materials.CreateMaterialDefinition(context.Background(), class.ID, "Engine Block", "ENG-001", "pcs", nil)
	spec, _ := specs.CreateRoutingSpec(context.Background(), md.ID, "v1.0", nil)

	centers.centers["wc-1"] = &db.WorkCenter{ID: "wc-1", Name: "CNC Cell A"}
	centers.centers["wc-2"] = &db.WorkCenter{ID: "wc-2", Name: "Wash Station"}

	_, _ = srv.AddRoutingStep(context.Background(), &resourcev1.AddRoutingStepRequest{
		RoutingSpecId: spec.ID, WorkCenterId: "wc-1", StepNumber: 1, EstimatedDuration: "45m",
	})
	_, _ = srv.AddRoutingStep(context.Background(), &resourcev1.AddRoutingStepRequest{
		RoutingSpecId: spec.ID, WorkCenterId: "wc-2", StepNumber: 2, EstimatedDuration: "10m",
	})

	resp, err := srv.ListRoutingSteps(context.Background(), &resourcev1.ListRoutingStepsRequest{
		RoutingSpecId: spec.ID,
	})
	if err != nil {
		t.Fatalf("ListRoutingSteps() = %v, want nil", err)
	}
	if len(resp.GetSteps()) != 2 {
		t.Errorf("ListRoutingSteps() returned %d, want 2", len(resp.GetSteps()))
	}
}

func TestListRoutingSteps_MissingRoutingSpecID(t *testing.T) {
	srv := newTestRoutingStepServer()

	_, err := srv.ListRoutingSteps(context.Background(), &resourcev1.ListRoutingStepsRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("ListRoutingSteps() = %v, want InvalidArgument", err)
	}
}

func TestListRoutingSteps_Empty(t *testing.T) {
	srv := newTestRoutingStepServer()

	resp, err := srv.ListRoutingSteps(context.Background(), &resourcev1.ListRoutingStepsRequest{
		RoutingSpecId: "rs-non-existent",
	})
	if err != nil {
		t.Fatalf("ListRoutingSteps() = %v, want nil", err)
	}
	if len(resp.GetSteps()) != 0 {
		t.Errorf("ListRoutingSteps() returned %d, want 0", len(resp.GetSteps()))
	}
}

// ============================================================================
// Error-path tests
// ============================================================================

type mockRoutingStepRepoErr struct{}

func (m *mockRoutingStepRepoErr) AddRoutingStep(_ context.Context, _, _ string, _ int32, _ string, _ *string) (*db.ProductRoutingStep, error) {
	return nil, fmt.Errorf("simulated db error")
}
func (m *mockRoutingStepRepoErr) ListRoutingSteps(_ context.Context, _ string) ([]*db.ProductRoutingStep, error) {
	return nil, fmt.Errorf("simulated db error")
}

func TestAddRoutingStep_RepoError(t *testing.T) {
	specs := newMockRoutingSpecRepo()
	centers := newMockWorkCenterRepo()
	srv := NewProductRoutingStepServer(specs, centers, &mockRoutingStepRepoErr{})

	class, _ := newMockMaterialClassRepo().CreateMaterialClass(context.Background(), "Finished Good", nil)
	materials := newMockMaterialDefinitionRepo()
	md, _ := materials.CreateMaterialDefinition(context.Background(), class.ID, "Engine Block", "ENG-001", "pcs", nil)
	spec, _ := specs.CreateRoutingSpec(context.Background(), md.ID, "v1.0", nil)

	centers.centers["wc-1"] = &db.WorkCenter{ID: "wc-1", Name: "CNC Cell A"}

	_, err := srv.AddRoutingStep(context.Background(), &resourcev1.AddRoutingStepRequest{
		RoutingSpecId: spec.ID, WorkCenterId: "wc-1", StepNumber: 1, EstimatedDuration: "45m",
	})
	if status.Code(err) != codes.Internal {
		t.Errorf("AddRoutingStep() = %v, want Internal", err)
	}
}

func TestListRoutingSteps_RepoError(t *testing.T) {
	specs := newMockRoutingSpecRepo()
	centers := newMockWorkCenterRepo()
	srv := NewProductRoutingStepServer(specs, centers, &mockRoutingStepRepoErr{})

	_, err := srv.ListRoutingSteps(context.Background(), &resourcev1.ListRoutingStepsRequest{RoutingSpecId: "rs-1"})
	if status.Code(err) != codes.Internal {
		t.Errorf("ListRoutingSteps() = %v, want Internal", err)
	}
}
