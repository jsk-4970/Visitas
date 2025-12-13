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

func TestCarePlan_Integration_CreateAndGet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Test: Create a care plan
	t.Run("Create care plan", func(t *testing.T) {
		periodStart := time.Now().Format(time.RFC3339)
		periodEnd := time.Now().AddDate(0, 3, 0).Format(time.RFC3339)

		goals := []map[string]interface{}{
			{
				"goal_id":     "goal-1",
				"description": "Improve mobility",
				"target_date": time.Now().AddDate(0, 1, 0).Format(time.RFC3339),
				"status":      "in_progress",
			},
		}
		goalsJSON, err := json.Marshal(goals)
		require.NoError(t, err)

		carePlanJSON := fmt.Sprintf(`{
			"title": "General Care Plan",
			"status": "active",
			"intent": "plan",
			"period_start": "%s",
			"period_end": "%s",
			"goals": %s
		}`, periodStart, periodEnd, string(goalsJSON))

		resp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/care-plans", patientID), strings.NewReader(carePlanJSON))
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var carePlan models.CarePlan
		DecodeJSONResponse(t, resp, &carePlan)

		assert.NotEmpty(t, carePlan.PlanID)
		assert.Equal(t, patientID, carePlan.PatientID)
		assert.Equal(t, "General Care Plan", carePlan.Title)
		assert.Equal(t, "active", carePlan.Status)
		assert.Equal(t, "plan", carePlan.Intent)

		// Test: Get the created care plan
		t.Run("Get care plan by ID", func(t *testing.T) {
			getResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/patients/%s/care-plans/%s", patientID, carePlan.PlanID), nil)
			assert.Equal(t, http.StatusOK, getResp.StatusCode)

			var retrievedCarePlan models.CarePlan
			DecodeJSONResponse(t, getResp, &retrievedCarePlan)

			assert.Equal(t, carePlan.PlanID, retrievedCarePlan.PlanID)
			assert.Equal(t, carePlan.PatientID, retrievedCarePlan.PatientID)
			assert.Equal(t, carePlan.Title, retrievedCarePlan.Title)
		})
	})
}

func TestCarePlan_Integration_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create a care plan
	periodStart := time.Now().Format(time.RFC3339)
	periodEnd := time.Now().AddDate(0, 3, 0).Format(time.RFC3339)

	carePlanJSON := fmt.Sprintf(`{
		"title": "General Care Plan",
		"status": "draft",
		"intent": "plan",
		"period_start": "%s",
		"period_end": "%s"
	}`, periodStart, periodEnd)

	createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/care-plans", patientID), strings.NewReader(carePlanJSON))
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var carePlan models.CarePlan
	DecodeJSONResponse(t, createResp, &carePlan)

	// Test: Update the care plan
	t.Run("Update care plan", func(t *testing.T) {
		updateJSON := `{
			"title": "Updated Care Plan",
			"status": "active"
		}`

		updateResp := ts.MakeRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/patients/%s/care-plans/%s", patientID, carePlan.PlanID), strings.NewReader(updateJSON))
		assert.Equal(t, http.StatusOK, updateResp.StatusCode)

		var updatedCarePlan models.CarePlan
		DecodeJSONResponse(t, updateResp, &updatedCarePlan)

		assert.Equal(t, carePlan.PlanID, updatedCarePlan.PlanID)
		assert.Equal(t, "Updated Care Plan", updatedCarePlan.Title)
		assert.Equal(t, "active", updatedCarePlan.Status)
	})
}

func TestCarePlan_Integration_List(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create multiple care plans
	carePlanIDs := make([]string, 0)
	for i := 0; i < 3; i++ {
		periodStart := time.Now().AddDate(0, i, 0).Format(time.RFC3339)
		periodEnd := time.Now().AddDate(0, i+3, 0).Format(time.RFC3339)

		carePlanJSON := fmt.Sprintf(`{
			"title": "Care Plan %d",
			"status": "active",
			"intent": "plan",
			"period_start": "%s",
			"period_end": "%s"
		}`, i+1, periodStart, periodEnd)

		createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/care-plans", patientID), strings.NewReader(carePlanJSON))
		require.Equal(t, http.StatusCreated, createResp.StatusCode)

		var carePlan models.CarePlan
		DecodeJSONResponse(t, createResp, &carePlan)
		carePlanIDs = append(carePlanIDs, carePlan.PlanID)
	}

	// Test: List care plans
	t.Run("List care plans", func(t *testing.T) {
		listResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/patients/%s/care-plans", patientID), nil)
		assert.Equal(t, http.StatusOK, listResp.StatusCode)

		var carePlans []models.CarePlan
		DecodeJSONResponse(t, listResp, &carePlans)

		assert.GreaterOrEqual(t, len(carePlans), 3)

		// Verify all created care plans are in the list
		carePlanIDSet := make(map[string]bool)
		for _, carePlan := range carePlans {
			carePlanIDSet[carePlan.PlanID] = true
		}

		for _, carePlanID := range carePlanIDs {
			assert.True(t, carePlanIDSet[carePlanID], "Care plan %s should be in the list", carePlanID)
		}
	})
}

func TestCarePlan_Integration_GetActiveCarePlans(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create an active care plan
	periodStart := time.Now().Format(time.RFC3339)
	periodEnd := time.Now().AddDate(0, 3, 0).Format(time.RFC3339)

	carePlanJSON := fmt.Sprintf(`{
		"title": "Active Care Plan",
		"status": "active",
		"intent": "plan",
		"period_start": "%s",
		"period_end": "%s"
	}`, periodStart, periodEnd)

	createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/care-plans", patientID), strings.NewReader(carePlanJSON))
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var carePlan models.CarePlan
	DecodeJSONResponse(t, createResp, &carePlan)

	// Test: Get active care plans
	t.Run("Get active care plans", func(t *testing.T) {
		activeResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/patients/%s/care-plans/active", patientID), nil)
		assert.Equal(t, http.StatusOK, activeResp.StatusCode)

		var activeCarePlans []models.CarePlan
		DecodeJSONResponse(t, activeResp, &activeCarePlans)

		assert.GreaterOrEqual(t, len(activeCarePlans), 1)

		// Verify the created care plan is in the active list
		foundCarePlan := false
		for _, cp := range activeCarePlans {
			if cp.PlanID == carePlan.PlanID {
				foundCarePlan = true
				assert.Equal(t, "active", cp.Status)
				break
			}
		}
		assert.True(t, foundCarePlan, "Created care plan should be in active care plans")
	})
}

func TestCarePlan_Integration_WithGoalsAndActivities(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Test: Create care plan with goals and activities
	t.Run("Create care plan with goals and activities", func(t *testing.T) {
		periodStart := time.Now().Format(time.RFC3339)
		periodEnd := time.Now().AddDate(0, 3, 0).Format(time.RFC3339)

		goals := []map[string]interface{}{
			{
				"goal_id":     "goal-1",
				"description": "Improve mobility",
				"target_date": time.Now().AddDate(0, 1, 0).Format(time.RFC3339),
				"status":      "in_progress",
			},
			{
				"goal_id":     "goal-2",
				"description": "Reduce pain level",
				"target_date": time.Now().AddDate(0, 2, 0).Format(time.RFC3339),
				"status":      "in_progress",
			},
		}
		goalsJSON, err := json.Marshal(goals)
		require.NoError(t, err)

		activities := []map[string]interface{}{
			{
				"activity_id":  "activity-1",
				"category":     "physical_therapy",
				"description":  "Daily walking exercises",
				"scheduled_at": time.Now().AddDate(0, 0, 1).Format(time.RFC3339),
				"status":       "scheduled",
			},
			{
				"activity_id":  "activity-2",
				"category":     "medication",
				"description":  "Pain medication administration",
				"scheduled_at": time.Now().AddDate(0, 0, 1).Format(time.RFC3339),
				"status":       "scheduled",
			},
		}
		activitiesJSON, err := json.Marshal(activities)
		require.NoError(t, err)

		carePlanJSON := fmt.Sprintf(`{
			"title": "Comprehensive Care Plan",
			"status": "active",
			"intent": "plan",
			"period_start": "%s",
			"period_end": "%s",
			"goals": %s,
			"activities": %s
		}`, periodStart, periodEnd, string(goalsJSON), string(activitiesJSON))

		resp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/care-plans", patientID), strings.NewReader(carePlanJSON))
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var carePlan models.CarePlan
		DecodeJSONResponse(t, resp, &carePlan)

		assert.NotEmpty(t, carePlan.PlanID)

		// Verify goals are stored correctly
		var storedGoals []map[string]interface{}
		err = json.Unmarshal(carePlan.Goals, &storedGoals)
		assert.NoError(t, err)
		assert.Len(t, storedGoals, 2)
		assert.Equal(t, "Improve mobility", storedGoals[0]["description"])

		// Verify activities are stored correctly
		var storedActivities []map[string]interface{}
		err = json.Unmarshal(carePlan.Activities, &storedActivities)
		assert.NoError(t, err)
		assert.Len(t, storedActivities, 2)
		assert.Equal(t, "physical_therapy", storedActivities[0]["category"])
	})
}

func TestCarePlan_Integration_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create a care plan
	periodStart := time.Now().Format(time.RFC3339)
	periodEnd := time.Now().AddDate(0, 3, 0).Format(time.RFC3339)

	carePlanJSON := fmt.Sprintf(`{
		"title": "General Care Plan",
		"status": "draft",
		"intent": "plan",
		"period_start": "%s",
		"period_end": "%s"
	}`, periodStart, periodEnd)

	createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/care-plans", patientID), strings.NewReader(carePlanJSON))
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var carePlan models.CarePlan
	DecodeJSONResponse(t, createResp, &carePlan)

	// Test: Delete the care plan
	t.Run("Delete care plan", func(t *testing.T) {
		deleteResp := ts.MakeRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/patients/%s/care-plans/%s", patientID, carePlan.PlanID), nil)
		assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

		// Verify the care plan is deleted (should return 404)
		getResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/patients/%s/care-plans/%s", patientID, carePlan.PlanID), nil)
		assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
	})
}

func TestCarePlan_Integration_ValidationErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	periodStart := time.Now().Format(time.RFC3339)
	periodEnd := time.Now().AddDate(0, 3, 0).Format(time.RFC3339)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "Missing title",
			requestBody:    fmt.Sprintf(`{"status": "active", "intent": "plan", "period_start": "%s", "period_end": "%s"}`, periodStart, periodEnd),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing status",
			requestBody:    fmt.Sprintf(`{"title": "Care Plan", "intent": "plan", "period_start": "%s", "period_end": "%s"}`, periodStart, periodEnd),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid status",
			requestBody:    fmt.Sprintf(`{"title": "Care Plan", "status": "invalid", "intent": "plan", "period_start": "%s", "period_end": "%s"}`, periodStart, periodEnd),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing intent",
			requestBody:    fmt.Sprintf(`{"title": "Care Plan", "status": "active", "period_start": "%s", "period_end": "%s"}`, periodStart, periodEnd),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid intent",
			requestBody:    fmt.Sprintf(`{"title": "Care Plan", "status": "active", "intent": "invalid", "period_start": "%s", "period_end": "%s"}`, periodStart, periodEnd),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing period_start",
			requestBody:    fmt.Sprintf(`{"title": "Care Plan", "status": "active", "intent": "plan", "period_end": "%s"}`, periodEnd),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/care-plans", patientID), strings.NewReader(tt.requestBody))
			assert.Equal(t, tt.expectedStatus, resp.StatusCode, "Test case: %s", tt.name)
			resp.Body.Close()
		})
	}
}
