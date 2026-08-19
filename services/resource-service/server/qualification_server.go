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

// QualificationServer implements Qualification Record gRPC RPCs.
type QualificationServer struct {
	resourcev1.UnimplementedEquipmentServiceServer
	persons       db.PersonRepository
	personClasses db.PersonClassRepository
	workUnits     db.WorkUnitRepository
	quals         db.QualificationRepository
}

// NewQualificationServer creates a new QualificationServer.
func NewQualificationServer(
	persons db.PersonRepository,
	personClasses db.PersonClassRepository,
	workUnits db.WorkUnitRepository,
	quals db.QualificationRepository,
) *QualificationServer {
	return &QualificationServer{
		persons:       persons,
		personClasses: personClasses,
		workUnits:     workUnits,
		quals:         quals,
	}
}

// QualifyPerson certifies a Person for a PersonClass at a Work Center.
func (s *QualificationServer) QualifyPerson(ctx context.Context, req *resourcev1.QualifyPersonRequest) (*resourcev1.QualifyPersonResponse, error) {
	if req.GetPersonId() == "" {
		return nil, status.Error(codes.InvalidArgument, "person_id is required")
	}
	if req.GetPersonClassId() == "" {
		return nil, status.Error(codes.InvalidArgument, "person_class_id is required")
	}
	if req.GetWorkCenterId() == "" {
		return nil, status.Error(codes.InvalidArgument, "work_center_id is required")
	}

	// Verify Person exists
	person, err := s.persons.GetPerson(ctx, req.GetPersonId())
	if err != nil {
		log.Printf("[QualificationServer][ERROR] GetPerson(%s): %v", req.GetPersonId(), err)
		return nil, status.Error(codes.Internal, "failed to verify person")
	}
	if person == nil {
		return nil, status.Error(codes.NotFound, "person not found")
	}

	// Verify PersonClass exists
	pc, err := s.personClasses.GetPersonClass(ctx, req.GetPersonClassId())
	if err != nil {
		log.Printf("[QualificationServer][ERROR] GetPersonClass(%s): %v", req.GetPersonClassId(), err)
		return nil, status.Error(codes.Internal, "failed to verify person class")
	}
	if pc == nil {
		return nil, status.Error(codes.NotFound, "person class not found")
	}

	// Parse optional expiry date
	var expiresAt *time.Time
	if req.GetExpiresAt() != "" {
		t, err := time.Parse(time.RFC3339, req.GetExpiresAt())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid expires_at format (use RFC3339): %v", err)
		}
		expiresAt = &t
	}

	qr, err := s.quals.QualifyPerson(ctx, req.GetPersonId(), req.GetPersonClassId(), req.GetWorkCenterId(), expiresAt)
	if err != nil {
		log.Printf("[QualificationServer][ERROR] QualifyPerson: %v", err)
		return nil, status.Error(codes.Internal, "failed to qualify person")
	}

	return &resourcev1.QualifyPersonResponse{
		Qualification: toProtoQualificationRecord(qr),
	}, nil
}

// GetQualification retrieves a Qualification Record by ID.
func (s *QualificationServer) GetQualification(ctx context.Context, req *resourcev1.GetQualificationRequest) (*resourcev1.GetQualificationResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	qr, err := s.quals.GetQualification(ctx, req.GetId())
	if err != nil {
		log.Printf("[QualificationServer][ERROR] GetQualification(%s): %v", req.GetId(), err)
		return nil, status.Error(codes.Internal, "failed to get qualification")
	}
	if qr == nil {
		return nil, status.Error(codes.NotFound, "qualification not found")
	}

	return &resourcev1.GetQualificationResponse{
		Qualification: toProtoQualificationRecord(qr),
	}, nil
}

// ListQualifications returns Qualification Records, optionally filtered.
func (s *QualificationServer) ListQualifications(ctx context.Context, req *resourcev1.ListQualificationsRequest) (*resourcev1.ListQualificationsResponse, error) {
	records, err := s.quals.ListQualifications(ctx, req.GetPersonId(), req.GetWorkCenterId())
	if err != nil {
		log.Printf("[QualificationServer][ERROR] ListQualifications: %v", err)
		return nil, status.Error(codes.Internal, "failed to list qualifications")
	}

	protoRecords := make([]*resourcev1.QualificationRecord, len(records))
	for i, qr := range records {
		protoRecords[i] = toProtoQualificationRecord(qr)
	}

	return &resourcev1.ListQualificationsResponse{
		Qualifications: protoRecords,
	}, nil
}

// RevokeQualification deletes a Qualification Record by ID.
func (s *QualificationServer) RevokeQualification(ctx context.Context, req *resourcev1.RevokeQualificationRequest) (*resourcev1.RevokeQualificationResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	revoked, err := s.quals.RevokeQualification(ctx, req.GetId())
	if err != nil {
		log.Printf("[QualificationServer][ERROR] RevokeQualification(%s): %v", req.GetId(), err)
		return nil, status.Error(codes.Internal, "failed to revoke qualification")
	}

	return &resourcev1.RevokeQualificationResponse{
		Revoked: revoked,
	}, nil
}
