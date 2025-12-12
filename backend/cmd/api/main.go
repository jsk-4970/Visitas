package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/visitas/backend/internal/config"
	"github.com/visitas/backend/internal/handlers"
	"github.com/visitas/backend/internal/middleware"
	"github.com/visitas/backend/internal/repository"
	"github.com/visitas/backend/internal/services"
	"github.com/visitas/backend/pkg/auth"
	"github.com/visitas/backend/pkg/encryption"
	"github.com/visitas/backend/pkg/logger"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logger.Info("No .env file found, using environment variables")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config", err)
	}

	// Set logger level
	if cfg.LogLevel == "debug" {
		logger.SetGlobalLevel(logger.LogLevelDebug)
	}

	// Initialize context
	ctx := context.Background()

	// Initialize Firebase client
	var firebaseClient *auth.FirebaseClient
	var authMiddleware *middleware.AuthMiddleware

	if cfg.FirebaseConfigPath != "" {
		firebaseClient, err = auth.NewFirebaseClient(ctx, cfg.FirebaseConfigPath)
		if err != nil {
			logger.Warn("Failed to initialize Firebase client - authentication will be disabled", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			authMiddleware = middleware.NewAuthMiddleware(firebaseClient)
			logger.Info("Firebase Authentication initialized successfully")
		}
	} else {
		logger.Warn("FIREBASE_CONFIG_PATH not set - authentication will be disabled")
	}

	// Initialize Spanner repository
	spannerRepo, err := repository.NewSpannerRepository(ctx, cfg)
	if err != nil {
		logger.Fatal("Failed to initialize Spanner repository", err)
	}
	defer spannerRepo.Close()
	logger.Info("Spanner repository initialized successfully")

	// Initialize KMS encryptor for My Number encryption
	var kmsEncryptor *encryption.KMSEncryptor
	kmsProjectID := os.Getenv("KMS_PROJECT_ID")
	kmsLocation := os.Getenv("KMS_LOCATION")
	kmsKeyRing := os.Getenv("KMS_KEYRING")
	kmsKey := os.Getenv("KMS_KEY")

	if kmsProjectID != "" && kmsLocation != "" && kmsKeyRing != "" && kmsKey != "" {
		kmsEncryptor, err = encryption.NewKMSEncryptor(ctx, kmsProjectID, kmsLocation, kmsKeyRing, kmsKey)
		if err != nil {
			logger.Warn("Failed to initialize KMS encryptor - My Number encryption will not be available", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			logger.Info("KMS encryptor initialized successfully")
		}
	} else {
		logger.Warn("KMS configuration not set - My Number encryption will not be available")
	}

	// Initialize repositories
	patientRepo := repository.NewPatientRepository(spannerRepo)
	identifierRepo := repository.NewIdentifierRepository(spannerRepo, kmsEncryptor)
	assignmentRepo := repository.NewAssignmentRepository(spannerRepo)
	auditRepo := repository.NewAuditRepository(spannerRepo)
	socialProfileRepo := repository.NewSocialProfileRepository(spannerRepo)
	coverageRepo := repository.NewCoverageRepository(spannerRepo)
	medicalConditionRepo := repository.NewMedicalConditionRepository(spannerRepo)
	allergyIntoleranceRepo := repository.NewAllergyIntoleranceRepository(spannerRepo)
	visitScheduleRepo := repository.NewVisitScheduleRepository(spannerRepo)
	clinicalObservationRepo := repository.NewClinicalObservationRepository(spannerRepo)
	carePlanRepo := repository.NewCarePlanRepository(spannerRepo)
	medicationOrderRepo := repository.NewMedicationOrderRepository(spannerRepo)
	acpRecordRepo := repository.NewACPRecordRepository(spannerRepo)
	medicalRecordRepo := repository.NewMedicalRecordRepository(spannerRepo)
	medicalRecordTemplateRepo := repository.NewMedicalRecordTemplateRepository(spannerRepo)

	// Initialize services
	patientService := services.NewPatientService(patientRepo, assignmentRepo, auditRepo)
	identifierService := services.NewIdentifierService(identifierRepo, patientRepo, auditRepo)
	medicalConditionService := services.NewMedicalConditionService(medicalConditionRepo, patientRepo)
	allergyIntoleranceService := services.NewAllergyIntoleranceService(allergyIntoleranceRepo, patientRepo)
	socialProfileService := services.NewSocialProfileService(socialProfileRepo, patientRepo)
	coverageService := services.NewCoverageService(coverageRepo, patientRepo)
	visitScheduleService := services.NewVisitScheduleService(visitScheduleRepo, patientRepo)
	clinicalObservationService := services.NewClinicalObservationService(clinicalObservationRepo, patientRepo)
	carePlanService := services.NewCarePlanService(carePlanRepo, patientRepo)
	medicationOrderService := services.NewMedicationOrderService(medicationOrderRepo, patientRepo)
	acpRecordService := services.NewACPRecordService(acpRecordRepo, patientRepo)
	medicalRecordTemplateService := services.NewMedicalRecordTemplateService(medicalRecordTemplateRepo)
	medicalRecordService := services.NewMedicalRecordService(medicalRecordRepo, patientRepo, medicalRecordTemplateRepo)

	// Initialize middleware
	auditMiddleware := middleware.NewAuditLoggerMiddleware(auditRepo)

	// Initialize handlers
	patientHandler := handlers.NewPatientHandler(patientService)
	identifierHandler := handlers.NewIdentifierHandler(identifierService)
	socialProfileHandler := handlers.NewSocialProfileHandler(socialProfileService)
	coverageHandler := handlers.NewCoverageHandler(coverageService)
	medicalConditionHandler := handlers.NewMedicalConditionHandler(medicalConditionService)
	allergyIntoleranceHandler := handlers.NewAllergyIntoleranceHandler(allergyIntoleranceService)
	visitScheduleHandler := handlers.NewVisitScheduleHandler(visitScheduleService)
	clinicalObservationHandler := handlers.NewClinicalObservationHandler(clinicalObservationService)
	carePlanHandler := handlers.NewCarePlanHandler(carePlanService)
	medicationOrderHandler := handlers.NewMedicationOrderHandler(medicationOrderService)
	acpRecordHandler := handlers.NewACPRecordHandler(acpRecordService)
	medicalRecordHandler := handlers.NewMedicalRecordHandler(medicalRecordService)
	medicalRecordTemplateHandler := handlers.NewMedicalRecordTemplateHandler(medicalRecordTemplateService)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Routes
	// Health check endpoint (public, no auth required)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Apply authentication middleware if Firebase is configured
		if authMiddleware != nil {
			r.Use(authMiddleware.RequireAuth)
		}

		// Apply audit logging middleware
		r.Use(auditMiddleware.LogPatientAccess)

		// Patient routes (protected)
		r.Route("/patients", func(r chi.Router) {
			r.Get("/", patientHandler.GetMyPatients)        // List my assigned patients
			r.Post("/", patientHandler.CreatePatient)       // Create patient
			r.Get("/{id}", patientHandler.GetPatient)       // Get patient by ID
			r.Put("/{id}", patientHandler.UpdatePatient)    // Update patient
			r.Delete("/{id}", patientHandler.DeletePatient) // Delete patient (soft delete)
			r.Post("/{id}/assign", patientHandler.AssignPatientToStaff) // Assign patient to staff
		})

		// Patient identifier routes (protected)
		r.Route("/patients/{patient_id}/identifiers", func(r chi.Router) {
			r.Get("/", identifierHandler.GetIdentifiers)       // List identifiers
			r.Post("/", identifierHandler.CreateIdentifier)    // Create identifier
			r.Get("/{id}", identifierHandler.GetIdentifier)    // Get identifier by ID
			r.Put("/{id}", identifierHandler.UpdateIdentifier) // Update identifier
			r.Delete("/{id}", identifierHandler.DeleteIdentifier) // Delete identifier
		})

		// Social profile routes (protected)
		r.Route("/patients/{patient_id}/social-profiles", func(r chi.Router) {
			r.Get("/", socialProfileHandler.GetSocialProfiles)       // List social profiles
			r.Post("/", socialProfileHandler.CreateSocialProfile)    // Create social profile
			r.Get("/{id}", socialProfileHandler.GetSocialProfile)    // Get social profile by ID
			r.Put("/{id}", socialProfileHandler.UpdateSocialProfile) // Update social profile
			r.Delete("/{id}", socialProfileHandler.DeleteSocialProfile) // Delete social profile
		})

		// Coverage routes (protected)
		r.Route("/patients/{patient_id}/coverages", func(r chi.Router) {
			r.Get("/", coverageHandler.GetCoverages)       // List coverages
			r.Post("/", coverageHandler.CreateCoverage)    // Create coverage
			r.Get("/{id}", coverageHandler.GetCoverage)    // Get coverage by ID
			r.Put("/{id}", coverageHandler.UpdateCoverage) // Update coverage
			r.Delete("/{id}", coverageHandler.DeleteCoverage) // Delete coverage
			r.Post("/{id}/verify", coverageHandler.VerifyCoverage) // Verify coverage
		})

		// Medical condition routes (protected)
		r.Route("/patients/{patient_id}/conditions", func(r chi.Router) {
			r.Get("/", medicalConditionHandler.GetMedicalConditions)       // List medical conditions
			r.Post("/", medicalConditionHandler.CreateMedicalCondition)    // Create medical condition
			r.Get("/{id}", medicalConditionHandler.GetMedicalCondition)    // Get medical condition by ID
			r.Put("/{id}", medicalConditionHandler.UpdateMedicalCondition) // Update medical condition
			r.Delete("/{id}", medicalConditionHandler.DeleteMedicalCondition) // Delete medical condition
		})

		// Allergy intolerance routes (protected)
		r.Route("/patients/{patient_id}/allergies", func(r chi.Router) {
			r.Get("/", allergyIntoleranceHandler.GetAllergyIntolerances)       // List allergy intolerances
			r.Post("/", allergyIntoleranceHandler.CreateAllergyIntolerance)    // Create allergy intolerance
			r.Get("/{id}", allergyIntoleranceHandler.GetAllergyIntolerance)    // Get allergy intolerance by ID
			r.Put("/{id}", allergyIntoleranceHandler.UpdateAllergyIntolerance) // Update allergy intolerance
			r.Delete("/{id}", allergyIntoleranceHandler.DeleteAllergyIntolerance) // Delete allergy intolerance
		})

		// Visit schedule routes (protected)
		r.Route("/patients/{patient_id}/schedules", func(r chi.Router) {
			r.Get("/", visitScheduleHandler.GetVisitSchedules)       // List visit schedules
			r.Post("/", visitScheduleHandler.CreateVisitSchedule)    // Create visit schedule
			r.Get("/upcoming", visitScheduleHandler.GetUpcomingSchedules) // Get upcoming schedules
			r.Get("/{id}", visitScheduleHandler.GetVisitSchedule)    // Get visit schedule by ID
			r.Put("/{id}", visitScheduleHandler.UpdateVisitSchedule) // Update visit schedule
			r.Delete("/{id}", visitScheduleHandler.DeleteVisitSchedule) // Delete visit schedule
			r.Post("/{id}/assign-staff", visitScheduleHandler.AssignStaff) // Assign staff to schedule
			r.Post("/{id}/status", visitScheduleHandler.UpdateStatus) // Update schedule status
		})

		// Clinical observation routes (protected)
		r.Route("/patients/{patient_id}/observations", func(r chi.Router) {
			r.Get("/", clinicalObservationHandler.GetClinicalObservations)       // List clinical observations
			r.Post("/", clinicalObservationHandler.CreateClinicalObservation)    // Create clinical observation
			r.Get("/latest/{category}", clinicalObservationHandler.GetLatestObservation) // Get latest observation by category
			r.Get("/timeseries/{category}", clinicalObservationHandler.GetTimeSeriesData) // Get time series data
			r.Get("/{id}", clinicalObservationHandler.GetClinicalObservation)    // Get clinical observation by ID
			r.Put("/{id}", clinicalObservationHandler.UpdateClinicalObservation) // Update clinical observation
			r.Delete("/{id}", clinicalObservationHandler.DeleteClinicalObservation) // Delete clinical observation
		})

		// Care plan routes (protected)
		r.Route("/patients/{patient_id}/care-plans", func(r chi.Router) {
			r.Get("/", carePlanHandler.GetCarePlans)       // List care plans
			r.Post("/", carePlanHandler.CreateCarePlan)    // Create care plan
			r.Get("/active", carePlanHandler.GetActiveCarePlans) // Get active care plans
			r.Get("/{id}", carePlanHandler.GetCarePlan)    // Get care plan by ID
			r.Put("/{id}", carePlanHandler.UpdateCarePlan) // Update care plan
			r.Delete("/{id}", carePlanHandler.DeleteCarePlan) // Delete care plan
		})

		// Medication order routes (protected)
		r.Route("/patients/{patient_id}/medication-orders", func(r chi.Router) {
			r.Get("/", medicationOrderHandler.GetMedicationOrders)       // List medication orders
			r.Post("/", medicationOrderHandler.CreateMedicationOrder)    // Create medication order
			r.Get("/active", medicationOrderHandler.GetActiveOrders)     // Get active medication orders
			r.Get("/{id}", medicationOrderHandler.GetMedicationOrder)    // Get medication order by ID
			r.Put("/{id}", medicationOrderHandler.UpdateMedicationOrder) // Update medication order
			r.Delete("/{id}", medicationOrderHandler.DeleteMedicationOrder) // Delete medication order
		})

		// ACP record routes (protected)
		r.Route("/patients/{patient_id}/acp-records", func(r chi.Router) {
			r.Get("/", acpRecordHandler.GetACPRecords)         // List ACP records
			r.Post("/", acpRecordHandler.CreateACPRecord)      // Create ACP record
			r.Get("/latest", acpRecordHandler.GetLatestACP)    // Get latest active ACP
			r.Get("/history", acpRecordHandler.GetACPHistory)  // Get complete ACP history
			r.Get("/{id}", acpRecordHandler.GetACPRecord)      // Get ACP record by ID
			r.Put("/{id}", acpRecordHandler.UpdateACPRecord)   // Update ACP record
			r.Delete("/{id}", acpRecordHandler.DeleteACPRecord) // Delete ACP record
		})

		// Medical record routes (protected) - Phase 1 Sprint 6: 基本カルテ機能
		r.Route("/patients/{patient_id}/medical-records", func(r chi.Router) {
			r.Get("/", medicalRecordHandler.ListMedicalRecords)           // List medical records
			r.Post("/", medicalRecordHandler.CreateMedicalRecord)         // Create medical record
			r.Get("/latest", medicalRecordHandler.GetLatestRecords)       // Get latest records
			r.Post("/from-template", medicalRecordHandler.CreateFromTemplate) // Create from template
			r.Get("/{id}", medicalRecordHandler.GetMedicalRecord)         // Get medical record by ID
			r.Put("/{id}", medicalRecordHandler.UpdateMedicalRecord)      // Update medical record
			r.Delete("/{id}", medicalRecordHandler.DeleteMedicalRecord)   // Delete medical record
		})

		// Medical record copy route (protected)
		r.Route("/medical-records/{record_id}", func(r chi.Router) {
			r.Post("/copy", medicalRecordHandler.CopyMedicalRecord) // Copy medical record
		})

		// Draft records route (protected)
		r.Get("/medical-records/drafts", medicalRecordHandler.GetDraftRecords) // Get my draft records

		// Medical record template routes (protected)
		r.Route("/medical-record-templates", func(r chi.Router) {
			r.Get("/", medicalRecordTemplateHandler.ListTemplates)              // List templates
			r.Post("/", medicalRecordTemplateHandler.CreateTemplate)            // Create template
			r.Get("/system", medicalRecordTemplateHandler.GetSystemTemplates)   // Get system templates
			r.Get("/specialty/{specialty}", medicalRecordTemplateHandler.GetTemplatesBySpecialty) // Get by specialty
			r.Get("/{id}", medicalRecordTemplateHandler.GetTemplate)            // Get template by ID
			r.Put("/{id}", medicalRecordTemplateHandler.UpdateTemplate)         // Update template
			r.Delete("/{id}", medicalRecordTemplateHandler.DeleteTemplate)      // Delete template
		})
	})

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		logger.Info("Starting server", map[string]interface{}{
			"addr": addr,
			"env":  cfg.Env,
		})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Server forced to shutdown", err)
	}

	// Close KMS encryptor if initialized
	if kmsEncryptor != nil {
		kmsEncryptor.Close()
	}

	logger.Info("Server exited")
}
