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

// mockPersonRepo implements db.PersonRepository for testing.
type mockPersonRepo struct {
	persons map[string]*db.Person
	nextID  int
}

func newMockPersonRepo() *mockPersonRepo {
	return &mockPersonRepo{persons: make(map[string]*db.Person), nextID: 1}
}

func (m *mockPersonRepo) CreatePerson(_ context.Context, personClassID, employeeID, firstName, lastName string, email *string) (*db.Person, error) {
	id := fmt.Sprintf("p-%d", m.nextID)
	m.nextID++
	p := &db.Person{
		ID: id, PersonClassID: personClassID, EmployeeID: employeeID,
		FirstName: firstName, LastName: lastName, Email: email,
		Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	m.persons[p.ID] = p
	return p, nil
}

func (m *mockPersonRepo) GetPerson(_ context.Context, id string) (*db.Person, error) {
	p, ok := m.persons[id]
	if !ok {
		return nil, nil
	}
	return p, nil
}

func (m *mockPersonRepo) ListPersons(_ context.Context, personClassID string) ([]*db.Person, error) {
	var result []*db.Person
	for _, p := range m.persons {
		if personClassID == "" || p.PersonClassID == personClassID {
			result = append(result, p)
		}
	}
	return result, nil
}

// ============================================================================
// Create
// ============================================================================

func TestCreatePerson_Success(t *testing.T) {
	srv := NewPersonServer(newMockPersonRepo())

	resp, err := srv.CreatePerson(context.Background(), &resourcev1.CreatePersonRequest{
		PersonClassId: "pc-1",
		EmployeeId:    "EMP-0042",
		FirstName:     "Jane",
		LastName:      "Doe",
		Email:         "jane.doe@example.com",
	})
	if err != nil {
		t.Fatalf("CreatePerson() = %v, want nil", err)
	}
	if resp.GetPerson().GetFirstName() != "Jane" {
		t.Errorf("FirstName = %q, want 'Jane'", resp.GetPerson().GetFirstName())
	}
	if resp.GetPerson().GetLastName() != "Doe" {
		t.Errorf("LastName = %q, want 'Doe'", resp.GetPerson().GetLastName())
	}
	if resp.GetPerson().GetEmployeeId() != "EMP-0042" {
		t.Errorf("EmployeeId = %q, want 'EMP-0042'", resp.GetPerson().GetEmployeeId())
	}
	if resp.GetPerson().GetEmail() != "jane.doe@example.com" {
		t.Errorf("Email = %q, want 'jane.doe@example.com'", resp.GetPerson().GetEmail())
	}
}

func TestCreatePerson_NoEmail(t *testing.T) {
	srv := NewPersonServer(newMockPersonRepo())

	resp, err := srv.CreatePerson(context.Background(), &resourcev1.CreatePersonRequest{
		PersonClassId: "pc-1",
		EmployeeId:    "EMP-0099",
		FirstName:     "John",
		LastName:      "Smith",
	})
	if err != nil {
		t.Fatalf("CreatePerson() = %v, want nil", err)
	}
	if resp.GetPerson().GetEmail() != "" {
		t.Errorf("Email = %q, want empty", resp.GetPerson().GetEmail())
	}
}

func TestCreatePerson_MissingPersonClassID(t *testing.T) {
	srv := NewPersonServer(newMockPersonRepo())

	_, err := srv.CreatePerson(context.Background(), &resourcev1.CreatePersonRequest{
		EmployeeId: "EMP-0042", FirstName: "Jane", LastName: "Doe",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreatePerson() = %v, want InvalidArgument", err)
	}
}

func TestCreatePerson_MissingEmployeeID(t *testing.T) {
	srv := NewPersonServer(newMockPersonRepo())

	_, err := srv.CreatePerson(context.Background(), &resourcev1.CreatePersonRequest{
		PersonClassId: "pc-1", FirstName: "Jane", LastName: "Doe",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreatePerson() = %v, want InvalidArgument", err)
	}
}

func TestCreatePerson_MissingFirstName(t *testing.T) {
	srv := NewPersonServer(newMockPersonRepo())

	_, err := srv.CreatePerson(context.Background(), &resourcev1.CreatePersonRequest{
		PersonClassId: "pc-1", EmployeeId: "EMP-0042", LastName: "Doe",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreatePerson() = %v, want InvalidArgument", err)
	}
}

func TestCreatePerson_MissingLastName(t *testing.T) {
	srv := NewPersonServer(newMockPersonRepo())

	_, err := srv.CreatePerson(context.Background(), &resourcev1.CreatePersonRequest{
		PersonClassId: "pc-1", EmployeeId: "EMP-0042", FirstName: "Jane",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("CreatePerson() = %v, want InvalidArgument", err)
	}
}

// ============================================================================
// Get
// ============================================================================

func TestGetPerson_Success(t *testing.T) {
	srv := NewPersonServer(newMockPersonRepo())

	createResp, _ := srv.CreatePerson(context.Background(), &resourcev1.CreatePersonRequest{
		PersonClassId: "pc-1", EmployeeId: "EMP-0042",
		FirstName: "Jane", LastName: "Doe",
	})

	getResp, err := srv.GetPerson(context.Background(), &resourcev1.GetPersonRequest{
		Id: createResp.GetPerson().GetId(),
	})
	if err != nil {
		t.Fatalf("GetPerson() = %v, want nil", err)
	}
	if getResp.GetPerson().GetId() != createResp.GetPerson().GetId() {
		t.Errorf("ID = %s, want %s", getResp.GetPerson().GetId(), createResp.GetPerson().GetId())
	}
}

func TestGetPerson_MissingID(t *testing.T) {
	srv := NewPersonServer(newMockPersonRepo())

	_, err := srv.GetPerson(context.Background(), &resourcev1.GetPersonRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("GetPerson() = %v, want InvalidArgument", err)
	}
}

func TestGetPerson_NotFound(t *testing.T) {
	srv := NewPersonServer(newMockPersonRepo())

	_, err := srv.GetPerson(context.Background(), &resourcev1.GetPersonRequest{Id: "non-existent"})
	if status.Code(err) != codes.NotFound {
		t.Errorf("GetPerson() = %v, want NotFound", err)
	}
}

// ============================================================================
// List
// ============================================================================

func TestListPersons_Success(t *testing.T) {
	srv := NewPersonServer(newMockPersonRepo())

	_, _ = srv.CreatePerson(context.Background(), &resourcev1.CreatePersonRequest{
		PersonClassId: "pc-1", EmployeeId: "EMP-001", FirstName: "Alice", LastName: "Smith",
	})
	_, _ = srv.CreatePerson(context.Background(), &resourcev1.CreatePersonRequest{
		PersonClassId: "pc-1", EmployeeId: "EMP-002", FirstName: "Bob", LastName: "Jones",
	})

	resp, err := srv.ListPersons(context.Background(), &resourcev1.ListPersonsRequest{})
	if err != nil {
		t.Fatalf("ListPersons() = %v, want nil", err)
	}
	if len(resp.GetPersons()) != 2 {
		t.Errorf("ListPersons() returned %d, want 2", len(resp.GetPersons()))
	}
}

func TestListPersons_FilterByPersonClass(t *testing.T) {
	srv := NewPersonServer(newMockPersonRepo())

	_, _ = srv.CreatePerson(context.Background(), &resourcev1.CreatePersonRequest{
		PersonClassId: "pc-1", EmployeeId: "EMP-001", FirstName: "Alice", LastName: "Smith",
	})
	_, _ = srv.CreatePerson(context.Background(), &resourcev1.CreatePersonRequest{
		PersonClassId: "pc-2", EmployeeId: "EMP-002", FirstName: "Bob", LastName: "Jones",
	})

	resp, err := srv.ListPersons(context.Background(), &resourcev1.ListPersonsRequest{PersonClassId: "pc-1"})
	if err != nil {
		t.Fatalf("ListPersons(pc-1) = %v, want nil", err)
	}
	if len(resp.GetPersons()) != 1 {
		t.Errorf("ListPersons(pc-1) returned %d, want 1", len(resp.GetPersons()))
	}
}

func TestListPersons_Empty(t *testing.T) {
	srv := NewPersonServer(newMockPersonRepo())

	resp, err := srv.ListPersons(context.Background(), &resourcev1.ListPersonsRequest{})
	if err != nil {
		t.Fatalf("ListPersons() = %v, want nil", err)
	}
	if len(resp.GetPersons()) != 0 {
		t.Errorf("ListPersons() returned %d, want 0", len(resp.GetPersons()))
	}
}

// ============================================================================
// Error-path tests
// ============================================================================

// mockPersonRepoErr returns errors on all methods.
type mockPersonRepoErr struct{}

func (m *mockPersonRepoErr) CreatePerson(_ context.Context, _, _, _, _ string, _ *string) (*db.Person, error) {
	return nil, fmt.Errorf("simulated db error")
}

func (m *mockPersonRepoErr) GetPerson(_ context.Context, _ string) (*db.Person, error) {
	return nil, fmt.Errorf("simulated db error")
}

func (m *mockPersonRepoErr) ListPersons(_ context.Context, _ string) ([]*db.Person, error) {
	return nil, fmt.Errorf("simulated db error")
}

func TestCreatePerson_RepoError(t *testing.T) {
	srv := NewPersonServer(&mockPersonRepoErr{})

	_, err := srv.CreatePerson(context.Background(), &resourcev1.CreatePersonRequest{
		PersonClassId: "pc-1", EmployeeId: "EMP-001", FirstName: "Jane", LastName: "Doe",
	})
	if status.Code(err) != codes.Internal {
		t.Errorf("CreatePerson() = %v, want Internal", err)
	}
}

func TestGetPerson_RepoError(t *testing.T) {
	srv := NewPersonServer(&mockPersonRepoErr{})

	_, err := srv.GetPerson(context.Background(), &resourcev1.GetPersonRequest{Id: "p-1"})
	if status.Code(err) != codes.Internal {
		t.Errorf("GetPerson() = %v, want Internal", err)
	}
}

func TestListPersons_RepoError(t *testing.T) {
	srv := NewPersonServer(&mockPersonRepoErr{})

	_, err := srv.ListPersons(context.Background(), &resourcev1.ListPersonsRequest{})
	if status.Code(err) != codes.Internal {
		t.Errorf("ListPersons() = %v, want Internal", err)
	}
}
