package server

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	resourcev1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/resource/v1"
	"github.com/iamraydoan/factoryos/services/resource-service/db"
)

// ProductRoutingStepServer implements Product Routing Step gRPC RPCs.
type ProductRoutingStepServer struct {
	resourcev1.UnimplementedEquipmentServiceServer
	routingSpecs db.ProductRoutingSpecRepository
	workCenters  db.WorkCenterRepository
	steps        db.ProductRoutingStepRepository
}

// NewProductRoutingStepServer creates a new ProductRoutingStepServer.
func NewProductRoutingStepServer(routingSpecs db.ProductRoutingSpecRepository, workCenters db.WorkCenterRepository, steps db.ProductRoutingStepRepository) *ProductRoutingStepServer {
	return &ProductRoutingStepServer{
		routingSpecs: routingSpecs,
		workCenters:  workCenters,
		steps:        steps,
	}
}

// AddRoutingStep adds a step to a Product Routing Spec.
func (s *ProductRoutingStepServer) AddRoutingStep(ctx context.Context, req *resourcev1.AddRoutingStepRequest) (*resourcev1.AddRoutingStepResponse, error) {
	if req.GetRoutingSpecId() == "" {
		return nil, status.Error(codes.InvalidArgument, "routing_spec_id is required")
	}
	if req.GetWorkCenterId() == "" {
		return nil, status.Error(codes.InvalidArgument, "work_center_id is required")
	}
	if req.GetStepNumber() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "step_number must be > 0")
	}
	if req.GetEstimatedDuration() == "" {
		return nil, status.Error(codes.InvalidArgument, "estimated_duration is required")
	}

	// Verify RoutingSpec exists
	spec, err := s.routingSpecs.GetRoutingSpec(ctx, req.GetRoutingSpecId())
	if err != nil {
		log.Printf("[ProductRoutingStepServer][ERROR] GetRoutingSpec(%s): %v", req.GetRoutingSpecId(), err)
		return nil, status.Error(codes.Internal, "failed to verify routing spec")
	}
	if spec == nil {
		return nil, status.Error(codes.NotFound, "routing spec not found")
	}

	// Verify WorkCenter exists
	wc, err := s.workCenters.GetWorkCenter(ctx, req.GetWorkCenterId())
	if err != nil {
		log.Printf("[ProductRoutingStepServer][ERROR] GetWorkCenter(%s): %v", req.GetWorkCenterId(), err)
		return nil, status.Error(codes.Internal, "failed to verify work center")
	}
	if wc == nil {
		return nil, status.Error(codes.NotFound, "work center not found")
	}

	var desc *string
	if req.GetDescription() != "" {
		d := req.GetDescription()
		desc = &d
	}

	step, err := s.steps.AddRoutingStep(ctx, req.GetRoutingSpecId(), req.GetWorkCenterId(), req.GetStepNumber(), req.GetEstimatedDuration(), desc)
	if err != nil {
		log.Printf("[ProductRoutingStepServer][ERROR] AddRoutingStep: %v", err)
		return nil, status.Error(codes.Internal, "failed to add routing step")
	}

	return &resourcev1.AddRoutingStepResponse{
		Step: toProtoRoutingStep(step),
	}, nil
}

// ListRoutingSteps returns all steps for a given Routing Spec.
func (s *ProductRoutingStepServer) ListRoutingSteps(ctx context.Context, req *resourcev1.ListRoutingStepsRequest) (*resourcev1.ListRoutingStepsResponse, error) {
	if req.GetRoutingSpecId() == "" {
		return nil, status.Error(codes.InvalidArgument, "routing_spec_id is required")
	}

	steps, err := s.steps.ListRoutingSteps(ctx, req.GetRoutingSpecId())
	if err != nil {
		log.Printf("[ProductRoutingStepServer][ERROR] ListRoutingSteps(%s): %v", req.GetRoutingSpecId(), err)
		return nil, status.Error(codes.Internal, "failed to list routing steps")
	}

	protoSteps := make([]*resourcev1.ProductRoutingStep, len(steps))
	for i, step := range steps {
		protoSteps[i] = toProtoRoutingStep(step)
	}

	return &resourcev1.ListRoutingStepsResponse{
		Steps: protoSteps,
	}, nil
}
