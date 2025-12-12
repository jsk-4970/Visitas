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

func TestMedicalRecord_Integration_CreateAndGet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Test: Create a medical record
	t.Run("Create medical record", func(t *testing.T) {
		visitStartedAt := time.Now().Format(time.RFC3339)

		soapContent := map[string]interface{}{
			"subjective": map[string]interface{}{
				"chiefComplaint":   "呼吸苦、SpO2低下",
				"patientNarrative": "横になると苦しい",
			},
			"objective": map[string]interface{}{
				"vitalSigns": map[string]interface{}{
					"bloodPressure": map[string]interface{}{
						"systolic":  135,
						"diastolic": 82,
					},
					"heartRate": map[string]interface{}{
						"value": 78,
					},
					"spo2": map[string]interface{}{
						"value": 92,
					},
				},
			},
			"assessment": map[string]interface{}{
				"clinicalImpression": "左下葉肺炎疑い",
			},
			"plan": map[string]interface{}{
				"careInstructions": map[string]interface{}{
					"monitoring": []string{"SpO2測定 1日4回"},
				},
			},
		}
		soapContentJSON, err := json.Marshal(soapContent)
		require.NoError(t, err)

		medicalRecordJSON := fmt.Sprintf(`{
			"visit_started_at": "%s",
			"visit_type": "regular",
			"performed_by": "test-staff-id",
			"status": "draft",
			"source_type": "manual",
			"soap_content": %s
		}`, visitStartedAt, string(soapContentJSON))

		resp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/medical-records", patientID), strings.NewReader(medicalRecordJSON))
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var medicalRecord models.MedicalRecord
		DecodeJSONResponse(t, resp, &medicalRecord)

		assert.NotEmpty(t, medicalRecord.RecordID)
		assert.Equal(t, patientID, medicalRecord.PatientID)
		assert.Equal(t, "regular", medicalRecord.VisitType)
		assert.Equal(t, "draft", medicalRecord.Status)
		assert.Equal(t, "manual", medicalRecord.SourceType)

		// Test: Get the created medical record
		t.Run("Get medical record by ID", func(t *testing.T) {
			getResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/patients/%s/medical-records/%s", patientID, medicalRecord.RecordID), nil)
			assert.Equal(t, http.StatusOK, getResp.StatusCode)

			var retrievedRecord models.MedicalRecord
			DecodeJSONResponse(t, getResp, &retrievedRecord)

			assert.Equal(t, medicalRecord.RecordID, retrievedRecord.RecordID)
			assert.Equal(t, medicalRecord.PatientID, retrievedRecord.PatientID)
			assert.Equal(t, medicalRecord.VisitType, retrievedRecord.VisitType)
		})
	})
}

func TestMedicalRecord_Integration_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create a medical record
	visitStartedAt := time.Now().Format(time.RFC3339)

	medicalRecordJSON := fmt.Sprintf(`{
		"visit_started_at": "%s",
		"visit_type": "regular",
		"performed_by": "test-staff-id",
		"status": "draft",
		"source_type": "manual"
	}`, visitStartedAt)

	createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/medical-records", patientID), strings.NewReader(medicalRecordJSON))
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var medicalRecord models.MedicalRecord
	DecodeJSONResponse(t, createResp, &medicalRecord)

	// Test: Update the medical record
	t.Run("Update medical record", func(t *testing.T) {
		newSOAPContent := map[string]interface{}{
			"subjective": map[string]interface{}{
				"chiefComplaint": "更新された訴え",
			},
		}
		soapContentJSON, err := json.Marshal(newSOAPContent)
		require.NoError(t, err)

		updateJSON := fmt.Sprintf(`{
			"status": "in_progress",
			"soap_content": %s,
			"expected_version": %d
		}`, string(soapContentJSON), medicalRecord.Version)

		updateResp := ts.MakeRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/patients/%s/medical-records/%s", patientID, medicalRecord.RecordID), strings.NewReader(updateJSON))
		assert.Equal(t, http.StatusOK, updateResp.StatusCode)

		var updatedRecord models.MedicalRecord
		DecodeJSONResponse(t, updateResp, &updatedRecord)

		assert.Equal(t, medicalRecord.RecordID, updatedRecord.RecordID)
		assert.Equal(t, "in_progress", updatedRecord.Status)
		assert.Equal(t, medicalRecord.Version+1, updatedRecord.Version)
	})
}

func TestMedicalRecord_Integration_OptimisticLocking(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create a medical record
	visitStartedAt := time.Now().Format(time.RFC3339)

	medicalRecordJSON := fmt.Sprintf(`{
		"visit_started_at": "%s",
		"visit_type": "regular",
		"performed_by": "test-staff-id",
		"status": "draft",
		"source_type": "manual"
	}`, visitStartedAt)

	createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/medical-records", patientID), strings.NewReader(medicalRecordJSON))
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var medicalRecord models.MedicalRecord
	DecodeJSONResponse(t, createResp, &medicalRecord)

	// First update (should succeed)
	updateJSON1 := fmt.Sprintf(`{
		"status": "in_progress",
		"expected_version": %d
	}`, medicalRecord.Version)

	updateResp1 := ts.MakeRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/patients/%s/medical-records/%s", patientID, medicalRecord.RecordID), strings.NewReader(updateJSON1))
	require.Equal(t, http.StatusOK, updateResp1.StatusCode)

	var updatedRecord models.MedicalRecord
	DecodeJSONResponse(t, updateResp1, &updatedRecord)

	// Test: Concurrent update with stale version (should fail with 409 Conflict)
	t.Run("Concurrent edit conflict", func(t *testing.T) {
		updateJSON2 := fmt.Sprintf(`{
			"status": "completed",
			"expected_version": %d
		}`, medicalRecord.Version) // Using stale version

		updateResp2 := ts.MakeRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/patients/%s/medical-records/%s", patientID, medicalRecord.RecordID), strings.NewReader(updateJSON2))
		assert.Equal(t, http.StatusConflict, updateResp2.StatusCode)
	})
}

func TestMedicalRecord_Integration_List(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create multiple medical records
	recordIDs := make([]string, 0)
	for i := 0; i < 3; i++ {
		visitStartedAt := time.Now().Add(time.Duration(i) * time.Hour).Format(time.RFC3339)

		medicalRecordJSON := fmt.Sprintf(`{
			"visit_started_at": "%s",
			"visit_type": "regular",
			"performed_by": "test-staff-id",
			"status": "draft",
			"source_type": "manual"
		}`, visitStartedAt)

		createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/medical-records", patientID), strings.NewReader(medicalRecordJSON))
		require.Equal(t, http.StatusCreated, createResp.StatusCode)

		var medicalRecord models.MedicalRecord
		DecodeJSONResponse(t, createResp, &medicalRecord)
		recordIDs = append(recordIDs, medicalRecord.RecordID)
	}

	// Test: List medical records
	t.Run("List medical records", func(t *testing.T) {
		listResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/patients/%s/medical-records", patientID), nil)
		assert.Equal(t, http.StatusOK, listResp.StatusCode)

		var records []models.MedicalRecord
		DecodeJSONResponse(t, listResp, &records)

		assert.GreaterOrEqual(t, len(records), 3)

		// Verify all created records are in the list
		recordIDSet := make(map[string]bool)
		for _, record := range records {
			recordIDSet[record.RecordID] = true
		}

		for _, recordID := range recordIDs {
			assert.True(t, recordIDSet[recordID], "Record %s should be in the list", recordID)
		}
	})

	// Test: List with filter
	t.Run("List medical records with status filter", func(t *testing.T) {
		listResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/patients/%s/medical-records?status=draft", patientID), nil)
		assert.Equal(t, http.StatusOK, listResp.StatusCode)

		var records []models.MedicalRecord
		DecodeJSONResponse(t, listResp, &records)

		for _, record := range records {
			assert.Equal(t, "draft", record.Status)
		}
	})
}

func TestMedicalRecord_Integration_GetLatest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create a medical record
	visitStartedAt := time.Now().Format(time.RFC3339)

	medicalRecordJSON := fmt.Sprintf(`{
		"visit_started_at": "%s",
		"visit_type": "regular",
		"performed_by": "test-staff-id",
		"status": "completed",
		"source_type": "manual"
	}`, visitStartedAt)

	createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/medical-records", patientID), strings.NewReader(medicalRecordJSON))
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var medicalRecord models.MedicalRecord
	DecodeJSONResponse(t, createResp, &medicalRecord)

	// Test: Get latest records
	t.Run("Get latest medical records", func(t *testing.T) {
		latestResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/patients/%s/medical-records/latest?limit=5", patientID), nil)
		assert.Equal(t, http.StatusOK, latestResp.StatusCode)

		var latestRecords []models.MedicalRecord
		DecodeJSONResponse(t, latestResp, &latestRecords)

		assert.GreaterOrEqual(t, len(latestRecords), 1)

		// Verify the created record is in the list
		foundRecord := false
		for _, record := range latestRecords {
			if record.RecordID == medicalRecord.RecordID {
				foundRecord = true
				break
			}
		}
		assert.True(t, foundRecord, "Created medical record should be in the latest records")
	})
}

func TestMedicalRecord_Integration_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create a medical record
	visitStartedAt := time.Now().Format(time.RFC3339)

	medicalRecordJSON := fmt.Sprintf(`{
		"visit_started_at": "%s",
		"visit_type": "regular",
		"performed_by": "test-staff-id",
		"status": "draft",
		"source_type": "manual"
	}`, visitStartedAt)

	createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/medical-records", patientID), strings.NewReader(medicalRecordJSON))
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var medicalRecord models.MedicalRecord
	DecodeJSONResponse(t, createResp, &medicalRecord)

	// Test: Delete the medical record
	t.Run("Delete medical record (soft delete)", func(t *testing.T) {
		deleteResp := ts.MakeRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/patients/%s/medical-records/%s", patientID, medicalRecord.RecordID), nil)
		assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

		// Verify the record is deleted (should return 404)
		getResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/patients/%s/medical-records/%s", patientID, medicalRecord.RecordID), nil)
		assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
	})
}

func TestMedicalRecord_Integration_WithSOAPContent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Test: Create medical record with full SOAP content
	t.Run("Create medical record with SOAP content", func(t *testing.T) {
		visitStartedAt := time.Now().Format(time.RFC3339)

		soapContent := map[string]interface{}{
			"subjective": map[string]interface{}{
				"chiefComplaint":   "呼吸苦、SpO2低下",
				"patientNarrative": "横になると苦しい。食欲も低下している。",
				"symptoms": []map[string]interface{}{
					{
						"code":     "SNOMED:267036007",
						"display":  "呼吸困難",
						"severity": "moderate",
					},
				},
			},
			"objective": map[string]interface{}{
				"vitalSigns": map[string]interface{}{
					"bloodPressure": map[string]interface{}{
						"systolic":  135,
						"diastolic": 82,
						"unit":      "mmHg",
					},
					"heartRate": map[string]interface{}{
						"value": 78,
						"unit":  "bpm",
					},
					"spo2": map[string]interface{}{
						"value": 92,
						"unit":  "%",
					},
				},
				"physicalExam": map[string]interface{}{
					"general":     "軽度呼吸促迫あり",
					"respiratory": "呼吸音減弱、左下肺野湿性ラ音聴取",
				},
			},
			"assessment": map[string]interface{}{
				"diagnoses": []map[string]interface{}{
					{
						"code": map[string]interface{}{
							"system":  "ICD-10",
							"code":    "J18.9",
							"display": "肺炎、詳細不明",
						},
						"status": "provisional",
					},
				},
				"clinicalImpression": "左下葉肺炎疑い。酸素化不良あり。",
			},
			"plan": map[string]interface{}{
				"medications": []map[string]interface{}{
					{
						"action":   "prescribe",
						"drugName": "レボフロキサシン錠500mg",
						"dosage":   "1日1回 朝食後 7日間",
					},
				},
				"careInstructions": map[string]interface{}{
					"monitoring": []string{"SpO2測定 1日4回", "体温測定 1日2回"},
				},
			},
		}
		soapContentJSON, err := json.Marshal(soapContent)
		require.NoError(t, err)

		medicalRecordJSON := fmt.Sprintf(`{
			"visit_started_at": "%s",
			"visit_type": "regular",
			"performed_by": "test-staff-id",
			"status": "completed",
			"source_type": "manual",
			"soap_content": %s
		}`, visitStartedAt, string(soapContentJSON))

		resp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/medical-records", patientID), strings.NewReader(medicalRecordJSON))
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var medicalRecord models.MedicalRecord
		DecodeJSONResponse(t, resp, &medicalRecord)

		assert.NotEmpty(t, medicalRecord.RecordID)

		// Verify SOAP content is stored correctly
		var storedSOAP map[string]interface{}
		err = json.Unmarshal(medicalRecord.SOAPContent, &storedSOAP)
		assert.NoError(t, err)

		// Check subjective
		subjective, ok := storedSOAP["subjective"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "呼吸苦、SpO2低下", subjective["chiefComplaint"])

		// Check objective
		objective, ok := storedSOAP["objective"].(map[string]interface{})
		assert.True(t, ok)
		vitalSigns, ok := objective["vitalSigns"].(map[string]interface{})
		assert.True(t, ok)
		spo2, ok := vitalSigns["spo2"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, float64(92), spo2["value"])

		// Check assessment
		assessment, ok := storedSOAP["assessment"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "左下葉肺炎疑い。酸素化不良あり。", assessment["clinicalImpression"])
	})
}

func TestMedicalRecord_Integration_ValidationErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	visitStartedAt := time.Now().Format(time.RFC3339)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "Missing visit_started_at",
			requestBody:    `{"visit_type": "regular", "performed_by": "test-staff-id", "status": "draft", "source_type": "manual"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing visit_type",
			requestBody:    fmt.Sprintf(`{"visit_started_at": "%s", "performed_by": "test-staff-id", "status": "draft", "source_type": "manual"}`, visitStartedAt),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid visit_type",
			requestBody:    fmt.Sprintf(`{"visit_started_at": "%s", "visit_type": "invalid", "performed_by": "test-staff-id", "status": "draft", "source_type": "manual"}`, visitStartedAt),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing performed_by",
			requestBody:    fmt.Sprintf(`{"visit_started_at": "%s", "visit_type": "regular", "status": "draft", "source_type": "manual"}`, visitStartedAt),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid status",
			requestBody:    fmt.Sprintf(`{"visit_started_at": "%s", "visit_type": "regular", "performed_by": "test-staff-id", "status": "invalid", "source_type": "manual"}`, visitStartedAt),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid source_type",
			requestBody:    fmt.Sprintf(`{"visit_started_at": "%s", "visit_type": "regular", "performed_by": "test-staff-id", "status": "draft", "source_type": "invalid"}`, visitStartedAt),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/medical-records", patientID), strings.NewReader(tt.requestBody))
			assert.Equal(t, tt.expectedStatus, resp.StatusCode, "Test case: %s", tt.name)
			resp.Body.Close()
		})
	}
}

// Template tests

func TestMedicalRecordTemplate_Integration_CreateAndGet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Test: Create a template
	t.Run("Create template", func(t *testing.T) {
		soapTemplate := map[string]interface{}{
			"subjective": map[string]interface{}{
				"placeholder":     "患者の訴え・症状を記載",
				"common_phrases":  []string{"自覚症状なし", "疼痛あり"},
			},
			"objective": map[string]interface{}{
				"placeholder": "バイタルサイン・検査所見を記載",
				"fields":      []string{"血圧", "脈拍", "体温", "SpO2"},
			},
			"assessment": map[string]interface{}{
				"placeholder": "病状の評価・診断を記載",
			},
			"plan": map[string]interface{}{
				"placeholder": "治療方針・処方を記載",
			},
		}
		soapTemplateJSON, err := json.Marshal(soapTemplate)
		require.NoError(t, err)

		templateJSON := fmt.Sprintf(`{
			"template_name": "テスト用SOAPテンプレート",
			"template_description": "統合テスト用のテンプレート",
			"specialty": "general",
			"soap_template": %s,
			"is_system_template": false
		}`, string(soapTemplateJSON))

		resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/medical-record-templates", strings.NewReader(templateJSON))
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var template models.MedicalRecordTemplate
		DecodeJSONResponse(t, resp, &template)

		assert.NotEmpty(t, template.TemplateID)
		assert.Equal(t, "テスト用SOAPテンプレート", template.TemplateName)
		assert.False(t, template.IsSystemTemplate)

		// Test: Get the created template
		t.Run("Get template by ID", func(t *testing.T) {
			getResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/medical-record-templates/%s", template.TemplateID), nil)
			assert.Equal(t, http.StatusOK, getResp.StatusCode)

			var retrievedTemplate models.MedicalRecordTemplate
			DecodeJSONResponse(t, getResp, &retrievedTemplate)

			assert.Equal(t, template.TemplateID, retrievedTemplate.TemplateID)
			assert.Equal(t, template.TemplateName, retrievedTemplate.TemplateName)
		})

		// Cleanup: Delete the template
		t.Cleanup(func() {
			deleteResp := ts.MakeRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/medical-record-templates/%s", template.TemplateID), nil)
			if deleteResp.StatusCode != http.StatusNoContent && deleteResp.StatusCode != http.StatusNotFound {
				t.Logf("Warning: Failed to delete test template %s: status %d", template.TemplateID, deleteResp.StatusCode)
			}
			deleteResp.Body.Close()
		})
	})
}

func TestMedicalRecordTemplate_Integration_List(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create test templates
	templateIDs := make([]string, 0)
	for i := 0; i < 3; i++ {
		soapTemplate := map[string]interface{}{
			"subjective": map[string]interface{}{
				"placeholder": fmt.Sprintf("テンプレート%d", i+1),
			},
		}
		soapTemplateJSON, err := json.Marshal(soapTemplate)
		require.NoError(t, err)

		templateJSON := fmt.Sprintf(`{
			"template_name": "テストテンプレート%d",
			"specialty": "general",
			"soap_template": %s,
			"is_system_template": false
		}`, i+1, string(soapTemplateJSON))

		createResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/medical-record-templates", strings.NewReader(templateJSON))
		require.Equal(t, http.StatusCreated, createResp.StatusCode)

		var template models.MedicalRecordTemplate
		DecodeJSONResponse(t, createResp, &template)
		templateIDs = append(templateIDs, template.TemplateID)
	}

	// Cleanup
	t.Cleanup(func() {
		for _, templateID := range templateIDs {
			deleteResp := ts.MakeRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/medical-record-templates/%s", templateID), nil)
			deleteResp.Body.Close()
		}
	})

	// Test: List templates
	t.Run("List templates", func(t *testing.T) {
		listResp := ts.MakeRequest(t, http.MethodGet, "/api/v1/medical-record-templates", nil)
		assert.Equal(t, http.StatusOK, listResp.StatusCode)

		var templates []models.MedicalRecordTemplate
		DecodeJSONResponse(t, listResp, &templates)

		assert.GreaterOrEqual(t, len(templates), 3)

		// Verify all created templates are in the list
		templateIDSet := make(map[string]bool)
		for _, template := range templates {
			templateIDSet[template.TemplateID] = true
		}

		for _, templateID := range templateIDs {
			assert.True(t, templateIDSet[templateID], "Template %s should be in the list", templateID)
		}
	})
}

func TestMedicalRecordTemplate_Integration_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a template
	soapTemplate := map[string]interface{}{
		"subjective": map[string]interface{}{
			"placeholder": "テスト",
		},
	}
	soapTemplateJSON, err := json.Marshal(soapTemplate)
	require.NoError(t, err)

	templateJSON := fmt.Sprintf(`{
		"template_name": "更新前テンプレート",
		"specialty": "general",
		"soap_template": %s,
		"is_system_template": false
	}`, string(soapTemplateJSON))

	createResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/medical-record-templates", strings.NewReader(templateJSON))
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var template models.MedicalRecordTemplate
	DecodeJSONResponse(t, createResp, &template)

	// Cleanup
	t.Cleanup(func() {
		deleteResp := ts.MakeRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/medical-record-templates/%s", template.TemplateID), nil)
		deleteResp.Body.Close()
	})

	// Test: Update the template
	t.Run("Update template", func(t *testing.T) {
		updateJSON := `{
			"template_name": "更新後テンプレート",
			"template_description": "更新された説明"
		}`

		updateResp := ts.MakeRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/medical-record-templates/%s", template.TemplateID), strings.NewReader(updateJSON))
		assert.Equal(t, http.StatusOK, updateResp.StatusCode)

		var updatedTemplate models.MedicalRecordTemplate
		DecodeJSONResponse(t, updateResp, &updatedTemplate)

		assert.Equal(t, template.TemplateID, updatedTemplate.TemplateID)
		assert.Equal(t, "更新後テンプレート", updatedTemplate.TemplateName)
		assert.NotNil(t, updatedTemplate.TemplateDescription)
		assert.Equal(t, "更新された説明", *updatedTemplate.TemplateDescription)
	})
}

func TestMedicalRecordTemplate_Integration_GetBySpecialty(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create templates for different specialties
	specialties := []string{"internal_medicine", "neurology"}
	templateIDs := make([]string, 0)

	for _, specialty := range specialties {
		soapTemplate := map[string]interface{}{
			"subjective": map[string]interface{}{
				"placeholder": specialty,
			},
		}
		soapTemplateJSON, err := json.Marshal(soapTemplate)
		require.NoError(t, err)

		templateJSON := fmt.Sprintf(`{
			"template_name": "%s用テンプレート",
			"specialty": "%s",
			"soap_template": %s,
			"is_system_template": false
		}`, specialty, specialty, string(soapTemplateJSON))

		createResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/medical-record-templates", strings.NewReader(templateJSON))
		require.Equal(t, http.StatusCreated, createResp.StatusCode)

		var template models.MedicalRecordTemplate
		DecodeJSONResponse(t, createResp, &template)
		templateIDs = append(templateIDs, template.TemplateID)
	}

	// Cleanup
	t.Cleanup(func() {
		for _, templateID := range templateIDs {
			deleteResp := ts.MakeRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/medical-record-templates/%s", templateID), nil)
			deleteResp.Body.Close()
		}
	})

	// Test: Get templates by specialty
	t.Run("Get templates by specialty", func(t *testing.T) {
		getResp := ts.MakeRequest(t, http.MethodGet, "/api/v1/medical-record-templates/specialty/internal_medicine", nil)
		assert.Equal(t, http.StatusOK, getResp.StatusCode)

		var templates []models.MedicalRecordTemplate
		DecodeJSONResponse(t, getResp, &templates)

		assert.GreaterOrEqual(t, len(templates), 1)

		// Verify all templates have the correct specialty
		for _, template := range templates {
			assert.Equal(t, "internal_medicine", *template.Specialty)
		}
	})
}

func TestMedicalRecordTemplate_Integration_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a template
	soapTemplate := map[string]interface{}{
		"subjective": map[string]interface{}{
			"placeholder": "テスト",
		},
	}
	soapTemplateJSON, err := json.Marshal(soapTemplate)
	require.NoError(t, err)

	templateJSON := fmt.Sprintf(`{
		"template_name": "削除用テンプレート",
		"specialty": "general",
		"soap_template": %s,
		"is_system_template": false
	}`, string(soapTemplateJSON))

	createResp := ts.MakeRequest(t, http.MethodPost, "/api/v1/medical-record-templates", strings.NewReader(templateJSON))
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var template models.MedicalRecordTemplate
	DecodeJSONResponse(t, createResp, &template)

	// Test: Delete the template
	t.Run("Delete template", func(t *testing.T) {
		deleteResp := ts.MakeRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/medical-record-templates/%s", template.TemplateID), nil)
		assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode)

		// Verify the template is deleted (should return 404)
		getResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/medical-record-templates/%s", template.TemplateID), nil)
		assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
	})
}

func TestMedicalRecordTemplate_Integration_ValidationErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "Missing template_name",
			requestBody:    `{"specialty": "general", "soap_template": {"subjective": {}}, "is_system_template": false}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing soap_template",
			requestBody:    `{"template_name": "テスト", "specialty": "general", "is_system_template": false}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid specialty",
			requestBody:    `{"template_name": "テスト", "specialty": "invalid", "soap_template": {"subjective": {}}, "is_system_template": false}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/medical-record-templates", strings.NewReader(tt.requestBody))
			assert.Equal(t, tt.expectedStatus, resp.StatusCode, "Test case: %s", tt.name)
			resp.Body.Close()
		})
	}
}
