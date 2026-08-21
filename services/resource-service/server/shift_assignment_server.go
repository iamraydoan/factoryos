package server

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	resourcev1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/resource/v1"
	"github.com/iamraydoan/factoryos/services/resource-service/db"
)

// ShiftAssignmentServer implements Shift Assignment gRPC RPCs.
type ShiftAssignmentServer struct {
	resourcev1.UnimplementedEquipmentServiceServer
	persons db.PersonRepository
	shifts  db.ShiftRepository
	assigns db.ShiftAssignmentRepository
}

// NewShiftAssignmentServer creates a new ShiftAssignmentServer.
func NewShiftAssignmentServer(
	persons db.PersonRepository,
	shifts db.ShiftRepository,
	assigns db.ShiftAssignmentRepository,
) *ShiftAssignmentServer {
	return &ShiftAssignmentServer{
		persons: persons,
		shifts:  shifts,
		assigns: assigns,
	}
}

// AssignShift assigns a Person to a Shift at a Work Center.
func (s *ShiftAssignmentServer) AssignShift(ctx context.Context, req *resourcev1.AssignShiftRequest) (*resourcev1.AssignShiftResponse, error) {
	if req.GetPersonId() == "" {
		return nil, status.Error(codes.InvalidArgument, "person_id is required")
	}
	if req.GetShiftId() == "" {
		return nil, status.Error(codes.InvalidArgument, "shift_id is required")
	}
	if req.GetWorkCenterId() == "" {
		return nil, status.Error(codes.InvalidArgument, "work_center_id is required")
	}
	if req.GetEffectiveFrom() == "" {
		return nil, status.Error(codes.InvalidArgument, "effective_from is required (RFC3339 timestamp, e.g., 2026-08-21T00:00:00Z)")
	}

	// Verify Person exists
	person, err := s.persons.GetPerson(ctx, req.GetPersonId())
	if err != nil {
		log.Printf("[ShiftAssignmentServer][ERROR] GetPerson(%s): %v", req.GetPersonId(), err)
		return nil, status.Error(codes.Internal, "failed to verify person")
	}
	if person == nil {
		return nil, status.Error(codes.NotFound, "person not found")
	}

	// Verify Shift exists
	shift, err := s.shifts.GetShift(ctx, req.GetShiftId())
	if err != nil {
		log.Printf("[ShiftAssignmentServer][ERROR] GetShift(%s): %v", req.GetShiftId(), err)
		return nil, status.Error(codes.Internal, "failed to verify shift")
	}
	if shift == nil {
		return nil, status.Error(codes.NotFound, "shift not found")
	}

	// Parse effective_from
	effectiveFrom, err := time.Parse(time.RFC3339, req.GetEffectiveFrom())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid effective_from %q: use RFC3339 format (e.g., 2026-08-21T00:00:00Z)", req.GetEffectiveFrom())
	}

	// Verify Work Center exists (rely on DB FK constraint)
	sa, err := s.assigns.AssignShift(ctx, req.GetPersonId(), req.GetShiftId(), req.GetWorkCenterId(), effectiveFrom)
	if err != nil {
		log.Printf("[ShiftAssignmentServer][ERROR] AssignShift: %v", err)
		return nil, status.Error(codes.Internal, "failed to assign shift")
	}

	return &resourcev1.AssignShiftResponse{
		Assignment: toProtoShiftAssignment(sa),
	}, nil
}

// GetShiftAssignment retrieves a Shift Assignment by ID.
func (s *ShiftAssignmentServer) GetShiftAssignment(ctx context.Context, req *resourcev1.GetShiftAssignmentRequest) (*resourcev1.GetShiftAssignmentResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	sa, err := s.assigns.GetShiftAssignment(ctx, req.GetId())
	if err != nil {
		log.Printf("[ShiftAssignmentServer][ERROR] GetShiftAssignment(%s): %v", req.GetId(), err)
		return nil, status.Error(codes.Internal, "failed to get shift assignment")
	}
	if sa == nil {
		return nil, status.Error(codes.NotFound, "shift assignment not found")
	}

	return &resourcev1.GetShiftAssignmentResponse{
		Assignment: toProtoShiftAssignment(sa),
	}, nil
}

// ListShiftAssignments returns Shift Assignments, optionally filtered.
func (s *ShiftAssignmentServer) ListShiftAssignments(ctx context.Context, req *resourcev1.ListShiftAssignmentsRequest) (*resourcev1.ListShiftAssignmentsResponse, error) {
	assignments, err := s.assigns.ListShiftAssignments(ctx, req.GetPersonId(), req.GetShiftId(), req.GetWorkCenterId())
	if err != nil {
		log.Printf("[ShiftAssignmentServer][ERROR] ListShiftAssignments: %v", err)
		return nil, status.Error(codes.Internal, "failed to list shift assignments")
	}

	protoAssignments := make([]*resourcev1.ShiftAssignment, len(assignments))
	for i, sa := range assignments {
		protoAssignments[i] = toProtoShiftAssignment(sa)
	}

	return &resourcev1.ListShiftAssignmentsResponse{
		Assignments: protoAssignments,
	}, nil
}

// UnassignShift deletes a Shift Assignment by ID.
func (s *ShiftAssignmentServer) UnassignShift(ctx context.Context, req *resourcev1.UnassignShiftRequest) (*resourcev1.UnassignShiftResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	unassigned, err := s.assigns.UnassignShift(ctx, req.GetId())
	if err != nil {
		log.Printf("[ShiftAssignmentServer][ERROR] UnassignShift(%s): %v", req.GetId(), err)
		return nil, status.Error(codes.Internal, "failed to unassign shift")
	}

	return &resourcev1.UnassignShiftResponse{
		Unassigned: unassigned,
	}, nil
}
