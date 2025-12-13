package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/visitas/backend/internal/models"
)

func TestVisitSchedule_Integration_CreateAndGet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Test: Create a visit schedule
	t.Run("Create visit schedule", func(t *testing.T) {
		visitDate := time.Now().AddDate(0, 0, 1).Format(time.RFC3339)
		scheduleJSON := fmt.Sprintf(`{
			"visit_date": "%s",
			"visit_type": "regular",
			"estimated_duration_minutes": 60,
			"status": "draft",
			"priority_score": 5
		}`, visitDate)

		resp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/schedules", patientID), strings.NewReader(scheduleJSON))
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var schedule models.VisitSchedule
		DecodeJSONResponse(t, resp, &schedule)

		assert.NotEmpty(t, schedule.ScheduleID)
		assert.Equal(t, patientID, schedule.PatientID)
		assert.Equal(t, "regular", schedule.VisitType)
		assert.Equal(t, int64(60), schedule.EstimatedDurationMinutes)
		assert.Equal(t, "draft", schedule.Status)
		assert.Equal(t, int64(5), schedule.PriorityScore)

		// Test: Get the created schedule
		t.Run("Get visit schedule by ID", func(t *testing.T) {
			getResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/patients/%s/schedules/%s", patientID, schedule.ScheduleID), nil)
			assert.Equal(t, http.StatusOK, getResp.StatusCode)

			var retrievedSchedule models.VisitSchedule
			DecodeJSONResponse(t, getResp, &retrievedSchedule)

			assert.Equal(t, schedule.ScheduleID, retrievedSchedule.ScheduleID)
			assert.Equal(t, schedule.PatientID, retrievedSchedule.PatientID)
			assert.Equal(t, schedule.VisitType, retrievedSchedule.VisitType)
		})
	})
}

func TestVisitSchedule_Integration_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create a visit schedule
	visitDate := time.Now().AddDate(0, 0, 1).Format(time.RFC3339)
	scheduleJSON := fmt.Sprintf(`{
		"visit_date": "%s",
		"visit_type": "regular",
		"estimated_duration_minutes": 60,
		"status": "draft",
		"priority_score": 5
	}`, visitDate)

	createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/schedules", patientID), strings.NewReader(scheduleJSON))
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var schedule models.VisitSchedule
	DecodeJSONResponse(t, createResp, &schedule)

	// Test: Update the schedule
	t.Run("Update visit schedule", func(t *testing.T) {
		updateJSON := `{
			"visit_type": "emergency",
			"estimated_duration_minutes": 90,
			"status": "assigned",
			"priority_score": 10
		}`

		updateResp := ts.MakeRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/patients/%s/schedules/%s", patientID, schedule.ScheduleID), strings.NewReader(updateJSON))
		assert.Equal(t, http.StatusOK, updateResp.StatusCode)

		var updatedSchedule models.VisitSchedule
		DecodeJSONResponse(t, updateResp, &updatedSchedule)

		assert.Equal(t, schedule.ScheduleID, updatedSchedule.ScheduleID)
		assert.Equal(t, "emergency", updatedSchedule.VisitType)
		assert.Equal(t, int64(90), updatedSchedule.EstimatedDurationMinutes)
		assert.Equal(t, "assigned", updatedSchedule.Status)
		assert.Equal(t, int64(10), updatedSchedule.PriorityScore)
	})
}

func TestVisitSchedule_Integration_List(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create multiple visit schedules
	scheduleIDs := make([]string, 0)
	for i := 0; i < 3; i++ {
		visitDate := time.Now().AddDate(0, 0, i+1).Format(time.RFC3339)
		scheduleJSON := fmt.Sprintf(`{
			"visit_date": "%s",
			"visit_type": "regular",
			"estimated_duration_minutes": 60,
			"status": "draft",
			"priority_score": %d
		}`, visitDate, i+1)

		createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/schedules", patientID), strings.NewReader(scheduleJSON))
		require.Equal(t, http.StatusCreated, createResp.StatusCode)

		var schedule models.VisitSchedule
		DecodeJSONResponse(t, createResp, &schedule)
		scheduleIDs = append(scheduleIDs, schedule.ScheduleID)
	}

	// Test: List visit schedules
	t.Run("List visit schedules", func(t *testing.T) {
		listResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/patients/%s/schedules", patientID), nil)
		assert.Equal(t, http.StatusOK, listResp.StatusCode)

		var schedules []models.VisitSchedule
		DecodeJSONResponse(t, listResp, &schedules)

		assert.GreaterOrEqual(t, len(schedules), 3)

		// Verify all created schedules are in the list
		scheduleIDSet := make(map[string]bool)
		for _, schedule := range schedules {
			scheduleIDSet[schedule.ScheduleID] = true
		}

		for _, scheduleID := range scheduleIDs {
			assert.True(t, scheduleIDSet[scheduleID], "Schedule %s should be in the list", scheduleID)
		}
	})
}

func TestVisitSchedule_Integration_GetUpcoming(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create a future schedule
	visitDate := time.Now().AddDate(0, 0, 1).Format(time.RFC3339)
	scheduleJSON := fmt.Sprintf(`{
		"visit_date": "%s",
		"visit_type": "regular",
		"estimated_duration_minutes": 60,
		"status": "draft",
		"priority_score": 5
	}`, visitDate)

	createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/schedules", patientID), strings.NewReader(scheduleJSON))
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var schedule models.VisitSchedule
	DecodeJSONResponse(t, createResp, &schedule)

	// Test: Get upcoming schedules
	t.Run("Get upcoming schedules", func(t *testing.T) {
		upcomingResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/patients/%s/schedules/upcoming", patientID), nil)
		assert.Equal(t, http.StatusOK, upcomingResp.StatusCode)

		var upcomingSchedules []models.VisitSchedule
		DecodeJSONResponse(t, upcomingResp, &upcomingSchedules)

		assert.GreaterOrEqual(t, len(upcomingSchedules), 1)

		// Verify the created schedule is in the upcoming list
		foundSchedule := false
		for _, s := range upcomingSchedules {
			if s.ScheduleID == schedule.ScheduleID {
				foundSchedule = true
				break
			}
		}
		assert.True(t, foundSchedule, "Created schedule should be in upcoming schedules")
	})
}

func TestVisitSchedule_Integration_AssignStaff(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create a visit schedule
	visitDate := time.Now().AddDate(0, 0, 1).Format(time.RFC3339)
	scheduleJSON := fmt.Sprintf(`{
		"visit_date": "%s",
		"visit_type": "regular",
		"estimated_duration_minutes": 60,
		"status": "draft",
		"priority_score": 5
	}`, visitDate)

	createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/schedules", patientID), strings.NewReader(scheduleJSON))
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var schedule models.VisitSchedule
	DecodeJSONResponse(t, createResp, &schedule)

	// Test: Assign staff to schedule
	t.Run("Assign staff to schedule", func(t *testing.T) {
		assignJSON := `{
			"staff_id": "staff-123"
		}`

		assignResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/schedules/%s/assign-staff", patientID, schedule.ScheduleID), strings.NewReader(assignJSON))
		assert.Equal(t, http.StatusOK, assignResp.StatusCode)

		var updatedSchedule models.VisitSchedule
		DecodeJSONResponse(t, assignResp, &updatedSchedule)

		assert.Equal(t, schedule.ScheduleID, updatedSchedule.ScheduleID)
		assert.True(t, updatedSchedule.AssignedStaffID.Valid)
		assert.Equal(t, "staff-123", updatedSchedule.AssignedStaffID.StringVal)
	})
}

func TestVisitSchedule_Integration_UpdateStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create a visit schedule
	visitDate := time.Now().AddDate(0, 0, 1).Format(time.RFC3339)
	scheduleJSON := fmt.Sprintf(`{
		"visit_date": "%s",
		"visit_type": "regular",
		"estimated_duration_minutes": 60,
		"status": "draft",
		"priority_score": 5
	}`, visitDate)

	createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/schedules", patientID), strings.NewReader(scheduleJSON))
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var schedule models.VisitSchedule
	DecodeJSONResponse(t, createResp, &schedule)

	// Test: Update schedule status
	t.Run("Update schedule status", func(t *testing.T) {
		statusJSON := `{
			"status": "in_progress"
		}`

		statusResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/schedules/%s/status", patientID, schedule.ScheduleID), strings.NewReader(statusJSON))
		assert.Equal(t, http.StatusOK, statusResp.StatusCode)

		var updatedSchedule models.VisitSchedule
		DecodeJSONResponse(t, statusResp, &updatedSchedule)

		assert.Equal(t, schedule.ScheduleID, updatedSchedule.ScheduleID)
		assert.Equal(t, "in_progress", updatedSchedule.Status)
	})
}

func TestVisitSchedule_Integration_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create a visit schedule
	visitDate := time.Now().AddDate(0, 0, 1).Format(time.RFC3339)
	scheduleJSON := fmt.Sprintf(`{
		"visit_date": "%s",
		"visit_type": "regular",
		"estimated_duration_minutes": 60,
		"status": "draft",
		"priority_score": 5
	}`, visitDate)

	createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/schedules", patientID), strings.NewReader(scheduleJSON))
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var schedule models.VisitSchedule
	DecodeJSONResponse(t, createResp, &schedule)

	// Test: Delete the schedule
	t.Run("Delete visit schedule", func(t *testing.T) {
		deleteResp := ts.MakeRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/patients/%s/schedules/%s", patientID, schedule.ScheduleID), nil)
		assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

		// Verify the schedule is deleted (should return 404)
		getResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/patients/%s/schedules/%s", patientID, schedule.ScheduleID), nil)
		assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
	})
}

func TestVisitSchedule_Integration_ValidationErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "Missing visit_date",
			requestBody:    `{"visit_type": "regular", "estimated_duration_minutes": 60, "status": "draft", "priority_score": 5}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing visit_type",
			requestBody:    fmt.Sprintf(`{"visit_date": "%s", "estimated_duration_minutes": 60, "status": "draft", "priority_score": 5}`, time.Now().AddDate(0, 0, 1).Format(time.RFC3339)),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid visit_type",
			requestBody:    fmt.Sprintf(`{"visit_date": "%s", "visit_type": "invalid", "estimated_duration_minutes": 60, "status": "draft", "priority_score": 5}`, time.Now().AddDate(0, 0, 1).Format(time.RFC3339)),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid status",
			requestBody:    fmt.Sprintf(`{"visit_date": "%s", "visit_type": "regular", "estimated_duration_minutes": 60, "status": "invalid", "priority_score": 5}`, time.Now().AddDate(0, 0, 1).Format(time.RFC3339)),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid estimated_duration_minutes (too low)",
			requestBody:    fmt.Sprintf(`{"visit_date": "%s", "visit_type": "regular", "estimated_duration_minutes": 0, "status": "draft", "priority_score": 5}`, time.Now().AddDate(0, 0, 1).Format(time.RFC3339)),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid estimated_duration_minutes (too high)",
			requestBody:    fmt.Sprintf(`{"visit_date": "%s", "visit_type": "regular", "estimated_duration_minutes": 500, "status": "draft", "priority_score": 5}`, time.Now().AddDate(0, 0, 1).Format(time.RFC3339)),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid priority_score (too high)",
			requestBody:    fmt.Sprintf(`{"visit_date": "%s", "visit_type": "regular", "estimated_duration_minutes": 60, "status": "draft", "priority_score": 11}`, time.Now().AddDate(0, 0, 1).Format(time.RFC3339)),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/schedules", patientID), strings.NewReader(tt.requestBody))
			assert.Equal(t, tt.expectedStatus, resp.StatusCode, "Test case: %s", tt.name)
			resp.Body.Close()
		})
	}
}

func TestVisitSchedule_Integration_WithConstraints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Test: Create visit schedule with constraints
	t.Run("Create visit schedule with constraints", func(t *testing.T) {
		visitDate := time.Now().AddDate(0, 0, 1).Format(time.RFC3339)
		constraints := map[string]interface{}{
			"time_windows": []map[string]string{
				{
					"start_time": "09:00",
					"end_time":   "12:00",
				},
			},
			"required_skills": []string{"nursing", "medication_management"},
			"patient_notes":   "Patient prefers morning visits",
		}
		constraintsJSON, err := json.Marshal(constraints)
		require.NoError(t, err)

		scheduleJSON := fmt.Sprintf(`{
			"visit_date": "%s",
			"visit_type": "regular",
			"estimated_duration_minutes": 60,
			"status": "draft",
			"priority_score": 5,
			"constraints": %s
		}`, visitDate, string(constraintsJSON))

		resp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/schedules", patientID), strings.NewReader(scheduleJSON))
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var schedule models.VisitSchedule
		DecodeJSONResponse(t, resp, &schedule)

		assert.NotEmpty(t, schedule.ScheduleID)
		assert.NotNil(t, schedule.Constraints)

		// Verify constraints are stored correctly
		var storedConstraints map[string]interface{}
		err = json.Unmarshal(schedule.Constraints, &storedConstraints)
		assert.NoError(t, err)
		assert.NotNil(t, storedConstraints["time_windows"])
		assert.NotNil(t, storedConstraints["required_skills"])
		assert.Equal(t, "Patient prefers morning visits", storedConstraints["patient_notes"])
	})
}
