// Package server implements the gRPC handlers for the Resource Service.
//
// equipment_service.go provides a combined EquipmentServiceServer adapter
// that delegates to the focused per-domain server structs. This is needed
// because the proto defines a single EquipmentService containing all RPCs,
// while the implementation is split into WorkUnitServer, EquipmentClassServer,
// CapabilityServer, PhysicalAssetServer, InstallationServer, PersonClassServer,
// PersonServer, and QualificationServer.
package server

import (
	"context"

	resourcev1 "github.com/iamraydoan/factoryos/platform/platform-sdk/go/gen/resource/v1"
	"github.com/iamraydoan/factoryos/services/resource-service/db"
)

// EquipmentService implements resourcev1.EquipmentServiceServer by delegating
// to focused per-domain server structs. Each domain server owns only the
// repository interfaces it needs.
type EquipmentService struct {
	resourcev1.UnimplementedEquipmentServiceServer

	WorkUnits       *WorkUnitServer
	Classes         *EquipmentClassServer
	Capabilities    *CapabilityServer
	Assets          *PhysicalAssetServer
	Installations   *InstallationServer
	PersonClasses   *PersonClassServer
	Persons         *PersonServer
	Qualifications  *QualificationServer
	Shifts          *ShiftServer
	Assignments     *ShiftAssignmentServer
	MaterialClasses *MaterialClassServer
	MaterialDefs    *MaterialDefinitionServer
}

// NewEquipmentService creates a combined EquipmentService from a single
// PostgresEquipmentRepository that implements all domain interfaces.
func NewEquipmentService(repo *db.PostgresEquipmentRepository) *EquipmentService {
	return &EquipmentService{
		WorkUnits:       NewWorkUnitServer(repo),
		Classes:         NewEquipmentClassServer(repo),
		Capabilities:    NewCapabilityServer(repo, repo, repo),
		Assets:          NewPhysicalAssetServer(repo),
		Installations:   NewInstallationServer(repo, repo, repo),
		PersonClasses:   NewPersonClassServer(repo),
		Persons:         NewPersonServer(repo),
		Qualifications:  NewQualificationServer(repo, repo, repo, repo),
		Shifts:          NewShiftServer(repo),
		Assignments:     NewShiftAssignmentServer(repo, repo, repo),
		MaterialClasses: NewMaterialClassServer(repo),
		MaterialDefs:    NewMaterialDefinitionServer(repo, repo),
	}
}

// --- Work Unit delegation ---

func (s *EquipmentService) CreateWorkUnit(ctx context.Context, req *resourcev1.CreateWorkUnitRequest) (*resourcev1.CreateWorkUnitResponse, error) {
	return s.WorkUnits.CreateWorkUnit(ctx, req)
}

func (s *EquipmentService) GetWorkUnit(ctx context.Context, req *resourcev1.GetWorkUnitRequest) (*resourcev1.GetWorkUnitResponse, error) {
	return s.WorkUnits.GetWorkUnit(ctx, req)
}

func (s *EquipmentService) ListWorkUnits(ctx context.Context, req *resourcev1.ListWorkUnitsRequest) (*resourcev1.ListWorkUnitsResponse, error) {
	return s.WorkUnits.ListWorkUnits(ctx, req)
}

func (s *EquipmentService) UpdateWorkUnitStatus(ctx context.Context, req *resourcev1.UpdateWorkUnitStatusRequest) (*resourcev1.UpdateWorkUnitStatusResponse, error) {
	return s.WorkUnits.UpdateWorkUnitStatus(ctx, req)
}

// --- Equipment Class delegation ---

func (s *EquipmentService) CreateEquipmentClass(ctx context.Context, req *resourcev1.CreateEquipmentClassRequest) (*resourcev1.CreateEquipmentClassResponse, error) {
	return s.Classes.CreateEquipmentClass(ctx, req)
}

func (s *EquipmentService) GetEquipmentClass(ctx context.Context, req *resourcev1.GetEquipmentClassRequest) (*resourcev1.GetEquipmentClassResponse, error) {
	return s.Classes.GetEquipmentClass(ctx, req)
}

func (s *EquipmentService) ListEquipmentClasses(ctx context.Context, req *resourcev1.ListEquipmentClassesRequest) (*resourcev1.ListEquipmentClassesResponse, error) {
	return s.Classes.ListEquipmentClasses(ctx, req)
}

// --- Capability delegation ---

func (s *EquipmentService) AssignCapability(ctx context.Context, req *resourcev1.AssignCapabilityRequest) (*resourcev1.AssignCapabilityResponse, error) {
	return s.Capabilities.AssignCapability(ctx, req)
}

func (s *EquipmentService) ListWorkUnitCapabilities(ctx context.Context, req *resourcev1.ListWorkUnitCapabilitiesRequest) (*resourcev1.ListWorkUnitCapabilitiesResponse, error) {
	return s.Capabilities.ListWorkUnitCapabilities(ctx, req)
}

func (s *EquipmentService) RemoveCapability(ctx context.Context, req *resourcev1.RemoveCapabilityRequest) (*resourcev1.RemoveCapabilityResponse, error) {
	return s.Capabilities.RemoveCapability(ctx, req)
}

// --- Physical Asset delegation ---

func (s *EquipmentService) CreatePhysicalAsset(ctx context.Context, req *resourcev1.CreatePhysicalAssetRequest) (*resourcev1.CreatePhysicalAssetResponse, error) {
	return s.Assets.CreatePhysicalAsset(ctx, req)
}

func (s *EquipmentService) GetPhysicalAsset(ctx context.Context, req *resourcev1.GetPhysicalAssetRequest) (*resourcev1.GetPhysicalAssetResponse, error) {
	return s.Assets.GetPhysicalAsset(ctx, req)
}

func (s *EquipmentService) ListPhysicalAssets(ctx context.Context, req *resourcev1.ListPhysicalAssetsRequest) (*resourcev1.ListPhysicalAssetsResponse, error) {
	return s.Assets.ListPhysicalAssets(ctx, req)
}

// --- Installation delegation ---

func (s *EquipmentService) InstallAsset(ctx context.Context, req *resourcev1.InstallAssetRequest) (*resourcev1.InstallAssetResponse, error) {
	return s.Installations.InstallAsset(ctx, req)
}

func (s *EquipmentService) UninstallAsset(ctx context.Context, req *resourcev1.UninstallAssetRequest) (*resourcev1.UninstallAssetResponse, error) {
	return s.Installations.UninstallAsset(ctx, req)
}

func (s *EquipmentService) GetCurrentInstallation(ctx context.Context, req *resourcev1.GetCurrentInstallationRequest) (*resourcev1.GetCurrentInstallationResponse, error) {
	return s.Installations.GetCurrentInstallation(ctx, req)
}

func (s *EquipmentService) ListInstallations(ctx context.Context, req *resourcev1.ListInstallationsRequest) (*resourcev1.ListInstallationsResponse, error) {
	return s.Installations.ListInstallations(ctx, req)
}

// --- Person Class delegation ---

func (s *EquipmentService) CreatePersonClass(ctx context.Context, req *resourcev1.CreatePersonClassRequest) (*resourcev1.CreatePersonClassResponse, error) {
	return s.PersonClasses.CreatePersonClass(ctx, req)
}

func (s *EquipmentService) GetPersonClass(ctx context.Context, req *resourcev1.GetPersonClassRequest) (*resourcev1.GetPersonClassResponse, error) {
	return s.PersonClasses.GetPersonClass(ctx, req)
}

func (s *EquipmentService) ListPersonClasses(ctx context.Context, req *resourcev1.ListPersonClassesRequest) (*resourcev1.ListPersonClassesResponse, error) {
	return s.PersonClasses.ListPersonClasses(ctx, req)
}

// --- Person delegation ---

func (s *EquipmentService) CreatePerson(ctx context.Context, req *resourcev1.CreatePersonRequest) (*resourcev1.CreatePersonResponse, error) {
	return s.Persons.CreatePerson(ctx, req)
}

func (s *EquipmentService) GetPerson(ctx context.Context, req *resourcev1.GetPersonRequest) (*resourcev1.GetPersonResponse, error) {
	return s.Persons.GetPerson(ctx, req)
}

func (s *EquipmentService) ListPersons(ctx context.Context, req *resourcev1.ListPersonsRequest) (*resourcev1.ListPersonsResponse, error) {
	return s.Persons.ListPersons(ctx, req)
}

// --- Qualification Record delegation ---

func (s *EquipmentService) QualifyPerson(ctx context.Context, req *resourcev1.QualifyPersonRequest) (*resourcev1.QualifyPersonResponse, error) {
	return s.Qualifications.QualifyPerson(ctx, req)
}

func (s *EquipmentService) GetQualification(ctx context.Context, req *resourcev1.GetQualificationRequest) (*resourcev1.GetQualificationResponse, error) {
	return s.Qualifications.GetQualification(ctx, req)
}

func (s *EquipmentService) ListQualifications(ctx context.Context, req *resourcev1.ListQualificationsRequest) (*resourcev1.ListQualificationsResponse, error) {
	return s.Qualifications.ListQualifications(ctx, req)
}

func (s *EquipmentService) RevokeQualification(ctx context.Context, req *resourcev1.RevokeQualificationRequest) (*resourcev1.RevokeQualificationResponse, error) {
	return s.Qualifications.RevokeQualification(ctx, req)
}

func (s *EquipmentService) CheckExpiringQualifications(ctx context.Context, req *resourcev1.CheckExpiringQualificationsRequest) (*resourcev1.CheckExpiringQualificationsResponse, error) {
	return s.Qualifications.CheckExpiringQualifications(ctx, req)
}

// --- Shift delegation ---

func (s *EquipmentService) CreateShift(ctx context.Context, req *resourcev1.CreateShiftRequest) (*resourcev1.CreateShiftResponse, error) {
	return s.Shifts.CreateShift(ctx, req)
}

func (s *EquipmentService) GetShift(ctx context.Context, req *resourcev1.GetShiftRequest) (*resourcev1.GetShiftResponse, error) {
	return s.Shifts.GetShift(ctx, req)
}

func (s *EquipmentService) ListShifts(ctx context.Context, req *resourcev1.ListShiftsRequest) (*resourcev1.ListShiftsResponse, error) {
	return s.Shifts.ListShifts(ctx, req)
}

// --- Shift Assignment delegation ---

func (s *EquipmentService) AssignShift(ctx context.Context, req *resourcev1.AssignShiftRequest) (*resourcev1.AssignShiftResponse, error) {
	return s.Assignments.AssignShift(ctx, req)
}

func (s *EquipmentService) GetShiftAssignment(ctx context.Context, req *resourcev1.GetShiftAssignmentRequest) (*resourcev1.GetShiftAssignmentResponse, error) {
	return s.Assignments.GetShiftAssignment(ctx, req)
}

func (s *EquipmentService) ListShiftAssignments(ctx context.Context, req *resourcev1.ListShiftAssignmentsRequest) (*resourcev1.ListShiftAssignmentsResponse, error) {
	return s.Assignments.ListShiftAssignments(ctx, req)
}

func (s *EquipmentService) UnassignShift(ctx context.Context, req *resourcev1.UnassignShiftRequest) (*resourcev1.UnassignShiftResponse, error) {
	return s.Assignments.UnassignShift(ctx, req)
}

// --- Material Class delegation ---

func (s *EquipmentService) CreateMaterialClass(ctx context.Context, req *resourcev1.CreateMaterialClassRequest) (*resourcev1.CreateMaterialClassResponse, error) {
	return s.MaterialClasses.CreateMaterialClass(ctx, req)
}

func (s *EquipmentService) GetMaterialClass(ctx context.Context, req *resourcev1.GetMaterialClassRequest) (*resourcev1.GetMaterialClassResponse, error) {
	return s.MaterialClasses.GetMaterialClass(ctx, req)
}

func (s *EquipmentService) ListMaterialClasses(ctx context.Context, req *resourcev1.ListMaterialClassesRequest) (*resourcev1.ListMaterialClassesResponse, error) {
	return s.MaterialClasses.ListMaterialClasses(ctx, req)
}

// --- Material Definition delegation ---

func (s *EquipmentService) CreateMaterialDefinition(ctx context.Context, req *resourcev1.CreateMaterialDefinitionRequest) (*resourcev1.CreateMaterialDefinitionResponse, error) {
	return s.MaterialDefs.CreateMaterialDefinition(ctx, req)
}

func (s *EquipmentService) GetMaterialDefinition(ctx context.Context, req *resourcev1.GetMaterialDefinitionRequest) (*resourcev1.GetMaterialDefinitionResponse, error) {
	return s.MaterialDefs.GetMaterialDefinition(ctx, req)
}

func (s *EquipmentService) ListMaterialDefinitions(ctx context.Context, req *resourcev1.ListMaterialDefinitionsRequest) (*resourcev1.ListMaterialDefinitionsResponse, error) {
	return s.MaterialDefs.ListMaterialDefinitions(ctx, req)
}
