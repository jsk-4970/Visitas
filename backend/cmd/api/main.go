・package main

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

	// Initialize services
	patientService := services.NewPatientService(patientRepo, assignmentRepo, auditRepo)

	// Initialize middleware
	auditMiddleware := middleware.NewAuditLoggerMiddleware(auditRepo)

	// Initialize handlers
	patientHandler := handlers.NewPatientHandler(patientService)
	identifierHandler := handlers.NewIdentifierHandler(identifierRepo, patientRepo, auditMiddleware)

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
