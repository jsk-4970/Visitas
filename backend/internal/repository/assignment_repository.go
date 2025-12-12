package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"cloud.google.com/go/spanner"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

// StaffRole represents the role of a staff member
type StaffRole string

const (
	StaffRoleDoctor      StaffRole = "doctor"
	StaffRoleNurse       StaffRole = "nurse"
	StaffRoleCareManager StaffRole = "care_manager"
)

// AssignmentType represents the type of assignment
type AssignmentType string

const (
	AssignmentTypePrimary AssignmentType = "primary"
	AssignmentTypeBackup  AssignmentType = "backup"
)

// AssignmentStatus represents the status of an assignment
type AssignmentStatus string

const (
	AssignmentStatusActive   AssignmentStatus = "active"
	AssignmentStatusInactive AssignmentStatus = "inactive"
)

// StaffPatientAssignment represents a staff-patient assignment
type StaffPatientAssignment struct {
	AssignmentID   string           `json:"assignment_id"`
	StaffID        string           `json:"staff_id"`
	PatientID      string           `json:"patient_id"`
	Role           StaffRole        `json:"role"`
	AssignmentType AssignmentType   `json:"assignment_type"`
	Status         AssignmentStatus `json:"status"`
	AssignedAt     time.Time        `json:"assigned_at"`
	AssignedBy     string           `json:"assigned_by"`
	InactivatedAt  sql.NullTime     `json:"inactivated_at,omitempty"`
	InactivatedBy  string           `json:"inactivated_by,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

// AssignmentRepository handles staff-patient assignment operations
type AssignmentRepository struct {
	client *spanner.Client
}

// NewAssignmentRepository creates a new assignment repository
func NewAssignmentRepository(spannerRepo *SpannerRepository) *AssignmentRepository {
	return &AssignmentRepository{
		client: spannerRepo.client,
	}
}

// CreateAssignment creates a new staff-patient assignment
func (r *AssignmentRepository) CreateAssignment(ctx context.Context, staffID, patientID string, role StaffRole, assignmentType AssignmentType, assignedBy string) (*StaffPatientAssignment, error) {
	assignmentID := uuid.New().String()
	now := time.Now()

	mutation := spanner.InsertMap("staff_patient_assignments", map[string]interface{}{
		"assignment_id":   assignmentID,
		"staff_id":        staffID,
		"patient_id":      patientID,
		"role":            string(role),
		"assignment_type": string(assignmentType),
		"status":          string(AssignmentStatusActive),
		"assigned_at":     now,
		"assigned_by":     assignedBy,
		"inactivated_at":  nil,
		"inactivated_by":  "",
		"created_at":      now,
		"updated_at":      now,
	})

	_, err := r.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return nil, fmt.Errorf("failed to create assignment: %w", err)
	}

	assignment, err := r.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created assignment: %w", err)
	}

	return assignment, nil
}

// GetAssignmentByID retrieves an assignment by ID
func (r *AssignmentRepository) GetAssignmentByID(ctx context.Context, assignmentID string) (*StaffPatientAssignment, error) {
	stmt := NewStatement(`SELECT
			assignment_id, staff_id, patient_id, role, assignment_type,
			status, assigned_at, assigned_by, inactivated_at, COALESCE(inactivated_by, ''),
			created_at, updated_at
		FROM staff_patient_assignments
		WHERE assignment_id = @assignmentID`,
		map[string]interface{}{
			"assignmentID": assignmentID,
		})

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	row, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("assignment not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query assignment: %w", err)
	}

	assignment, err := scanAssignment(row)
	if err != nil {
		return nil, fmt.Errorf("failed to scan assignment: %w", err)
	}

	return assignment, nil
}

// GetAssignmentsByStaffID retrieves all assignments for a staff member
func (r *AssignmentRepository) GetAssignmentsByStaffID(ctx context.Context, staffID string, activeOnly bool) ([]*StaffPatientAssignment, error) {
	sqlQuery := `SELECT
		assignment_id, staff_id, patient_id, role, assignment_type,
		status, assigned_at, assigned_by, inactivated_at, COALESCE(inactivated_by, ''),
		created_at, updated_at
	FROM staff_patient_assignments
	WHERE staff_id = @staffID`

	if activeOnly {
		sqlQuery += ` AND status = 'active'`
	}

	sqlQuery += ` ORDER BY assigned_at DESC`

	stmt := NewStatement(sqlQuery,
		map[string]interface{}{
			"staffID": staffID,
		})

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var assignments []*StaffPatientAssignment
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate assignments: %w", err)
		}

		assignment, err := scanAssignment(row)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment: %w", err)
		}

		assignments = append(assignments, assignment)
	}

	return assignments, nil
}

// GetAssignmentsByPatientID retrieves all assignments for a patient
func (r *AssignmentRepository) GetAssignmentsByPatientID(ctx context.Context, patientID string, activeOnly bool) ([]*StaffPatientAssignment, error) {
	sqlQuery := `SELECT
		assignment_id, staff_id, patient_id, role, assignment_type,
		status, assigned_at, assigned_by, inactivated_at, COALESCE(inactivated_by, ''),
		created_at, updated_at
	FROM staff_patient_assignments
	WHERE patient_id = @patientID`

	if activeOnly {
		sqlQuery += ` AND status = 'active'`
	}

	sqlQuery += ` ORDER BY assignment_type ASC, assigned_at DESC`

	stmt := NewStatement(sqlQuery,
		map[string]interface{}{
			"patientID": patientID,
		})

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	var assignments []*StaffPatientAssignment
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate assignments: %w", err)
		}

		assignment, err := scanAssignment(row)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment: %w", err)
		}

		assignments = append(assignments, assignment)
	}

	return assignments, nil
}

// InactivateAssignment sets an assignment to inactive
func (r *AssignmentRepository) InactivateAssignment(ctx context.Context, assignmentID, inactivatedBy string) error {
	now := time.Now()

	mutation := spanner.UpdateMap("staff_patient_assignments", map[string]interface{}{
		"assignment_id":  assignmentID,
		"status":         string(AssignmentStatusInactive),
		"inactivated_at": now,
		"inactivated_by": inactivatedBy,
		"updated_at":     now,
	})

	_, err := r.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return fmt.Errorf("failed to inactivate assignment: %w", err)
	}

	return nil
}

// ReactivateAssignment sets an assignment back to active
func (r *AssignmentRepository) ReactivateAssignment(ctx context.Context, assignmentID string) error {
	now := time.Now()

	mutation := spanner.UpdateMap("staff_patient_assignments", map[string]interface{}{
		"assignment_id":  assignmentID,
		"status":         string(AssignmentStatusActive),
		"inactivated_at": nil,
		"inactivated_by": "",
		"updated_at":     now,
	})

	_, err := r.client.Apply(ctx, []*spanner.Mutation{mutation})
	if err != nil {
		return fmt.Errorf("failed to reactivate assignment: %w", err)
	}

	return nil
}

// CheckAssignment verifies if a staff member is assigned to a patient
func (r *AssignmentRepository) CheckAssignment(ctx context.Context, staffID, patientID string) (bool, error) {
	stmt := NewStatement(`SELECT COUNT(*) as count
		FROM staff_patient_assignments
		WHERE staff_id = @staffID
			AND patient_id = @patientID
			AND status = 'active'`,
		map[string]interface{}{
			"staffID":   staffID,
			"patientID": patientID,
		})

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	row, err := iter.Next()
	if err != nil {
		return false, fmt.Errorf("failed to check assignment: %w", err)
	}

	var count int64
	if err := row.Columns(&count); err != nil {
		return false, fmt.Errorf("failed to scan count: %w", err)
	}

	return count > 0, nil
}

// GetPrimaryAssignment retrieves the primary assignment for a patient by role
func (r *AssignmentRepository) GetPrimaryAssignment(ctx context.Context, patientID string, role StaffRole) (*StaffPatientAssignment, error) {
	stmt := NewStatement(`SELECT
			assignment_id, staff_id, patient_id, role, assignment_type,
			status, assigned_at, assigned_by, inactivated_at, COALESCE(inactivated_by, ''),
			created_at, updated_at
		FROM staff_patient_assignments
		WHERE patient_id = @patientID
			AND role = @role
			AND assignment_type = 'primary'
			AND status = 'active'
		LIMIT 1`,
		map[string]interface{}{
			"patientID": patientID,
			"role":      string(role),
		})

	iter := r.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	row, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("primary assignment not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query primary assignment: %w", err)
	}

	assignment, err := scanAssignment(row)
	if err != nil {
		return nil, fmt.Errorf("failed to scan assignment: %w", err)
	}

	return assignment, nil
}

// scanAssignment scans a Spanner row into a StaffPatientAssignment model
func scanAssignment(row *spanner.Row) (*StaffPatientAssignment, error) {
	var assignment StaffPatientAssignment
	var roleStr, assignmentTypeStr, statusStr string

	err := row.Columns(
		&assignment.AssignmentID,
		&assignment.StaffID,
		&assignment.PatientID,
		&roleStr,
		&assignmentTypeStr,
		&statusStr,
		&assignment.AssignedAt,
		&assignment.AssignedBy,
		&assignment.InactivatedAt,
		&assignment.InactivatedBy,
		&assignment.CreatedAt,
		&assignment.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	assignment.Role = StaffRole(roleStr)
	assignment.AssignmentType = AssignmentType(assignmentTypeStr)
	assignment.Status = AssignmentStatus(statusStr)

	return &assignment, nil
}
