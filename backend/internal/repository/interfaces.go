package repository

import (
	"context"
	"time"

	"github.com/visitas/backend/internal/models"
)

// PatientRepositoryInterface defines the interface for patient repository operations
type PatientRepositoryInterface interface {
	CheckStaffAccess(ctx context.Context, staffID, patientID string) (bool, error)
	GetPatientByID(ctx context.Context, patientID string) (*models.Patient, error)
}

// IdentifierRepositoryInterface defines the interface for identifier repository operations
type IdentifierRepositoryInterface interface {
	CreateIdentifier(ctx context.Context, req *models.PatientIdentifierCreateRequest, createdBy string) (*models.PatientIdentifier, error)
	GetIdentifierByID(ctx context.Context, identifierID string, decrypt bool) (*models.PatientIdentifier, error)
	GetIdentifiersByPatientID(ctx context.Context, patientID string, decrypt bool) ([]*models.PatientIdentifier, error)
	UpdateIdentifier(ctx context.Context, identifierID string, req *models.PatientIdentifierUpdateRequest, updatedBy string) (*models.PatientIdentifier, error)
	DeleteIdentifier(ctx context.Context, identifierID string) error
	GetPrimaryIdentifier(ctx context.Context, patientID string, identifierType models.IdentifierType, decrypt bool) (*models.PatientIdentifier, error)
}

// AuditRepositoryInterface defines the interface for audit repository operations
type AuditRepositoryInterface interface {
	LogMyNumberAccess(ctx context.Context, patientID, identifierID, action, userID string) error
	LogAccess(ctx context.Context, log *AuditLog) error
}

// ClinicalObservationRepositoryInterface defines the interface for clinical observation repository operations
type ClinicalObservationRepositoryInterface interface {
	Create(ctx context.Context, patientID string, req *models.ClinicalObservationCreateRequest, createdBy string) (*models.ClinicalObservation, error)
	GetByID(ctx context.Context, patientID, observationID string) (*models.ClinicalObservation, error)
	List(ctx context.Context, filter *models.ClinicalObservationFilter) ([]*models.ClinicalObservation, error)
	Update(ctx context.Context, patientID, observationID string, req *models.ClinicalObservationUpdateRequest, updatedBy string) (*models.ClinicalObservation, error)
	Delete(ctx context.Context, patientID, observationID string) error
	GetLatestByCategory(ctx context.Context, patientID, category string) (*models.ClinicalObservation, error)
	GetTimeSeriesData(ctx context.Context, patientID, category string, from, to time.Time) ([]*models.ClinicalObservation, error)
}

// MedicationOrderRepositoryInterface defines the interface for medication order repository operations
type MedicationOrderRepositoryInterface interface {
	Create(ctx context.Context, patientID string, req *models.MedicationOrderCreateRequest, createdBy string) (*models.MedicationOrder, error)
	GetByID(ctx context.Context, patientID, orderID string) (*models.MedicationOrder, error)
	List(ctx context.Context, filter *models.MedicationOrderFilter) ([]*models.MedicationOrder, error)
	Update(ctx context.Context, patientID, orderID string, req *models.MedicationOrderUpdateRequest, updatedBy string) (*models.MedicationOrder, error)
	UpdateWithVersion(ctx context.Context, patientID, orderID string, expectedVersion int64, req *models.MedicationOrderUpdateRequest, updatedBy string) (*models.MedicationOrder, error)
	Delete(ctx context.Context, patientID, orderID string) error
	GetActiveOrders(ctx context.Context, patientID string) ([]*models.MedicationOrder, error)
	GetOrdersByPrescription(ctx context.Context, patientID, prescribedBy string, prescribedDate time.Time) ([]*models.MedicationOrder, error)
}

// CarePlanRepositoryInterface defines the interface for care plan repository operations
type CarePlanRepositoryInterface interface {
	Create(ctx context.Context, patientID string, req *models.CarePlanCreateRequest) (*models.CarePlan, error)
	GetByID(ctx context.Context, patientID, planID string) (*models.CarePlan, error)
	List(ctx context.Context, filter *models.CarePlanFilter) ([]*models.CarePlan, error)
	Update(ctx context.Context, patientID, planID string, req *models.CarePlanUpdateRequest) (*models.CarePlan, error)
	UpdateWithVersion(ctx context.Context, patientID, planID string, expectedVersion int64, req *models.CarePlanUpdateRequest) (*models.CarePlan, error)
	Delete(ctx context.Context, patientID, planID string) error
	GetActiveCarePlans(ctx context.Context, patientID string) ([]*models.CarePlan, error)
}

// VisitScheduleRepositoryInterface defines the interface for visit schedule repository operations
type VisitScheduleRepositoryInterface interface {
	Create(ctx context.Context, patientID string, req *models.VisitScheduleCreateRequest) (*models.VisitSchedule, error)
	GetByID(ctx context.Context, patientID, scheduleID string) (*models.VisitSchedule, error)
	List(ctx context.Context, filter *models.VisitScheduleFilter) ([]*models.VisitSchedule, error)
	Update(ctx context.Context, patientID, scheduleID string, req *models.VisitScheduleUpdateRequest) (*models.VisitSchedule, error)
	Delete(ctx context.Context, patientID, scheduleID string) error
	GetUpcomingSchedules(ctx context.Context, patientID string, days int) ([]*models.VisitSchedule, error)
}

// ACPRecordRepositoryInterface defines the interface for ACP record repository operations
type ACPRecordRepositoryInterface interface {
	Create(ctx context.Context, patientID string, req *models.ACPRecordCreateRequest) (*models.ACPRecord, error)
	GetByID(ctx context.Context, patientID, acpID string) (*models.ACPRecord, error)
	List(ctx context.Context, filter *models.ACPRecordFilter) ([]*models.ACPRecord, error)
	Update(ctx context.Context, patientID, acpID string, req *models.ACPRecordUpdateRequest) (*models.ACPRecord, error)
	Delete(ctx context.Context, patientID, acpID string) error
	GetLatestACP(ctx context.Context, patientID string) (*models.ACPRecord, error)
	GetACPHistory(ctx context.Context, patientID string) ([]*models.ACPRecord, error)
}

// MedicalRecordRepositoryInterface defines the interface for medical record repository operations
type MedicalRecordRepositoryInterface interface {
	Create(ctx context.Context, patientID string, req *models.MedicalRecordCreateRequest) (*models.MedicalRecord, error)
	GetByID(ctx context.Context, patientID, recordID string) (*models.MedicalRecord, error)
	List(ctx context.Context, filter *models.MedicalRecordFilter) ([]*models.MedicalRecord, error)
	Update(ctx context.Context, patientID, recordID string, req *models.MedicalRecordUpdateRequest) (*models.MedicalRecord, error)
	Delete(ctx context.Context, patientID, recordID string) error
}

// MedicalRecordTemplateRepositoryInterface defines the interface for medical record template repository operations
type MedicalRecordTemplateRepositoryInterface interface {
	Create(ctx context.Context, req *models.MedicalRecordTemplateCreateRequest, createdBy string) (*models.MedicalRecordTemplate, error)
	GetByID(ctx context.Context, templateID string) (*models.MedicalRecordTemplate, error)
	List(ctx context.Context, filter *models.MedicalRecordTemplateFilter) ([]*models.MedicalRecordTemplate, error)
	Update(ctx context.Context, templateID string, req *models.MedicalRecordTemplateUpdateRequest, updatedBy string) (*models.MedicalRecordTemplate, error)
	Delete(ctx context.Context, templateID string) error
	IncrementUsageCount(ctx context.Context, templateID string) error
	GetSystemTemplates(ctx context.Context) ([]*models.MedicalRecordTemplate, error)
	GetBySpecialty(ctx context.Context, specialty string) ([]*models.MedicalRecordTemplate, error)
}
