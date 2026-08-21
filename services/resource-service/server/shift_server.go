package server

import (
	"context"
	"log"
	"regexp"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	resourcev1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/resource/v1"
	"github.com/iamraydoan/factoryos/services/resource-service/db"
)

// timeHHMMRegex validates "HH:MM" format (00:00 – 23:59).
var timeHHMMRegex = regexp.MustCompile(`^([01]\d|2[0-3]):([0-5]\d)$`)

// ShiftServer implements Shift gRPC RPCs.
type ShiftServer struct {
	resourcev1.UnimplementedEquipmentServiceServer
	repo db.ShiftRepository
}

// NewShiftServer creates a new ShiftServer.
func NewShiftServer(repo db.ShiftRepository) *ShiftServer {
	return &ShiftServer{repo: repo}
}

// CreateShift creates a new Shift.
func (s *ShiftServer) CreateShift(ctx context.Context, req *resourcev1.CreateShiftRequest) (*resourcev1.CreateShiftResponse, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.GetStartTime() == "" {
		return nil, status.Error(codes.InvalidArgument, "start_time is required (HH:MM, e.g., 06:00)")
	}
	if req.GetEndTime() == "" {
		return nil, status.Error(codes.InvalidArgument, "end_time is required (HH:MM, e.g., 14:00)")
	}

	if !timeHHMMRegex.MatchString(req.GetStartTime()) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid start_time %q: use HH:MM format (e.g., 06:00)", req.GetStartTime())
	}
	if !timeHHMMRegex.MatchString(req.GetEndTime()) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid end_time %q: use HH:MM format (e.g., 14:00)", req.GetEndTime())
	}

	var desc *string
	if req.GetDescription() != "" {
		d := req.GetDescription()
		desc = &d
	}

	shift, err := s.repo.CreateShift(ctx, req.GetName(), req.GetStartTime(), req.GetEndTime(), desc)
	if err != nil {
		log.Printf("[ShiftServer][ERROR] CreateShift: %v", err)
		return nil, status.Error(codes.Internal, "failed to create shift")
	}

	return &resourcev1.CreateShiftResponse{
		Shift: toProtoShift(shift),
	}, nil
}

// GetShift retrieves a Shift by ID.
func (s *ShiftServer) GetShift(ctx context.Context, req *resourcev1.GetShiftRequest) (*resourcev1.GetShiftResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	shift, err := s.repo.GetShift(ctx, req.GetId())
	if err != nil {
		log.Printf("[ShiftServer][ERROR] GetShift(%s): %v", req.GetId(), err)
		return nil, status.Error(codes.Internal, "failed to get shift")
	}
	if shift == nil {
		return nil, status.Error(codes.NotFound, "shift not found")
	}

	return &resourcev1.GetShiftResponse{
		Shift: toProtoShift(shift),
	}, nil
}

// ListShifts returns all Shifts.
func (s *ShiftServer) ListShifts(ctx context.Context, req *resourcev1.ListShiftsRequest) (*resourcev1.ListShiftsResponse, error) {
	shifts, err := s.repo.ListShifts(ctx)
	if err != nil {
		log.Printf("[ShiftServer][ERROR] ListShifts: %v", err)
		return nil, status.Error(codes.Internal, "failed to list shifts")
	}

	protoShifts := make([]*resourcev1.Shift, len(shifts))
	for i, sh := range shifts {
		protoShifts[i] = toProtoShift(sh)
	}

	return &resourcev1.ListShiftsResponse{
		Shifts: protoShifts,
	}, nil
}
