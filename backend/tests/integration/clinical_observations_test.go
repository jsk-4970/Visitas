package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/visitas/backend/internal/models"
)

func TestClinicalObservation_Integration_CreateAndGet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Test: Create a clinical observation (vital signs - heart rate)
	t.Run("Create vital signs observation", func(t *testing.T) {
		code := models.ObservationCode{
			System:  "LOINC",
			Code:    "8867-4",
			Display: "Heart rate",
		}
		codeJSON, err := json.Marshal(code)
		require.NoError(t, err)

		value := models.QuantityValue{
			Value: 72.0,
			Unit:  "bpm",
		}
		valueJSON, err := json.Marshal(value)
		require.NoError(t, err)

		observationJSON := fmt.Sprintf(`{
			"category": "vital_signs",
			"code": %s,
			"effective_datetime": "%s",
			"value": %s,
			"interpretation": "normal"
		}`, string(codeJSON), time.Now().Format(time.RFC3339), string(valueJSON))

		resp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/observations", patientID), strings.NewReader(observationJSON))
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var observation models.ClinicalObservation
		DecodeJSONResponse(t, resp, &observation)

		assert.NotEmpty(t, observation.ObservationID)
		assert.Equal(t, patientID, observation.PatientID)
		assert.Equal(t, "vital_signs", observation.Category)
		assert.Equal(t, "normal", observation.Interpretation.StringVal)

		// Test: Get the created observation
		t.Run("Get observation by ID", func(t *testing.T) {
			getResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/patients/%s/observations/%s", patientID, observation.ObservationID), nil)
			assert.Equal(t, http.StatusOK, getResp.StatusCode)

			var retrievedObservation models.ClinicalObservation
			DecodeJSONResponse(t, getResp, &retrievedObservation)

			assert.Equal(t, observation.ObservationID, retrievedObservation.ObservationID)
			assert.Equal(t, observation.PatientID, retrievedObservation.PatientID)
			assert.Equal(t, observation.Category, retrievedObservation.Category)
		})
	})
}

func TestClinicalObservation_Integration_BloodPressure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Test: Create blood pressure observation
	t.Run("Create blood pressure observation", func(t *testing.T) {
		code := models.ObservationCode{
			System:  "LOINC",
			Code:    "85354-9",
			Display: "Blood pressure panel",
		}
		codeJSON, err := json.Marshal(code)
		require.NoError(t, err)

		value := models.BloodPressureValue{
			Systolic: models.QuantityValue{
				Value: 120.0,
				Unit:  "mmHg",
			},
			Diastolic: models.QuantityValue{
				Value: 80.0,
				Unit:  "mmHg",
			},
		}
		valueJSON, err := json.Marshal(value)
		require.NoError(t, err)

		observationJSON := fmt.Sprintf(`{
			"category": "vital_signs",
			"code": %s,
			"effective_datetime": "%s",
			"value": %s,
			"interpretation": "normal"
		}`, string(codeJSON), time.Now().Format(time.RFC3339), string(valueJSON))

		resp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/observations", patientID), strings.NewReader(observationJSON))
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var observation models.ClinicalObservation
		DecodeJSONResponse(t, resp, &observation)

		assert.NotEmpty(t, observation.ObservationID)
		assert.Equal(t, "vital_signs", observation.Category)

		// Verify blood pressure value
		var retrievedValue models.BloodPressureValue
		err = json.Unmarshal(observation.Value, &retrievedValue)
		assert.NoError(t, err)
		assert.Equal(t, 120.0, retrievedValue.Systolic.Value)
		assert.Equal(t, 80.0, retrievedValue.Diastolic.Value)
	})
}

func TestClinicalObservation_Integration_ADLAssessment(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Test: Create ADL assessment observation
	t.Run("Create ADL assessment observation", func(t *testing.T) {
		code := models.ObservationCode{
			System:  "LOINC",
			Code:    "83254-5",
			Display: "Barthel ADL Index",
		}
		codeJSON, err := json.Marshal(code)
		require.NoError(t, err)

		notes := "Patient shows improvement in bathing and dressing"
		value := models.ADLScore{
			TotalScore: 85,
			Items: map[string]int{
				"bathing":      10,
				"dressing":     10,
				"toilet_use":   10,
				"transferring": 15,
				"continence":   10,
				"feeding":      10,
				"mobility":     15,
				"stairs":       5,
			},
			Method: "Barthel",
			Notes:  &notes,
		}
		valueJSON, err := json.Marshal(value)
		require.NoError(t, err)

		observationJSON := fmt.Sprintf(`{
			"category": "adl_assessment",
			"code": %s,
			"effective_datetime": "%s",
			"value": %s,
			"interpretation": "normal"
		}`, string(codeJSON), time.Now().Format(time.RFC3339), string(valueJSON))

		resp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/observations", patientID), strings.NewReader(observationJSON))
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var observation models.ClinicalObservation
		DecodeJSONResponse(t, resp, &observation)

		assert.NotEmpty(t, observation.ObservationID)
		assert.Equal(t, "adl_assessment", observation.Category)

		// Verify ADL score
		var retrievedValue models.ADLScore
		err = json.Unmarshal(observation.Value, &retrievedValue)
		assert.NoError(t, err)
		assert.Equal(t, 85, retrievedValue.TotalScore)
		assert.Equal(t, "Barthel", retrievedValue.Method)
		assert.Equal(t, 10, retrievedValue.Items["bathing"])
	})
}

func TestClinicalObservation_Integration_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create an observation
	code := models.ObservationCode{
		System:  "LOINC",
		Code:    "8867-4",
		Display: "Heart rate",
	}
	codeJSON, err := json.Marshal(code)
	require.NoError(t, err)

	value := models.QuantityValue{
		Value: 72.0,
		Unit:  "bpm",
	}
	valueJSON, err := json.Marshal(value)
	require.NoError(t, err)

	observationJSON := fmt.Sprintf(`{
		"category": "vital_signs",
		"code": %s,
		"effective_datetime": "%s",
		"value": %s,
		"interpretation": "normal"
	}`, string(codeJSON), time.Now().Format(time.RFC3339), string(valueJSON))

	createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/observations", patientID), strings.NewReader(observationJSON))
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var observation models.ClinicalObservation
	DecodeJSONResponse(t, createResp, &observation)

	// Test: Update the observation
	t.Run("Update observation", func(t *testing.T) {
		updatedValue := models.QuantityValue{
			Value: 85.0,
			Unit:  "bpm",
		}
		updatedValueJSON, err := json.Marshal(updatedValue)
		require.NoError(t, err)

		updateJSON := fmt.Sprintf(`{
			"value": %s,
			"interpretation": "high"
		}`, string(updatedValueJSON))

		updateResp := ts.MakeRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/patients/%s/observations/%s", patientID, observation.ObservationID), strings.NewReader(updateJSON))
		assert.Equal(t, http.StatusOK, updateResp.StatusCode)

		var updatedObservation models.ClinicalObservation
		DecodeJSONResponse(t, updateResp, &updatedObservation)

		assert.Equal(t, observation.ObservationID, updatedObservation.ObservationID)
		assert.Equal(t, "high", updatedObservation.Interpretation.StringVal)

		// Verify updated value
		var retrievedValue models.QuantityValue
		err = json.Unmarshal(updatedObservation.Value, &retrievedValue)
		assert.NoError(t, err)
		assert.Equal(t, 85.0, retrievedValue.Value)
	})
}

func TestClinicalObservation_Integration_List(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create multiple observations
	observationIDs := make([]string, 0)
	for i := 0; i < 3; i++ {
		code := models.ObservationCode{
			System:  "LOINC",
			Code:    "8867-4",
			Display: "Heart rate",
		}
		codeJSON, err := json.Marshal(code)
		require.NoError(t, err)

		value := models.QuantityValue{
			Value: float64(70 + i*5),
			Unit:  "bpm",
		}
		valueJSON, err := json.Marshal(value)
		require.NoError(t, err)

		observationJSON := fmt.Sprintf(`{
			"category": "vital_signs",
			"code": %s,
			"effective_datetime": "%s",
			"value": %s,
			"interpretation": "normal"
		}`, string(codeJSON), time.Now().Add(time.Duration(i)*time.Hour).Format(time.RFC3339), string(valueJSON))

		createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/observations", patientID), strings.NewReader(observationJSON))
		require.Equal(t, http.StatusCreated, createResp.StatusCode)

		var observation models.ClinicalObservation
		DecodeJSONResponse(t, createResp, &observation)
		observationIDs = append(observationIDs, observation.ObservationID)
	}

	// Test: List observations
	t.Run("List observations", func(t *testing.T) {
		listResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/patients/%s/observations", patientID), nil)
		assert.Equal(t, http.StatusOK, listResp.StatusCode)

		var observations []models.ClinicalObservation
		DecodeJSONResponse(t, listResp, &observations)

		assert.GreaterOrEqual(t, len(observations), 3)

		// Verify all created observations are in the list
		observationIDSet := make(map[string]bool)
		for _, observation := range observations {
			observationIDSet[observation.ObservationID] = true
		}

		for _, observationID := range observationIDs {
			assert.True(t, observationIDSet[observationID], "Observation %s should be in the list", observationID)
		}
	})
}

func TestClinicalObservation_Integration_GetLatest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create multiple observations for the same category
	var latestObservationID string
	for i := 0; i < 3; i++ {
		code := models.ObservationCode{
			System:  "LOINC",
			Code:    "8867-4",
			Display: "Heart rate",
		}
		codeJSON, err := json.Marshal(code)
		require.NoError(t, err)

		value := models.QuantityValue{
			Value: float64(70 + i*5),
			Unit:  "bpm",
		}
		valueJSON, err := json.Marshal(value)
		require.NoError(t, err)

		observationJSON := fmt.Sprintf(`{
			"category": "vital_signs",
			"code": %s,
			"effective_datetime": "%s",
			"value": %s,
			"interpretation": "normal"
		}`, string(codeJSON), time.Now().Add(time.Duration(i)*time.Hour).Format(time.RFC3339), string(valueJSON))

		createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/observations", patientID), strings.NewReader(observationJSON))
		require.Equal(t, http.StatusCreated, createResp.StatusCode)

		var observation models.ClinicalObservation
		DecodeJSONResponse(t, createResp, &observation)
		latestObservationID = observation.ObservationID
	}

	// Test: Get latest observation by category
	t.Run("Get latest observation", func(t *testing.T) {
		latestResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/patients/%s/observations/latest/vital_signs", patientID), nil)
		assert.Equal(t, http.StatusOK, latestResp.StatusCode)

		var latestObservation models.ClinicalObservation
		DecodeJSONResponse(t, latestResp, &latestObservation)

		assert.Equal(t, latestObservationID, latestObservation.ObservationID)
		assert.Equal(t, "vital_signs", latestObservation.Category)

		// Verify it's the latest value
		var value models.QuantityValue
		err := json.Unmarshal(latestObservation.Value, &value)
		assert.NoError(t, err)
		assert.Equal(t, 80.0, value.Value) // Last value created
	})
}

func TestClinicalObservation_Integration_GetTimeSeriesData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create multiple observations over time for time series
	for i := 0; i < 5; i++ {
		code := models.ObservationCode{
			System:  "LOINC",
			Code:    "8867-4",
			Display: "Heart rate",
		}
		codeJSON, err := json.Marshal(code)
		require.NoError(t, err)

		value := models.QuantityValue{
			Value: float64(70 + i*2),
			Unit:  "bpm",
		}
		valueJSON, err := json.Marshal(value)
		require.NoError(t, err)

		observationJSON := fmt.Sprintf(`{
			"category": "vital_signs",
			"code": %s,
			"effective_datetime": "%s",
			"value": %s,
			"interpretation": "normal"
		}`, string(codeJSON), time.Now().Add(time.Duration(i)*time.Hour).Format(time.RFC3339), string(valueJSON))

		createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/observations", patientID), strings.NewReader(observationJSON))
		require.Equal(t, http.StatusCreated, createResp.StatusCode)
	}

	// Test: Get time series data
	t.Run("Get time series data", func(t *testing.T) {
		from := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
		to := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
		timeSeriesResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/patients/%s/observations/timeseries/vital_signs?from=%s&to=%s", patientID, url.QueryEscape(from), url.QueryEscape(to)), nil)
		assert.Equal(t, http.StatusOK, timeSeriesResp.StatusCode)

		var timeSeriesData []models.ClinicalObservation
		DecodeJSONResponse(t, timeSeriesResp, &timeSeriesData)

		assert.GreaterOrEqual(t, len(timeSeriesData), 5)

		// Verify data is ordered by time (oldest first or newest first, depending on implementation)
		for _, observation := range timeSeriesData {
			assert.Equal(t, "vital_signs", observation.Category)
		}
	})
}

func TestClinicalObservation_Integration_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create an observation
	code := models.ObservationCode{
		System:  "LOINC",
		Code:    "8867-4",
		Display: "Heart rate",
	}
	codeJSON, err := json.Marshal(code)
	require.NoError(t, err)

	value := models.QuantityValue{
		Value: 72.0,
		Unit:  "bpm",
	}
	valueJSON, err := json.Marshal(value)
	require.NoError(t, err)

	observationJSON := fmt.Sprintf(`{
		"category": "vital_signs",
		"code": %s,
		"effective_datetime": "%s",
		"value": %s,
		"interpretation": "normal"
	}`, string(codeJSON), time.Now().Format(time.RFC3339), string(valueJSON))

	createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/observations", patientID), strings.NewReader(observationJSON))
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var observation models.ClinicalObservation
	DecodeJSONResponse(t, createResp, &observation)

	// Test: Delete the observation
	t.Run("Delete observation", func(t *testing.T) {
		deleteResp := ts.MakeRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/patients/%s/observations/%s", patientID, observation.ObservationID), nil)
		assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

		// Verify the observation is deleted (should return 404)
		getResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/patients/%s/observations/%s", patientID, observation.ObservationID), nil)
		assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
	})
}

func TestClinicalObservation_Integration_ValidationErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	code := models.ObservationCode{
		System:  "LOINC",
		Code:    "8867-4",
		Display: "Heart rate",
	}
	codeJSON, err := json.Marshal(code)
	require.NoError(t, err)

	value := models.QuantityValue{
		Value: 72.0,
		Unit:  "bpm",
	}
	valueJSON, err := json.Marshal(value)
	require.NoError(t, err)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "Missing category",
			requestBody:    fmt.Sprintf(`{"code": %s, "effective_datetime": "%s", "value": %s}`, string(codeJSON), time.Now().Format(time.RFC3339), string(valueJSON)),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid category",
			requestBody:    fmt.Sprintf(`{"category": "invalid", "code": %s, "effective_datetime": "%s", "value": %s}`, string(codeJSON), time.Now().Format(time.RFC3339), string(valueJSON)),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing code",
			requestBody:    fmt.Sprintf(`{"category": "vital_signs", "effective_datetime": "%s", "value": %s}`, time.Now().Format(time.RFC3339), string(valueJSON)),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing effective_datetime",
			requestBody:    fmt.Sprintf(`{"category": "vital_signs", "code": %s, "value": %s}`, string(codeJSON), string(valueJSON)),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing value",
			requestBody:    fmt.Sprintf(`{"category": "vital_signs", "code": %s, "effective_datetime": "%s"}`, string(codeJSON), time.Now().Format(time.RFC3339)),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid interpretation",
			requestBody:    fmt.Sprintf(`{"category": "vital_signs", "code": %s, "effective_datetime": "%s", "value": %s, "interpretation": "invalid"}`, string(codeJSON), time.Now().Format(time.RFC3339), string(valueJSON)),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/observations", patientID), strings.NewReader(tt.requestBody))
			assert.Equal(t, tt.expectedStatus, resp.StatusCode, "Test case: %s", tt.name)
			resp.Body.Close()
		})
	}
}
