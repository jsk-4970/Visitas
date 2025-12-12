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

func TestMedicationOrder_Integration_CreateAndGet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Test: Create a medication order
	t.Run("Create medication order", func(t *testing.T) {
		medication := map[string]interface{}{
			"code": "610432015",
			"system": "YJ",
			"display": "ロキソプロフェンNa錠60mg",
		}
		medicationJSON, err := json.Marshal(medication)
		require.NoError(t, err)

		dosageInstruction := map[string]interface{}{
			"timing": map[string]interface{}{
				"repeat": map[string]interface{}{
					"frequency": 3,
					"period": 1,
					"periodUnit": "d",
				},
			},
			"dose": map[string]interface{}{
				"value": 1,
				"unit": "tablet",
			},
			"route": "oral",
		}
		dosageJSON, err := json.Marshal(dosageInstruction)
		require.NoError(t, err)

		orderJSON := fmt.Sprintf(`{
			"status": "active",
			"intent": "order",
			"medication": %s,
			"dosage_instruction": %s,
			"prescribed_date": "%s",
			"prescribed_by": "doctor-123"
		}`, string(medicationJSON), string(dosageJSON), time.Now().Format(time.RFC3339))

		resp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/medication-orders", patientID), strings.NewReader(orderJSON))
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var order models.MedicationOrder
		DecodeJSONResponse(t, resp, &order)

		assert.NotEmpty(t, order.OrderID)
		assert.Equal(t, patientID, order.PatientID)
		assert.Equal(t, "active", order.Status)
		assert.Equal(t, "order", order.Intent)
		assert.Equal(t, "doctor-123", order.PrescribedBy)

		// Test: Get the created order
		t.Run("Get medication order by ID", func(t *testing.T) {
			getResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/patients/%s/medication-orders/%s", patientID, order.OrderID), nil)
			assert.Equal(t, http.StatusOK, getResp.StatusCode)

			var retrievedOrder models.MedicationOrder
			DecodeJSONResponse(t, getResp, &retrievedOrder)

			assert.Equal(t, order.OrderID, retrievedOrder.OrderID)
			assert.Equal(t, order.PatientID, retrievedOrder.PatientID)
			assert.Equal(t, order.Status, retrievedOrder.Status)
		})
	})
}

func TestMedicationOrder_Integration_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create a medication order
	medication := map[string]interface{}{
		"code": "610432015",
		"system": "YJ",
		"display": "ロキソプロフェンNa錠60mg",
	}
	medicationJSON, err := json.Marshal(medication)
	require.NoError(t, err)

	dosageInstruction := map[string]interface{}{
		"timing": map[string]interface{}{
			"repeat": map[string]interface{}{
				"frequency": 3,
				"period": 1,
				"periodUnit": "d",
			},
		},
		"dose": map[string]interface{}{
			"value": 1,
			"unit": "tablet",
		},
	}
	dosageJSON, err := json.Marshal(dosageInstruction)
	require.NoError(t, err)

	orderJSON := fmt.Sprintf(`{
		"status": "active",
		"intent": "order",
		"medication": %s,
		"dosage_instruction": %s,
		"prescribed_date": "%s",
		"prescribed_by": "doctor-123"
	}`, string(medicationJSON), string(dosageJSON), time.Now().Format(time.RFC3339))

	createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/medication-orders", patientID), strings.NewReader(orderJSON))
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var order models.MedicationOrder
	DecodeJSONResponse(t, createResp, &order)

	// Test: Update the order
	t.Run("Update medication order", func(t *testing.T) {
		updateJSON := `{
			"status": "completed"
		}`

		updateResp := ts.MakeRequest(t, http.MethodPut, fmt.Sprintf("/api/v1/patients/%s/medication-orders/%s", patientID, order.OrderID), strings.NewReader(updateJSON))
		assert.Equal(t, http.StatusOK, updateResp.StatusCode)

		var updatedOrder models.MedicationOrder
		DecodeJSONResponse(t, updateResp, &updatedOrder)

		assert.Equal(t, order.OrderID, updatedOrder.OrderID)
		assert.Equal(t, "completed", updatedOrder.Status)
	})
}

func TestMedicationOrder_Integration_List(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create multiple medication orders
	orderIDs := make([]string, 0)
	for i := 0; i < 3; i++ {
		medication := map[string]interface{}{
			"code": fmt.Sprintf("61043201%d", i),
			"system": "YJ",
			"display": fmt.Sprintf("Medication %d", i+1),
		}
		medicationJSON, err := json.Marshal(medication)
		require.NoError(t, err)

		dosageInstruction := map[string]interface{}{
			"timing": map[string]interface{}{
				"repeat": map[string]interface{}{
					"frequency": i + 1,
					"period": 1,
					"periodUnit": "d",
				},
			},
			"dose": map[string]interface{}{
				"value": i + 1,
				"unit": "tablet",
			},
		}
		dosageJSON, err := json.Marshal(dosageInstruction)
		require.NoError(t, err)

		orderJSON := fmt.Sprintf(`{
			"status": "active",
			"intent": "order",
			"medication": %s,
			"dosage_instruction": %s,
			"prescribed_date": "%s",
			"prescribed_by": "doctor-123"
		}`, string(medicationJSON), string(dosageJSON), time.Now().Format(time.RFC3339))

		createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/medication-orders", patientID), strings.NewReader(orderJSON))
		require.Equal(t, http.StatusCreated, createResp.StatusCode)

		var order models.MedicationOrder
		DecodeJSONResponse(t, createResp, &order)
		orderIDs = append(orderIDs, order.OrderID)
	}

	// Test: List medication orders
	t.Run("List medication orders", func(t *testing.T) {
		listResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/patients/%s/medication-orders", patientID), nil)
		assert.Equal(t, http.StatusOK, listResp.StatusCode)

		var orders []models.MedicationOrder
		DecodeJSONResponse(t, listResp, &orders)

		assert.GreaterOrEqual(t, len(orders), 3)

		// Verify all created orders are in the list
		orderIDSet := make(map[string]bool)
		for _, order := range orders {
			orderIDSet[order.OrderID] = true
		}

		for _, orderID := range orderIDs {
			assert.True(t, orderIDSet[orderID], "Order %s should be in the list", orderID)
		}
	})
}

func TestMedicationOrder_Integration_GetActiveOrders(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create an active medication order
	medication := map[string]interface{}{
		"code": "610432015",
		"system": "YJ",
		"display": "ロキソプロフェンNa錠60mg",
	}
	medicationJSON, err := json.Marshal(medication)
	require.NoError(t, err)

	dosageInstruction := map[string]interface{}{
		"timing": map[string]interface{}{
			"repeat": map[string]interface{}{
				"frequency": 3,
				"period": 1,
				"periodUnit": "d",
			},
		},
	}
	dosageJSON, err := json.Marshal(dosageInstruction)
	require.NoError(t, err)

	orderJSON := fmt.Sprintf(`{
		"status": "active",
		"intent": "order",
		"medication": %s,
		"dosage_instruction": %s,
		"prescribed_date": "%s",
		"prescribed_by": "doctor-123"
	}`, string(medicationJSON), string(dosageJSON), time.Now().Format(time.RFC3339))

	createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/medication-orders", patientID), strings.NewReader(orderJSON))
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var order models.MedicationOrder
	DecodeJSONResponse(t, createResp, &order)

	// Test: Get active orders
	t.Run("Get active medication orders", func(t *testing.T) {
		activeResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/patients/%s/medication-orders/active", patientID), nil)
		assert.Equal(t, http.StatusOK, activeResp.StatusCode)

		var activeOrders []models.MedicationOrder
		DecodeJSONResponse(t, activeResp, &activeOrders)

		assert.GreaterOrEqual(t, len(activeOrders), 1)

		// Verify the created order is in the active list
		foundOrder := false
		for _, o := range activeOrders {
			if o.OrderID == order.OrderID {
				foundOrder = true
				assert.Equal(t, "active", o.Status)
				break
			}
		}
		assert.True(t, foundOrder, "Created order should be in active orders")
	})
}

func TestMedicationOrder_Integration_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Create a medication order
	medication := map[string]interface{}{
		"code": "610432015",
		"system": "YJ",
		"display": "ロキソプロフェンNa錠60mg",
	}
	medicationJSON, err := json.Marshal(medication)
	require.NoError(t, err)

	dosageInstruction := map[string]interface{}{
		"timing": map[string]interface{}{
			"repeat": map[string]interface{}{
				"frequency": 3,
				"period": 1,
				"periodUnit": "d",
			},
		},
	}
	dosageJSON, err := json.Marshal(dosageInstruction)
	require.NoError(t, err)

	orderJSON := fmt.Sprintf(`{
		"status": "active",
		"intent": "order",
		"medication": %s,
		"dosage_instruction": %s,
		"prescribed_date": "%s",
		"prescribed_by": "doctor-123"
	}`, string(medicationJSON), string(dosageJSON), time.Now().Format(time.RFC3339))

	createResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/medication-orders", patientID), strings.NewReader(orderJSON))
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var order models.MedicationOrder
	DecodeJSONResponse(t, createResp, &order)

	// Test: Delete the order
	t.Run("Delete medication order", func(t *testing.T) {
		deleteResp := ts.MakeRequest(t, http.MethodDelete, fmt.Sprintf("/api/v1/patients/%s/medication-orders/%s", patientID, order.OrderID), nil)
		assert.Equal(t, http.StatusOK, deleteResp.StatusCode)

		// Verify the order is deleted (should return 404)
		getResp := ts.MakeRequest(t, http.MethodGet, fmt.Sprintf("/api/v1/patients/%s/medication-orders/%s", patientID, order.OrderID), nil)
		assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
	})
}

func TestMedicationOrder_Integration_WithPharmacy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	// Test: Create medication order with pharmacy info
	t.Run("Create medication order with pharmacy", func(t *testing.T) {
		medication := map[string]interface{}{
			"code": "610432015",
			"system": "YJ",
			"display": "ロキソプロフェンNa錠60mg",
		}
		medicationJSON, err := json.Marshal(medication)
		require.NoError(t, err)

		dosageInstruction := map[string]interface{}{
			"timing": map[string]interface{}{
				"repeat": map[string]interface{}{
					"frequency": 3,
					"period": 1,
					"periodUnit": "d",
				},
			},
		}
		dosageJSON, err := json.Marshal(dosageInstruction)
		require.NoError(t, err)

		pharmacy := map[string]interface{}{
			"pharmacy_id": "pharmacy-123",
			"name": "さくら薬局",
			"address": "東京都新宿区...",
			"phone": "03-1234-5678",
		}
		pharmacyJSON, err := json.Marshal(pharmacy)
		require.NoError(t, err)

		orderJSON := fmt.Sprintf(`{
			"status": "active",
			"intent": "order",
			"medication": %s,
			"dosage_instruction": %s,
			"prescribed_date": "%s",
			"prescribed_by": "doctor-123",
			"dispense_pharmacy": %s
		}`, string(medicationJSON), string(dosageJSON), time.Now().Format(time.RFC3339), string(pharmacyJSON))

		resp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/medication-orders", patientID), strings.NewReader(orderJSON))
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var order models.MedicationOrder
		DecodeJSONResponse(t, resp, &order)

		assert.NotEmpty(t, order.OrderID)
		assert.NotNil(t, order.DispensePharmacy)

		// Verify pharmacy info is stored correctly
		var storedPharmacy map[string]interface{}
		err = json.Unmarshal(order.DispensePharmacy, &storedPharmacy)
		assert.NoError(t, err)
		assert.Equal(t, "pharmacy-123", storedPharmacy["pharmacy_id"])
		assert.Equal(t, "さくら薬局", storedPharmacy["name"])
	})
}

func TestMedicationOrder_Integration_ValidationErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test server
	ts := SetupTestServer(t)
	defer ts.Close()

	// Create a test patient
	patientID := ts.CreateTestPatient(t)

	medication := map[string]interface{}{
		"code": "610432015",
		"system": "YJ",
		"display": "ロキソプロフェンNa錠60mg",
	}
	medicationJSON, err := json.Marshal(medication)
	require.NoError(t, err)

	dosageInstruction := map[string]interface{}{
		"timing": map[string]interface{}{
			"repeat": map[string]interface{}{
				"frequency": 3,
				"period": 1,
				"periodUnit": "d",
			},
		},
	}
	dosageJSON, err := json.Marshal(dosageInstruction)
	require.NoError(t, err)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "Missing status",
			requestBody:    fmt.Sprintf(`{"intent": "order", "medication": %s, "dosage_instruction": %s, "prescribed_date": "%s", "prescribed_by": "doctor-123"}`, string(medicationJSON), string(dosageJSON), time.Now().Format(time.RFC3339)),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid status",
			requestBody:    fmt.Sprintf(`{"status": "invalid", "intent": "order", "medication": %s, "dosage_instruction": %s, "prescribed_date": "%s", "prescribed_by": "doctor-123"}`, string(medicationJSON), string(dosageJSON), time.Now().Format(time.RFC3339)),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing intent",
			requestBody:    fmt.Sprintf(`{"status": "active", "medication": %s, "dosage_instruction": %s, "prescribed_date": "%s", "prescribed_by": "doctor-123"}`, string(medicationJSON), string(dosageJSON), time.Now().Format(time.RFC3339)),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing medication",
			requestBody:    fmt.Sprintf(`{"status": "active", "intent": "order", "dosage_instruction": %s, "prescribed_date": "%s", "prescribed_by": "doctor-123"}`, string(dosageJSON), time.Now().Format(time.RFC3339)),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing dosage_instruction",
			requestBody:    fmt.Sprintf(`{"status": "active", "intent": "order", "medication": %s, "prescribed_date": "%s", "prescribed_by": "doctor-123"}`, string(medicationJSON), time.Now().Format(time.RFC3339)),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing prescribed_date",
			requestBody:    fmt.Sprintf(`{"status": "active", "intent": "order", "medication": %s, "dosage_instruction": %s, "prescribed_by": "doctor-123"}`, string(medicationJSON), string(dosageJSON)),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing prescribed_by",
			requestBody:    fmt.Sprintf(`{"status": "active", "intent": "order", "medication": %s, "dosage_instruction": %s, "prescribed_date": "%s"}`, string(medicationJSON), string(dosageJSON), time.Now().Format(time.RFC3339)),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/medication-orders", patientID), strings.NewReader(tt.requestBody))
			assert.Equal(t, tt.expectedStatus, resp.StatusCode, "Test case: %s", tt.name)
			resp.Body.Close()
		})
	}
}
