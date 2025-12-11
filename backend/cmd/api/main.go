・package main

import (
	"context"
	"fmt"
	"log"
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
	"github.com/visitas/backend/pkg/auth"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize context
	ctx := context.Background()

	// Initialize Firebase client
	var firebaseClient *auth.FirebaseClient
	var authMiddleware *middleware.AuthMiddleware

	if cfg.FirebaseConfigPath != "" {
		firebaseClient, err = auth.NewFirebaseClient(ctx, cfg.FirebaseConfigPath)
		if err != nil {
			log.Printf("Warning: Failed to initialize Firebase client: %v", err)
			log.Println("Authentication will be disabled")
		} else {
			authMiddleware = middleware.NewAuthMiddleware(firebaseClient)
			log.Println("Firebase Authentication initialized successfully")
		}
	} else {
		log.Println("Warning: FIREBASE_CONFIG_PATH not set, authentication will be disabled")
	}

	// Initialize Spanner client
	spannerRepo, err := repository.NewSpannerRepository(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Spanner repository: %v", err)
	}
	defer spannerRepo.Close()

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	patientHandler := handlers.NewPatientHandler(spannerRepo)

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
	r.Get("/health", healthHandler.Check)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Apply authentication middleware if Firebase is configured
		if authMiddleware != nil {
			r.Use(authMiddleware.RequireAuth)
		}

		// Patient routes (protected)
		r.Route("/patients", func(r chi.Router) {
			r.Get("/", patientHandler.List)
			r.Post("/", patientHandler.Create)
			r.Get("/{id}", patientHandler.Get)
			r.Put("/{id}", patientHandler.Update)
			r.Delete("/{id}", patientHandler.Delete)
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
		log.Printf("Starting server on %s (env: %s)", addr, cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
