package server

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	resourcev1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/resource/v1"
	"github.com/iamraydoan/factoryos/services/resource-service/db"
)

// PersonClassServer implements Person Class gRPC RPCs.
type PersonClassServer struct {
	resourcev1.UnimplementedEquipmentServiceServer
	repo db.PersonClassRepository
}

// NewPersonClassServer creates a new PersonClassServer.
func NewPersonClassServer(repo db.PersonClassRepository) *PersonClassServer {
	return &PersonClassServer{repo: repo}
}

// CreatePersonClass creates a new Person Class.
func (s *PersonClassServer) CreatePersonClass(ctx context.Context, req *resourcev1.CreatePersonClassRequest) (*resourcev1.CreatePersonClassResponse, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	var desc *string
	if req.GetDescription() != "" {
		d := req.GetDescription()
		desc = &d
	}

	pc, err := s.repo.CreatePersonClass(ctx, req.GetName(), desc)
	if err != nil {
		log.Printf("[PersonClassServer][ERROR] CreatePersonClass: %v", err)
		return nil, status.Error(codes.Internal, "failed to create person class")
	}

	return &resourcev1.CreatePersonClassResponse{
		PersonClass: toProtoPersonClass(pc),
	}, nil
}

// GetPersonClass retrieves a Person Class by ID.
func (s *PersonClassServer) GetPersonClass(ctx context.Context, req *resourcev1.GetPersonClassRequest) (*resourcev1.GetPersonClassResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	pc, err := s.repo.GetPersonClass(ctx, req.GetId())
	if err != nil {
		log.Printf("[PersonClassServer][ERROR] GetPersonClass(%s): %v", req.GetId(), err)
		return nil, status.Error(codes.Internal, "failed to get person class")
	}
	if pc == nil {
		return nil, status.Error(codes.NotFound, "person class not found")
	}

	return &resourcev1.GetPersonClassResponse{
		PersonClass: toProtoPersonClass(pc),
	}, nil
}

// ListPersonClasses returns all Person Classes.
func (s *PersonClassServer) ListPersonClasses(ctx context.Context, req *resourcev1.ListPersonClassesRequest) (*resourcev1.ListPersonClassesResponse, error) {
	classes, err := s.repo.ListPersonClasses(ctx)
	if err != nil {
		log.Printf("[PersonClassServer][ERROR] ListPersonClasses: %v", err)
		return nil, status.Error(codes.Internal, "failed to list person classes")
	}

	protoClasses := make([]*resourcev1.PersonClass, len(classes))
	for i, pc := range classes {
		protoClasses[i] = toProtoPersonClass(pc)
	}

	return &resourcev1.ListPersonClassesResponse{
		PersonClasses: protoClasses,
	}, nil
}
