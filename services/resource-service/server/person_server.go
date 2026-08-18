package server

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	resourcev1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/resource/v1"
	"github.com/iamraydoan/factoryos/services/resource-service/db"
)

// PersonServer implements Person gRPC RPCs.
type PersonServer struct {
	resourcev1.UnimplementedEquipmentServiceServer
	repo db.PersonRepository
}

// NewPersonServer creates a new PersonServer.
func NewPersonServer(repo db.PersonRepository) *PersonServer {
	return &PersonServer{repo: repo}
}

// CreatePerson creates a new Person.
func (s *PersonServer) CreatePerson(ctx context.Context, req *resourcev1.CreatePersonRequest) (*resourcev1.CreatePersonResponse, error) {
	if req.GetPersonClassId() == "" {
		return nil, status.Error(codes.InvalidArgument, "person_class_id is required")
	}
	if req.GetEmployeeId() == "" {
		return nil, status.Error(codes.InvalidArgument, "employee_id is required")
	}
	if req.GetFirstName() == "" {
		return nil, status.Error(codes.InvalidArgument, "first_name is required")
	}
	if req.GetLastName() == "" {
		return nil, status.Error(codes.InvalidArgument, "last_name is required")
	}

	var email *string
	if req.GetEmail() != "" {
		e := req.GetEmail()
		email = &e
	}

	person, err := s.repo.CreatePerson(ctx, req.GetPersonClassId(), req.GetEmployeeId(), req.GetFirstName(), req.GetLastName(), email)
	if err != nil {
		log.Printf("[PersonServer][ERROR] CreatePerson: %v", err)
		return nil, status.Error(codes.Internal, "failed to create person")
	}

	return &resourcev1.CreatePersonResponse{
		Person: toProtoPerson(person),
	}, nil
}

// GetPerson retrieves a Person by ID.
func (s *PersonServer) GetPerson(ctx context.Context, req *resourcev1.GetPersonRequest) (*resourcev1.GetPersonResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	person, err := s.repo.GetPerson(ctx, req.GetId())
	if err != nil {
		log.Printf("[PersonServer][ERROR] GetPerson(%s): %v", req.GetId(), err)
		return nil, status.Error(codes.Internal, "failed to get person")
	}
	if person == nil {
		return nil, status.Error(codes.NotFound, "person not found")
	}

	return &resourcev1.GetPersonResponse{
		Person: toProtoPerson(person),
	}, nil
}

// ListPersons returns Persons, optionally filtered by PersonClassID.
func (s *PersonServer) ListPersons(ctx context.Context, req *resourcev1.ListPersonsRequest) (*resourcev1.ListPersonsResponse, error) {
	persons, err := s.repo.ListPersons(ctx, req.GetPersonClassId())
	if err != nil {
		log.Printf("[PersonServer][ERROR] ListPersons: %v", err)
		return nil, status.Error(codes.Internal, "failed to list persons")
	}

	protoPersons := make([]*resourcev1.Person, len(persons))
	for i, p := range persons {
		protoPersons[i] = toProtoPerson(p)
	}

	return &resourcev1.ListPersonsResponse{
		Persons: protoPersons,
	}, nil
}
