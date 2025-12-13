package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/require"
	"github.com/visitas/backend/internal/config"
	"github.com/visitas/backend/internal/handlers"
	"github.com/visitas/backend/internal/middleware"
	"github.com/visitas/backend/internal/repository"
	"github.com/visitas/backend/internal/services"
)

// TestConfig holds configuration for integration tests
type TestConfig struct {
	SpannerRepo               *repository.SpannerRepository
	PatientRepo               *repository.PatientRepository
	VisitScheduleRepo         *repository.VisitScheduleRepository
	ClinicalObservationRepo   *repository.ClinicalObservationRepository
	CarePlanRepo              *repository.CarePlanRepository
	MedicationOrderRepo       *repository.MedicationOrderRepository
	ACPRecordRepo             *repository.ACPRecordRepository
	MedicalRecordRepo         *repository.MedicalRecordRepository
	MedicalRecordTemplateRepo *repository.MedicalRecordTemplateRepository
	AuditRepo                 *repository.AuditRepository
}

// TestServer wraps the test HTTP server and related resources
type TestServer struct {
	Server  *httptest.Server
	Router  chi.Router
	Config  *TestConfig
	Context context.Context
}

// SetupTestServer creates a test server with all routes configured
func SetupTestServer(t *testing.T) *TestServer {
	t.Helper()

	ctx := context.Background()

	// Load configuration (use test environment)
	cfg, err := loadTestConfig()
	require.NoError(t, err, "Failed to load test config")

	// Initialize Spanner repository
	spannerRepo, err := repository.NewSpannerRepository(ctx, cfg)
	require.NoError(t, err, "Failed to initialize Spanner repository")

	// Initialize repositories
	patientRepo := repository.NewPatientRepository(spannerRepo)
	assignmentRepo := repository.NewAssignmentRepository(spannerRepo)
	auditRepo := repository.NewAuditRepository(spannerRepo)
	visitScheduleRepo := repository.NewVisitScheduleRepository(spannerRepo)
	clinicalObservationRepo := repository.NewClinicalObservationRepository(spannerRepo)
	carePlanRepo := repository.NewCarePlanRepository(spannerRepo)
	medicationOrderRepo := repository.NewMedicationOrderRepository(spannerRepo)
	acpRecordRepo := repository.NewACPRecordRepository(spannerRepo)
	medicalRecordRepo := repository.NewMedicalRecordRepository(spannerRepo)
	medicalRecordTemplateRepo := repository.NewMedicalRecordTemplateRepository(spannerRepo)

	// Initialize services
	patientService := services.NewPatientService(patientRepo, assignmentRepo, auditRepo)
	visitScheduleService := services.NewVisitScheduleService(visitScheduleRepo, patientRepo)
	clinicalObservationService := services.NewClinicalObservationService(clinicalObservationRepo, patientRepo)
	carePlanService := services.NewCarePlanService(carePlanRepo, patientRepo)
	medicationOrderService := services.NewMedicationOrderService(medicationOrderRepo, patientRepo)
	acpRecordService := services.NewACPRecordService(acpRecordRepo, patientRepo)
	medicalRecordService := services.NewMedicalRecordService(medicalRecordRepo, patientRepo, medicalRecordTemplateRepo)
	medicalRecordTemplateService := services.NewMedicalRecordTemplateService(medicalRecordTemplateRepo)

	// Initialize handlers
	patientHandler := handlers.NewPatientHandler(patientService)
	visitScheduleHandler := handlers.NewVisitScheduleHandler(visitScheduleService)
	clinicalObservationHandler := handlers.NewClinicalObservationHandler(clinicalObservationService)
	carePlanHandler := handlers.NewCarePlanHandler(carePlanService)
	medicationOrderHandler := handlers.NewMedicationOrderHandler(medicationOrderService)
	acpRecordHandler := handlers.NewACPRecordHandler(acpRecordService)
	medicalRecordHandler := handlers.NewMedicalRecordHandler(medicalRecordService)
	medicalRecordTemplateHandler := handlers.NewMedicalRecordTemplateHandler(medicalRecordTemplateService)

	// Initialize middleware
	auditMiddleware := middleware.NewAuditLoggerMiddleware(auditRepo)

	// Setup router
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy"}`))
	})

	// API routes (without authentication for testing)
	r.Route("/api/v1", func(r chi.Router) {
		// Add test user context middleware (bypass real authentication)
		r.Use(testAuthMiddleware)
		r.Use(auditMiddleware.LogPatientAccess)

		// Patient routes
		r.Route("/patients", func(r chi.Router) {
			r.Get("/", patientHandler.GetMyPatients)
			r.Post("/", patientHandler.CreatePatient)
			r.Get("/{id}", patientHandler.GetPatient)
			r.Put("/{id}", patientHandler.UpdatePatient)
			r.Delete("/{id}", patientHandler.DeletePatient)
			r.Post("/{id}/assign", patientHandler.AssignPatientToStaff)
		})

		// Visit schedule routes
		r.Route("/patients/{patient_id}/schedules", func(r chi.Router) {
			r.Get("/", visitScheduleHandler.GetVisitSchedules)
			r.Post("/", visitScheduleHandler.CreateVisitSchedule)
			r.Get("/upcoming", visitScheduleHandler.GetUpcomingSchedules)
			r.Get("/{id}", visitScheduleHandler.GetVisitSchedule)
			r.Put("/{id}", visitScheduleHandler.UpdateVisitSchedule)
			r.Delete("/{id}", visitScheduleHandler.DeleteVisitSchedule)
			r.Post("/{id}/assign-staff", visitScheduleHandler.AssignStaff)
			r.Post("/{id}/status", visitScheduleHandler.UpdateStatus)
		})

		// Clinical observation routes
		r.Route("/patients/{patient_id}/observations", func(r chi.Router) {
			r.Get("/", clinicalObservationHandler.GetClinicalObservations)
			r.Post("/", clinicalObservationHandler.CreateClinicalObservation)
			r.Get("/latest/{category}", clinicalObservationHandler.GetLatestObservation)
			r.Get("/timeseries/{category}", clinicalObservationHandler.GetTimeSeriesData)
			r.Get("/{id}", clinicalObservationHandler.GetClinicalObservation)
			r.Put("/{id}", clinicalObservationHandler.UpdateClinicalObservation)
			r.Delete("/{id}", clinicalObservationHandler.DeleteClinicalObservation)
		})

		// Care plan routes
		r.Route("/patients/{patient_id}/care-plans", func(r chi.Router) {
			r.Get("/", carePlanHandler.GetCarePlans)
			r.Post("/", carePlanHandler.CreateCarePlan)
			r.Get("/active", carePlanHandler.GetActiveCarePlans)
			r.Get("/{id}", carePlanHandler.GetCarePlan)
			r.Put("/{id}", carePlanHandler.UpdateCarePlan)
			r.Delete("/{id}", carePlanHandler.DeleteCarePlan)
		})

		// Medication order routes
		r.Route("/patients/{patient_id}/medication-orders", func(r chi.Router) {
			r.Get("/", medicationOrderHandler.GetMedicationOrders)
			r.Post("/", medicationOrderHandler.CreateMedicationOrder)
			r.Get("/active", medicationOrderHandler.GetActiveOrders)
			r.Get("/{id}", medicationOrderHandler.GetMedicationOrder)
			r.Put("/{id}", medicationOrderHandler.UpdateMedicationOrder)
			r.Delete("/{id}", medicationOrderHandler.DeleteMedicationOrder)
		})

		// ACP record routes
		r.Route("/patients/{patient_id}/acp-records", func(r chi.Router) {
			r.Get("/", acpRecordHandler.GetACPRecords)
			r.Post("/", acpRecordHandler.CreateACPRecord)
			r.Get("/latest", acpRecordHandler.GetLatestACP)
			r.Get("/history", acpRecordHandler.GetACPHistory)
			r.Get("/{id}", acpRecordHandler.GetACPRecord)
			r.Put("/{id}", acpRecordHandler.UpdateACPRecord)
			r.Delete("/{id}", acpRecordHandler.DeleteACPRecord)
		})

		// Medical record routes
		r.Route("/patients/{patient_id}/medical-records", func(r chi.Router) {
			r.Get("/", medicalRecordHandler.ListMedicalRecords)
			r.Post("/", medicalRecordHandler.CreateMedicalRecord)
			r.Post("/from-template", medicalRecordHandler.CreateFromTemplate)
			r.Get("/latest", medicalRecordHandler.GetLatestRecords)
			r.Get("/{id}", medicalRecordHandler.GetMedicalRecord)
			r.Put("/{id}", medicalRecordHandler.UpdateMedicalRecord)
			r.Delete("/{id}", medicalRecordHandler.DeleteMedicalRecord)
		})

		// Medical record copy route
		r.Route("/medical-records/{record_id}", func(r chi.Router) {
			r.Post("/copy", medicalRecordHandler.CopyMedicalRecord)
		})

		// Medical record drafts route
		r.Get("/medical-records/drafts", medicalRecordHandler.GetDraftRecords)

		// Medical record template routes
		r.Route("/medical-record-templates", func(r chi.Router) {
			r.Get("/", medicalRecordTemplateHandler.ListTemplates)
			r.Post("/", medicalRecordTemplateHandler.CreateTemplate)
			r.Get("/system", medicalRecordTemplateHandler.GetSystemTemplates)
			r.Get("/specialty/{specialty}", medicalRecordTemplateHandler.GetTemplatesBySpecialty)
			r.Get("/{id}", medicalRecordTemplateHandler.GetTemplate)
			r.Put("/{id}", medicalRecordTemplateHandler.UpdateTemplate)
			r.Delete("/{id}", medicalRecordTemplateHandler.DeleteTemplate)
		})
	})

	// Create test server
	server := httptest.NewServer(r)

	testConfig := &TestConfig{
		SpannerRepo:               spannerRepo,
		PatientRepo:               patientRepo,
		VisitScheduleRepo:         visitScheduleRepo,
		ClinicalObservationRepo:   clinicalObservationRepo,
		CarePlanRepo:              carePlanRepo,
		MedicationOrderRepo:       medicationOrderRepo,
		ACPRecordRepo:             acpRecordRepo,
		MedicalRecordRepo:         medicalRecordRepo,
		MedicalRecordTemplateRepo: medicalRecordTemplateRepo,
		AuditRepo:                 auditRepo,
	}

	return &TestServer{
		Server:  server,
		Router:  r,
		Config:  testConfig,
		Context: ctx,
	}
}

// Close cleans up the test server and related resources
func (ts *TestServer) Close() {
	ts.Server.Close()
	if ts.Config.SpannerRepo != nil {
		ts.Config.SpannerRepo.Close()
	}
}

// testAuthMiddleware adds a test user ID to the context (bypassing real authentication)
func testAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add test user ID to context
		ctx := context.WithValue(r.Context(), middleware.UserIDContextKey, "test-staff-id")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// loadTestConfig loads test configuration from environment or defaults
func loadTestConfig() (*config.Config, error) {
	// Use test database if available, otherwise use default
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		projectID = "stunning-grin-480914-n1"
	}

	instance := os.Getenv("SPANNER_INSTANCE")
	if instance == "" {
		instance = "stunning-grin-480914-n1-instance"
	}

	database := os.Getenv("SPANNER_DATABASE")
	if database == "" {
		database = "stunning-grin-480914-n1-db"
	}

	// Set emulator host for tests
	emulatorHost := os.Getenv("SPANNER_EMULATOR_HOST")
	if emulatorHost == "" {
		emulatorHost = "localhost:9010"
		os.Setenv("SPANNER_EMULATOR_HOST", emulatorHost)
	}

	return &config.Config{
		ProjectID:        projectID,
		SpannerInstance:  instance,
		SpannerDatabase:  database,
		SpannerEmulator:  emulatorHost,
		Port:             "8080",
		Env:              "test",
		AllowedOrigins:   []string{"*"},
		LogLevel:         "info",
	}, nil
}

// MakeRequest is a helper function to make HTTP requests to the test server
func (ts *TestServer) MakeRequest(t *testing.T, method, path string, body io.Reader) *http.Response {
	t.Helper()

	req, err := http.NewRequest(method, ts.Server.URL+path, body)
	require.NoError(t, err, "Failed to create request")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "Failed to execute request")

	return resp
}

// DecodeJSONResponse decodes a JSON response into the provided interface
func DecodeJSONResponse(t *testing.T, resp *http.Response, v interface{}) {
	t.Helper()

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")

	err = json.Unmarshal(body, v)
	require.NoError(t, err, fmt.Sprintf("Failed to decode JSON response: %s", string(body)))
}

// CreateTestPatient creates a test patient and returns the patient ID
func (ts *TestServer) CreateTestPatient(t *testing.T) string {
	t.Helper()

	patientJSON := `{
		"birth_date": "1950-01-15",
		"gender": "male",
		"name": {
			"family": "山田",
			"given": "太郎",
			"kana": "ヤマダ タロウ"
		},
		"contact_points": [
			{
				"system": "phone",
				"value": "03-1234-5678",
				"use": "home"
			}
		],
		"addresses": [
			{
				"use": "home",
				"type": "postal",
				"text": "東京都新宿区西新宿2-8-1",
				"postal_code": "160-0023",
				"country": "JP"
			}
		],
		"consent_status": "obtained"
	}`

	resp := ts.MakeRequest(t, http.MethodPost, "/api/v1/patients", strings.NewReader(patientJSON))
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to create test patient")

	var patient map[string]interface{}
	DecodeJSONResponse(t, resp, &patient)

	patientID, ok := patient["patient_id"].(string)
	require.True(t, ok, "patient_id not found in response")
	require.NotEmpty(t, patientID, "patient_id is empty")

	// Assign the test patient to the test staff
	assignJSON := `{"staff_id": "test-staff-id", "role": "doctor", "assignment_type": "primary"}`
	assignResp := ts.MakeRequest(t, http.MethodPost, fmt.Sprintf("/api/v1/patients/%s/assign", patientID), strings.NewReader(assignJSON))
	require.Equal(t, http.StatusOK, assignResp.StatusCode, "Failed to assign patient to staff")

	// Clean up on test completion
	t.Cleanup(func() {
		ts.DeleteTestPatient(t, patientID)
	})

	return patientID
}

// DeleteTestPatient deletes a test patient
func (ts *TestServer) DeleteTestPatient(t *testing.T, patientID string) {
	t.Helper()

	// Try to make the request, but don't fail if server is already closed
	req, err := http.NewRequest(http.MethodDelete, ts.Server.URL+fmt.Sprintf("/api/v1/patients/%s", patientID), nil)
	if err != nil {
		t.Logf("Warning: Failed to create delete request for patient %s: %v", patientID, err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// Server might be already shut down, which is OK during cleanup
		t.Logf("Info: Could not delete test patient %s (server may be closed): %v", patientID, err)
		return
	}
	defer resp.Body.Close()

	// It's OK if deletion fails (e.g., patient already deleted or doesn't exist)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		t.Logf("Warning: Failed to delete test patient %s: status %d", patientID, resp.StatusCode)
	}
}

// CleanupTestData removes all test data from the database
func (ts *TestServer) CleanupTestData(t *testing.T) {
	t.Helper()
	// This function can be extended to clean up specific test data if needed
	// For now, we rely on individual test cleanup
}
