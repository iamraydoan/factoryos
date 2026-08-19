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

// mockQualificationRepo implements db.QualificationRepository for testing.
type mockQualificationRepo struct {
	records map[string]*db.QualificationRecord
	nextID  int
}

func newMockQualificationRepo() *mockQualificationRepo {
	return &mockQualificationRepo{records: make(map[string]*db.QualificationRecord), nextID: 1}
}

func (m *mockQualificationRepo) QualifyPerson(_ context.Context, personID, personClassID, workCenterID string, expiresAt *time.Time) (*db.QualificationRecord, error) {
	id := fmt.Sprintf("qr-%d", m.nextID)
	m.nextID++
	now := time.Now()
	qr := &db.QualificationRecord{
		ID: id, PersonID: personID, PersonClassID: personClassID,
		WorkCenterID: workCenterID, CertifiedAt: now, ExpiresAt: expiresAt,
		CreatedAt: now, UpdatedAt: now,
	}
	m.records[qr.ID] = qr
	return qr, nil
}

func (m *mockQualificationRepo) GetQualification(_ context.Context, id string) (*db.QualificationRecord, error) {
	qr, ok := m.records[id]
	if !ok {
		return nil, nil
	}
	return qr, nil
}

func (m *mockQualificationRepo) ListQualifications(_ context.Context, personID, workCenterID string) ([]*db.QualificationRecord, error) {
	var result []*db.QualificationRecord
	for _, qr := range m.records {
		matchPerson := personID == "" || qr.PersonID == personID
		matchWC := workCenterID == "" || qr.WorkCenterID == workCenterID
		if matchPerson && matchWC {
			result = append(result, qr)
		}
	}
	return result, nil
}

func (m *mockQualificationRepo) RevokeQualification(_ context.Context, id string) (bool, error) {
	if _, ok := m.records[id]; !ok {
		return false, nil
	}
	delete(m.records, id)
	return true, nil
}

// ============================================================================
// QualifyPerson
// ============================================================================

func TestQualifyPerson_Success(t *testing.T) {
	persons := newMockPersonRepo()
	personClasses := newMockPersonClassRepo()
	srv := NewQualificationServer(persons, personClasses, newMockWorkUnitRepo(), newMockQualificationRepo())

	person, _ := persons.CreatePerson(context.Background(), "pc-1", "EMP-001", "Jane", "Doe", nil)
	personClass, _ := personClasses.CreatePersonClass(context.Background(), "Operator", nil)

	resp, err := srv.QualifyPerson(context.Background(), &resourcev1.QualifyPersonRequest{
		PersonId:      person.ID,
		PersonClassId: personClass.ID,
		WorkCenterId:  "wc-1",
	})
	if err != nil {
		t.Fatalf("QualifyPerson() = %v, want nil", err)
	}
	if resp.GetQualification().GetPersonId() != person.ID {
		t.Errorf("PersonId = %q, want %q", resp.GetQualification().GetPersonId(), person.ID)
	}
	if resp.GetQualification().GetPersonClassId() != personClass.ID {
		t.Errorf("PersonClassId = %q, want %q", resp.GetQualification().GetPersonClassId(), personClass.ID)
	}
	if resp.GetQualification().GetExpiresAt() != nil {
		t.Errorf("ExpiresAt = %v, want nil (no expiry)", resp.GetQualification().GetExpiresAt())
	}
}

func TestQualifyPerson_WithExpiry(t *testing.T) {
	persons := newMockPersonRepo()
	personClasses := newMockPersonClassRepo()
	srv := NewQualificationServer(persons, personClasses, newMockWorkUnitRepo(), newMockQualificationRepo())

	person, _ := persons.CreatePerson(context.Background(), "pc-1", "EMP-001", "Jane", "Doe", nil)
	personClass, _ := personClasses.CreatePersonClass(context.Background(), "Operator", nil)

	resp, err := srv.QualifyPerson(context.Background(), &resourcev1.QualifyPersonRequest{
		PersonId:      person.ID,
		PersonClassId: personClass.ID,
		WorkCenterId:  "wc-1",
		ExpiresAt:     "2027-12-31T23:59:59Z",
	})
	if err != nil {
		t.Fatalf("QualifyPerson() = %v, want nil", err)
	}
	if resp.GetQualification().GetExpiresAt() == nil {
		t.Fatal("ExpiresAt should not be nil")
	}
}

func TestQualifyPerson_InvalidExpiryFormat(t *testing.T) {
	persons := newMockPersonRepo()
	personClasses := newMockPersonClassRepo()
	srv := NewQualificationServer(persons, personClasses, newMockWorkUnitRepo(), newMockQualificationRepo())

	person, _ := persons.CreatePerson(context.Background(), "pc-1", "EMP-001", "Jane", "Doe", nil)
	personClass, _ := personClasses.CreatePersonClass(context.Background(), "Operator", nil)

	_, err := srv.QualifyPerson(context.Background(), &resourcev1.QualifyPersonRequest{
		PersonId:      person.ID,
		PersonClassId: personClass.ID,
		WorkCenterId:  "wc-1",
		ExpiresAt:     "not-a-date",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("QualifyPerson() = %v, want InvalidArgument", err)
	}
}

func TestQualifyPerson_MissingPersonID(t *testing.T) {
	srv := NewQualificationServer(newMockPersonRepo(), newMockPersonClassRepo(), newMockWorkUnitRepo(), newMockQualificationRepo())

	_, err := srv.QualifyPerson(context.Background(), &resourcev1.QualifyPersonRequest{
		PersonClassId: "pc-1", WorkCenterId: "wc-1",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("QualifyPerson() = %v, want InvalidArgument", err)
	}
}

func TestQualifyPerson_MissingPersonClassID(t *testing.T) {
	srv := NewQualificationServer(newMockPersonRepo(), newMockPersonClassRepo(), newMockWorkUnitRepo(), newMockQualificationRepo())

	_, err := srv.QualifyPerson(context.Background(), &resourcev1.QualifyPersonRequest{
		PersonId: "p-1", WorkCenterId: "wc-1",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("QualifyPerson() = %v, want InvalidArgument", err)
	}
}

func TestQualifyPerson_MissingWorkCenterID(t *testing.T) {
	srv := NewQualificationServer(newMockPersonRepo(), newMockPersonClassRepo(), newMockWorkUnitRepo(), newMockQualificationRepo())

	_, err := srv.QualifyPerson(context.Background(), &resourcev1.QualifyPersonRequest{
		PersonId: "p-1", PersonClassId: "pc-1",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("QualifyPerson() = %v, want InvalidArgument", err)
	}
}

func TestQualifyPerson_PersonNotFound(t *testing.T) {
	personClasses := newMockPersonClassRepo()
	srv := NewQualificationServer(newMockPersonRepo(), personClasses, newMockWorkUnitRepo(), newMockQualificationRepo())

	personClass, _ := personClasses.CreatePersonClass(context.Background(), "Operator", nil)

	_, err := srv.QualifyPerson(context.Background(), &resourcev1.QualifyPersonRequest{
		PersonId: "non-existent", PersonClassId: personClass.ID, WorkCenterId: "wc-1",
	})
	if status.Code(err) != codes.NotFound {
		t.Errorf("QualifyPerson() = %v, want NotFound", err)
	}
}

func TestQualifyPerson_PersonClassNotFound(t *testing.T) {
	persons := newMockPersonRepo()
	srv := NewQualificationServer(persons, newMockPersonClassRepo(), newMockWorkUnitRepo(), newMockQualificationRepo())

	person, _ := persons.CreatePerson(context.Background(), "pc-1", "EMP-001", "Jane", "Doe", nil)

	_, err := srv.QualifyPerson(context.Background(), &resourcev1.QualifyPersonRequest{
		PersonId: person.ID, PersonClassId: "non-existent", WorkCenterId: "wc-1",
	})
	if status.Code(err) != codes.NotFound {
		t.Errorf("QualifyPerson() = %v, want NotFound", err)
	}
}

// ============================================================================
// GetQualification
// ============================================================================

func TestGetQualification_Success(t *testing.T) {
	persons := newMockPersonRepo()
	personClasses := newMockPersonClassRepo()
	quals := newMockQualificationRepo()
	srv := NewQualificationServer(persons, personClasses, newMockWorkUnitRepo(), quals)

	person, _ := persons.CreatePerson(context.Background(), "pc-1", "EMP-001", "Jane", "Doe", nil)
	personClass, _ := personClasses.CreatePersonClass(context.Background(), "Operator", nil)

	createResp, _ := srv.QualifyPerson(context.Background(), &resourcev1.QualifyPersonRequest{
		PersonId: person.ID, PersonClassId: personClass.ID, WorkCenterId: "wc-1",
	})

	getResp, err := srv.GetQualification(context.Background(), &resourcev1.GetQualificationRequest{
		Id: createResp.GetQualification().GetId(),
	})
	if err != nil {
		t.Fatalf("GetQualification() = %v, want nil", err)
	}
	if getResp.GetQualification().GetId() != createResp.GetQualification().GetId() {
		t.Errorf("ID = %s, want %s", getResp.GetQualification().GetId(), createResp.GetQualification().GetId())
	}
}

func TestGetQualification_MissingID(t *testing.T) {
	srv := NewQualificationServer(newMockPersonRepo(), newMockPersonClassRepo(), newMockWorkUnitRepo(), newMockQualificationRepo())

	_, err := srv.GetQualification(context.Background(), &resourcev1.GetQualificationRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("GetQualification() = %v, want InvalidArgument", err)
	}
}

func TestGetQualification_NotFound(t *testing.T) {
	srv := NewQualificationServer(newMockPersonRepo(), newMockPersonClassRepo(), newMockWorkUnitRepo(), newMockQualificationRepo())

	_, err := srv.GetQualification(context.Background(), &resourcev1.GetQualificationRequest{Id: "non-existent"})
	if status.Code(err) != codes.NotFound {
		t.Errorf("GetQualification() = %v, want NotFound", err)
	}
}

// ============================================================================
// ListQualifications
// ============================================================================

func TestListQualifications_All(t *testing.T) {
	persons := newMockPersonRepo()
	personClasses := newMockPersonClassRepo()
	quals := newMockQualificationRepo()
	srv := NewQualificationServer(persons, personClasses, newMockWorkUnitRepo(), quals)

	person1, _ := persons.CreatePerson(context.Background(), "pc-1", "EMP-001", "Jane", "Doe", nil)
	person2, _ := persons.CreatePerson(context.Background(), "pc-1", "EMP-002", "Bob", "Smith", nil)
	personClass, _ := personClasses.CreatePersonClass(context.Background(), "Operator", nil)

	_, _ = srv.QualifyPerson(context.Background(), &resourcev1.QualifyPersonRequest{
		PersonId: person1.ID, PersonClassId: personClass.ID, WorkCenterId: "wc-1",
	})
	_, _ = srv.QualifyPerson(context.Background(), &resourcev1.QualifyPersonRequest{
		PersonId: person2.ID, PersonClassId: personClass.ID, WorkCenterId: "wc-1",
	})

	resp, err := srv.ListQualifications(context.Background(), &resourcev1.ListQualificationsRequest{})
	if err != nil {
		t.Fatalf("ListQualifications() = %v, want nil", err)
	}
	if len(resp.GetQualifications()) != 2 {
		t.Errorf("ListQualifications() returned %d, want 2", len(resp.GetQualifications()))
	}
}

func TestListQualifications_FilterByPerson(t *testing.T) {
	persons := newMockPersonRepo()
	personClasses := newMockPersonClassRepo()
	quals := newMockQualificationRepo()
	srv := NewQualificationServer(persons, personClasses, newMockWorkUnitRepo(), quals)

	person1, _ := persons.CreatePerson(context.Background(), "pc-1", "EMP-001", "Jane", "Doe", nil)
	person2, _ := persons.CreatePerson(context.Background(), "pc-1", "EMP-002", "Bob", "Smith", nil)
	personClass, _ := personClasses.CreatePersonClass(context.Background(), "Operator", nil)

	_, _ = srv.QualifyPerson(context.Background(), &resourcev1.QualifyPersonRequest{
		PersonId: person1.ID, PersonClassId: personClass.ID, WorkCenterId: "wc-1",
	})
	_, _ = srv.QualifyPerson(context.Background(), &resourcev1.QualifyPersonRequest{
		PersonId: person2.ID, PersonClassId: personClass.ID, WorkCenterId: "wc-1",
	})

	resp, err := srv.ListQualifications(context.Background(), &resourcev1.ListQualificationsRequest{PersonId: person1.ID})
	if err != nil {
		t.Fatalf("ListQualifications() = %v, want nil", err)
	}
	if len(resp.GetQualifications()) != 1 {
		t.Errorf("ListQualifications() returned %d, want 1", len(resp.GetQualifications()))
	}
}

func TestListQualifications_FilterByWorkCenter(t *testing.T) {
	persons := newMockPersonRepo()
	personClasses := newMockPersonClassRepo()
	quals := newMockQualificationRepo()
	srv := NewQualificationServer(persons, personClasses, newMockWorkUnitRepo(), quals)

	person, _ := persons.CreatePerson(context.Background(), "pc-1", "EMP-001", "Jane", "Doe", nil)
	personClass, _ := personClasses.CreatePersonClass(context.Background(), "Operator", nil)

	_, _ = srv.QualifyPerson(context.Background(), &resourcev1.QualifyPersonRequest{
		PersonId: person.ID, PersonClassId: personClass.ID, WorkCenterId: "wc-1",
	})
	_, _ = srv.QualifyPerson(context.Background(), &resourcev1.QualifyPersonRequest{
		PersonId: person.ID, PersonClassId: personClass.ID, WorkCenterId: "wc-2",
	})

	resp, err := srv.ListQualifications(context.Background(), &resourcev1.ListQualificationsRequest{WorkCenterId: "wc-1"})
	if err != nil {
		t.Fatalf("ListQualifications() = %v, want nil", err)
	}
	if len(resp.GetQualifications()) != 1 {
		t.Errorf("ListQualifications() returned %d, want 1", len(resp.GetQualifications()))
	}
}

func TestListQualifications_Empty(t *testing.T) {
	srv := NewQualificationServer(newMockPersonRepo(), newMockPersonClassRepo(), newMockWorkUnitRepo(), newMockQualificationRepo())

	resp, err := srv.ListQualifications(context.Background(), &resourcev1.ListQualificationsRequest{})
	if err != nil {
		t.Fatalf("ListQualifications() = %v, want nil", err)
	}
	if len(resp.GetQualifications()) != 0 {
		t.Errorf("ListQualifications() returned %d, want 0", len(resp.GetQualifications()))
	}
}

// ============================================================================
// RevokeQualification
// ============================================================================

func TestRevokeQualification_Success(t *testing.T) {
	persons := newMockPersonRepo()
	personClasses := newMockPersonClassRepo()
	quals := newMockQualificationRepo()
	srv := NewQualificationServer(persons, personClasses, newMockWorkUnitRepo(), quals)

	person, _ := persons.CreatePerson(context.Background(), "pc-1", "EMP-001", "Jane", "Doe", nil)
	personClass, _ := personClasses.CreatePersonClass(context.Background(), "Operator", nil)

	createResp, _ := srv.QualifyPerson(context.Background(), &resourcev1.QualifyPersonRequest{
		PersonId: person.ID, PersonClassId: personClass.ID, WorkCenterId: "wc-1",
	})

	revokeResp, err := srv.RevokeQualification(context.Background(), &resourcev1.RevokeQualificationRequest{
		Id: createResp.GetQualification().GetId(),
	})
	if err != nil {
		t.Fatalf("RevokeQualification() = %v, want nil", err)
	}
	if !revokeResp.GetRevoked() {
		t.Error("Revoked = false, want true")
	}
}

func TestRevokeQualification_NotFound(t *testing.T) {
	srv := NewQualificationServer(newMockPersonRepo(), newMockPersonClassRepo(), newMockWorkUnitRepo(), newMockQualificationRepo())

	revokeResp, err := srv.RevokeQualification(context.Background(), &resourcev1.RevokeQualificationRequest{
		Id: "non-existent",
	})
	if err != nil {
		t.Fatalf("RevokeQualification() = %v, want nil", err)
	}
	if revokeResp.GetRevoked() {
		t.Error("Revoked = true, want false for non-existent ID")
	}
}

func TestRevokeQualification_MissingID(t *testing.T) {
	srv := NewQualificationServer(newMockPersonRepo(), newMockPersonClassRepo(), newMockWorkUnitRepo(), newMockQualificationRepo())

	_, err := srv.RevokeQualification(context.Background(), &resourcev1.RevokeQualificationRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("RevokeQualification() = %v, want InvalidArgument", err)
	}
}

// ============================================================================
// Error-path tests
// ============================================================================

// mockQualificationRepoErr returns errors on all methods.
type mockQualificationRepoErr struct{}

func (m *mockQualificationRepoErr) QualifyPerson(_ context.Context, _, _, _ string, _ *time.Time) (*db.QualificationRecord, error) {
	return nil, fmt.Errorf("simulated db error")
}

func (m *mockQualificationRepoErr) GetQualification(_ context.Context, _ string) (*db.QualificationRecord, error) {
	return nil, fmt.Errorf("simulated db error")
}

func (m *mockQualificationRepoErr) ListQualifications(_ context.Context, _, _ string) ([]*db.QualificationRecord, error) {
	return nil, fmt.Errorf("simulated db error")
}

func (m *mockQualificationRepoErr) RevokeQualification(_ context.Context, _ string) (bool, error) {
	return false, fmt.Errorf("simulated db error")
}

func TestQualifyPerson_RepoError(t *testing.T) {
	persons := newMockPersonRepo()
	personClasses := newMockPersonClassRepo()
	srv := NewQualificationServer(persons, personClasses, newMockWorkUnitRepo(), &mockQualificationRepoErr{})

	person, _ := persons.CreatePerson(context.Background(), "pc-1", "EMP-001", "Jane", "Doe", nil)
	personClass, _ := personClasses.CreatePersonClass(context.Background(), "Operator", nil)

	_, err := srv.QualifyPerson(context.Background(), &resourcev1.QualifyPersonRequest{
		PersonId: person.ID, PersonClassId: personClass.ID, WorkCenterId: "wc-1",
	})
	if status.Code(err) != codes.Internal {
		t.Errorf("QualifyPerson() = %v, want Internal", err)
	}
}

func TestGetQualification_RepoError(t *testing.T) {
	srv := NewQualificationServer(newMockPersonRepo(), newMockPersonClassRepo(), newMockWorkUnitRepo(), &mockQualificationRepoErr{})

	_, err := srv.GetQualification(context.Background(), &resourcev1.GetQualificationRequest{Id: "qr-1"})
	if status.Code(err) != codes.Internal {
		t.Errorf("GetQualification() = %v, want Internal", err)
	}
}

func TestListQualifications_RepoError(t *testing.T) {
	srv := NewQualificationServer(newMockPersonRepo(), newMockPersonClassRepo(), newMockWorkUnitRepo(), &mockQualificationRepoErr{})

	_, err := srv.ListQualifications(context.Background(), &resourcev1.ListQualificationsRequest{})
	if status.Code(err) != codes.Internal {
		t.Errorf("ListQualifications() = %v, want Internal", err)
	}
}

func TestRevokeQualification_RepoError(t *testing.T) {
	srv := NewQualificationServer(newMockPersonRepo(), newMockPersonClassRepo(), newMockWorkUnitRepo(), &mockQualificationRepoErr{})

	_, err := srv.RevokeQualification(context.Background(), &resourcev1.RevokeQualificationRequest{Id: "qr-1"})
	if status.Code(err) != codes.Internal {
		t.Errorf("RevokeQualification() = %v, want Internal", err)
	}
}
