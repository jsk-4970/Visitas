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

func TestACPRecord_Integration_CreateAndGet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Test: Create an ACP record
	t.Run("Create ACP record", func(t *testing.T) {
		directives := map[string]interface{}{
			"dnar": true,
			"mechanical_ventilation": false,
			"artificial_nutrition": true,
			"cardiopulmonary_resuscitation": false,
			"life_sustaining_treatment": "limited",
		}
		directivesJSON, err := json.Marshal(directives)
		require.NoError(t, err)

		acpJSON := fmt.Sprintf(`{
			"recorded_date": "%s",
			"status": "active",
			"decision_maker": "patient",
			"directives": %s,
			"values_narrative": "患者は自然な最期を希望されています",
			"created_by": "test-staff-id"
		}`, time.Now().Format(time.RFC3339), string(directivesJSON))

		resp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/acp-records", patientID), strings.NewReader(acpJSON))
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var acp models.ACPRecord
		DecodeJSONResponse(t, resp, &acp)

		assert.NotEmpty(t, acp.ACPID)
		assert.Equal(t, patientID, acp.PatientID)
		assert.Equal(t, "active", acp.Status)
		assert.Equal(t, "patient", acp.DecisionMaker)
		assert.Equal(t, int64(1), acp.Version)
		assert.Equal(t, "highly_confidential", acp.DataSensitivity)

		// Test: Get the created ACP record
		t.Run("Get ACP record by ID", func(t *testing.T) {
			getResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/patients/%s/acp-records/%s", patientID, acp.ACPID), nil)
			assert.Equal(t, http.StatusOK, getResp.StatusCode)

			var retrievedACP models.ACPRecord
			DecodeJSONResponse(t, getResp, &retrievedACP)

			assert.Equal(t, acp.ACPID, retrievedACP.ACPID)
			assert.Equal(t, acp.PatientID, retrievedACP.PatientID)
			assert.Equal(t, acp.Status, retrievedACP.Status)
		})
	})
}

func TestACPRecord_Integration_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create an ACP record
	directives := map[string]interface{}{
		"dnar": true,
		"mechanical_ventilation": false,
	}
	directivesJSON, err := json.Marshal(directives)
	require.NoError(t, err)

	acpJSON := fmt.Sprintf(`{
		"recorded_date": "%s",
		"status": "draft",
		"decision_maker": "patient",
		"directives": %s,
		"created_by": "test-staff-id"
	}`, time.Now().Format(time.RFC3339), string(directivesJSON))

	createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/acp-records", patientID), strings.NewReader(acpJSON))
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var acp models.ACPRecord
	DecodeJSONResponse(t, createResp, &acp)

	// Test: Update the ACP record
	t.Run("Update ACP record", func(t *testing.T) {
		updatedDirectives := map[string]interface{}{
			"dnar": true,
			"mechanical_ventilation": false,
			"artificial_nutrition": true,
		}
		updatedDirectivesJSON, err := json.Marshal(updatedDirectives)
		require.NoError(t, err)

		updateJSON := fmt.Sprintf(`{
			"status": "active",
			"directives": %s,
			"values_narrative": "Updated: 患者の意思が明確になりました"
		}`, string(updatedDirectivesJSON))

		updateResp := ts.MakeRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/patients/%s/acp-records/%s", patientID, acp.ACPID), strings.NewReader(updateJSON))
		assert.Equal(t, http.StatusOK, updateResp.StatusCode)

		var updatedACP models.ACPRecord
		DecodeJSONResponse(t, updateResp, &updatedACP)

		assert.Equal(t, acp.ACPID, updatedACP.ACPID)
		assert.Equal(t, "active", updatedACP.Status)
		assert.True(t, updatedACP.ValuesNarrative.Valid)
		assert.Contains(t, updatedACP.ValuesNarrative.StringVal, "Updated")
	})
}

func TestACPRecord_Integration_List(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create multiple ACP records
	acpIDs := make([]string, 0)
	for i := 0; i < 3; i++ {
		directives := map[string]interface{}{
			"dnar": i%2 == 0,
			"version": i + 1,
		}
		directivesJSON, err := json.Marshal(directives)
		require.NoError(t, err)

		acpJSON := fmt.Sprintf(`{
			"recorded_date": "%s",
			"status": "active",
			"decision_maker": "patient",
			"directives": %s,
			"created_by": "test-staff-id"
		}`, time.Now().Add(time.Duration(i)*time.Hour).Format(time.RFC3339), string(directivesJSON))

		createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/acp-records", patientID), strings.NewReader(acpJSON))
		require.Equal(t, http.StatusCreated, createResp.StatusCode)

		var acp models.ACPRecord
		DecodeJSONResponse(t, createResp, &acp)
		acpIDs = append(acpIDs, acp.ACPID)
	}

	// Test: List ACP records
	t.Run("List ACP records", func(t *testing.T) {
		listResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/patients/%s/acp-records", patientID), nil)
		assert.Equal(t, http.StatusOK, listResp.StatusCode)

		var acpRecords []models.ACPRecord
		DecodeJSONResponse(t, listResp, &acpRecords)

		assert.GreaterOrEqual(t, len(acpRecords), 3)

		// Verify all created ACP records are in the list
		acpIDSet := make(map[string]bool)
		for _, acp := range acpRecords {
			acpIDSet[acp.ACPID] = true
		}

		for _, acpID := range acpIDs {
			assert.True(t, acpIDSet[acpID], "ACP record %s should be in the list", acpID)
		}
	})
}

func TestACPRecord_Integration_GetLatest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create multiple ACP records to simulate version history
	var latestACPID string
	for i := 0; i < 3; i++ {
		directives := map[string]interface{}{
			"dnar": true,
			"version": i + 1,
		}
		directivesJSON, err := json.Marshal(directives)
		require.NoError(t, err)

		status := "active"
		if i < 2 {
			status = "superseded" // Older versions are superseded
		}

		acpJSON := fmt.Sprintf(`{
			"recorded_date": "%s",
			"status": "%s",
			"decision_maker": "patient",
			"directives": %s,
			"created_by": "test-staff-id"
		}`, time.Now().Add(time.Duration(i)*time.Hour).Format(time.RFC3339), status, string(directivesJSON))

		createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/acp-records", patientID), strings.NewReader(acpJSON))
		require.Equal(t, http.StatusCreated, createResp.StatusCode)

		var acp models.ACPRecord
		DecodeJSONResponse(t, createResp, &acp)
		if i == 2 {
			latestACPID = acp.ACPID
		}
	}

	// Test: Get latest ACP
	t.Run("Get latest ACP", func(t *testing.T) {
		latestResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/patients/%s/acp-records/latest", patientID), nil)
		assert.Equal(t, http.StatusOK, latestResp.StatusCode)

		var latestACP models.ACPRecord
		DecodeJSONResponse(t, latestResp, &latestACP)

		assert.Equal(t, latestACPID, latestACP.ACPID)
		assert.Equal(t, "active", latestACP.Status)
	})
}

func TestACPRecord_Integration_GetHistory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create multiple ACP records for version history
	for i := 0; i < 3; i++ {
		directives := map[string]interface{}{
			"dnar": true,
			"version": i + 1,
		}
		directivesJSON, err := json.Marshal(directives)
		require.NoError(t, err)

		status := "active"
		if i < 2 {
			status = "superseded"
		}

		acpJSON := fmt.Sprintf(`{
			"recorded_date": "%s",
			"status": "%s",
			"decision_maker": "patient",
			"directives": %s,
			"created_by": "test-staff-id"
		}`, time.Now().Add(time.Duration(i)*time.Hour).Format(time.RFC3339), status, string(directivesJSON))

		createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/acp-records", patientID), strings.NewReader(acpJSON))
		require.Equal(t, http.StatusCreated, createResp.StatusCode)
	}

	// Test: Get ACP history
	t.Run("Get ACP history", func(t *testing.T) {
		historyResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/patients/%s/acp-records/history", patientID), nil)
		assert.Equal(t, http.StatusOK, historyResp.StatusCode)

		var acpHistory []models.ACPRecord
		DecodeJSONResponse(t, historyResp, &acpHistory)

		assert.GreaterOrEqual(t, len(acpHistory), 3)

		// Verify that history includes both active and superseded records
		hasActive := false
		hasSuperseded := false
		for _, acp := range acpHistory {
			if acp.Status == "active" {
				hasActive = true
			}
			if acp.Status == "superseded" {
				hasSuperseded = true
			}
		}
		assert.True(t, hasActive, "History should include active ACP")
		assert.True(t, hasSuperseded, "History should include superseded ACPs")
	})
}

func TestACPRecord_Integration_WithProxy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Test: Create ACP record with proxy decision maker
	t.Run("Create ACP record with proxy", func(t *testing.T) {
		directives := map[string]interface{}{
			"dnar": true,
			"mechanical_ventilation": false,
		}
		directivesJSON, err := json.Marshal(directives)
		require.NoError(t, err)

		acpJSON := fmt.Sprintf(`{
			"recorded_date": "%s",
			"status": "active",
			"decision_maker": "proxy",
			"proxy_person_id": "proxy-123",
			"directives": %s,
			"values_narrative": "代理人による意思決定",
			"created_by": "test-staff-id"
		}`, time.Now().Format(time.RFC3339), string(directivesJSON))

		resp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/acp-records", patientID), strings.NewReader(acpJSON))
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var acp models.ACPRecord
		DecodeJSONResponse(t, resp, &acp)

		assert.NotEmpty(t, acp.ACPID)
		assert.Equal(t, "proxy", acp.DecisionMaker)
		assert.True(t, acp.ProxyPersonID.Valid)
		assert.Equal(t, "proxy-123", acp.ProxyPersonID.StringVal)
	})
}

func TestACPRecord_Integration_WithLegalDocuments(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Test: Create ACP record with legal documents
	t.Run("Create ACP record with legal documents", func(t *testing.T) {
		directives := map[string]interface{}{
			"dnar": true,
			"mechanical_ventilation": false,
		}
		directivesJSON, err := json.Marshal(directives)
		require.NoError(t, err)

		legalDocuments := []map[string]interface{}{
			{
				"document_type": "living_will",
				"signed_date": time.Now().Format(time.RFC3339),
				"document_url": "gs://bucket/living_will.pdf",
				"witness_1": "witness-1-id",
				"witness_2": "witness-2-id",
			},
			{
				"document_type": "healthcare_proxy_form",
				"signed_date": time.Now().Format(time.RFC3339),
				"document_url": "gs://bucket/proxy_form.pdf",
			},
		}
		legalDocumentsJSON, err := json.Marshal(legalDocuments)
		require.NoError(t, err)

		acpJSON := fmt.Sprintf(`{
			"recorded_date": "%s",
			"status": "active",
			"decision_maker": "patient",
			"directives": %s,
			"legal_documents": %s,
			"created_by": "test-staff-id"
		}`, time.Now().Format(time.RFC3339), string(directivesJSON), string(legalDocumentsJSON))

		resp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/acp-records", patientID), strings.NewReader(acpJSON))
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var acp models.ACPRecord
		DecodeJSONResponse(t, resp, &acp)

		assert.NotEmpty(t, acp.ACPID)
		assert.NotNil(t, acp.LegalDocuments)

		// Verify legal documents are stored correctly
		var storedDocuments []map[string]interface{}
		err = json.Unmarshal(acp.LegalDocuments, &storedDocuments)
		assert.NoError(t, err)
		assert.Len(t, storedDocuments, 2)
		assert.Equal(t, "living_will", storedDocuments[0]["document_type"])
		assert.Equal(t, "healthcare_proxy_form", storedDocuments[1]["document_type"])
	})
}

func TestACPRecord_Integration_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create an ACP record
	directives := map[string]interface{}{
		"dnar": true,
	}
	directivesJSON, err := json.Marshal(directives)
	require.NoError(t, err)

	acpJSON := fmt.Sprintf(`{
		"recorded_date": "%s",
		"status": "draft",
		"decision_maker": "patient",
		"directives": %s,
		"created_by": "test-staff-id"
	}`, time.Now().Format(time.RFC3339), string(directivesJSON))

	createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/acp-records", patientID), strings.NewReader(acpJSON))
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var acp models.ACPRecord
	DecodeJSONResponse(t, createResp, &acp)

	// Test: Delete the ACP record
	t.Run("Delete ACP record", func(t *testing.T) {
		deleteResp := ts.MakeRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/patients/%s/acp-records/%s", patientID, acp.ACPID), nil)
		assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

		// Verify the ACP record is deleted (should return 404)
		getResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/patients/%s/acp-records/%s", patientID, acp.ACPID), nil)
		assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
	})
}

func TestACPRecord_Integration_ValidationErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	directives := map[string]interface{}{
		"dnar": true,
	}
	directivesJSON, err := json.Marshal(directives)
	require.NoError(t, err)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "Missing recorded_date",
			requestBody:    fmt.Sprintf(`{"status": "active", "decision_maker": "patient", "directives": %s, "created_by": "test-staff-id"}`, string(directivesJSON)),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing status",
			requestBody:    fmt.Sprintf(`{"recorded_date": "%s", "decision_maker": "patient", "directives": %s, "created_by": "test-staff-id"}`, time.Now().Format(time.RFC3339), string(directivesJSON)),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid status",
			requestBody:    fmt.Sprintf(`{"recorded_date": "%s", "status": "invalid", "decision_maker": "patient", "directives": %s, "created_by": "test-staff-id"}`, time.Now().Format(time.RFC3339), string(directivesJSON)),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing decision_maker",
			requestBody:    fmt.Sprintf(`{"recorded_date": "%s", "status": "active", "directives": %s, "created_by": "test-staff-id"}`, time.Now().Format(time.RFC3339), string(directivesJSON)),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid decision_maker",
			requestBody:    fmt.Sprintf(`{"recorded_date": "%s", "status": "active", "decision_maker": "invalid", "directives": %s, "created_by": "test-staff-id"}`, time.Now().Format(time.RFC3339), string(directivesJSON)),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing directives",
			requestBody:    fmt.Sprintf(`{"recorded_date": "%s", "status": "active", "decision_maker": "patient", "created_by": "test-staff-id"}`, time.Now().Format(time.RFC3339)),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing created_by",
			requestBody:    fmt.Sprintf(`{"recorded_date": "%s", "status": "active", "decision_maker": "patient", "directives": %s}`, time.Now().Format(time.RFC3339), string(directivesJSON)),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/acp-records", patientID), strings.NewReader(tt.requestBody))
			assert.Equal(t, tt.expectedStatus, resp.StatusCode, "Test case: %s", tt.name)
			resp.Body.Close()
		})
	}
}
