package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"auth-service/internal/handler"
	"auth-service/internal/middleware"
	"auth-service/internal/repository"
	"auth-service/internal/usecase"
	"auth-service/pkg/config"
	"auth-service/pkg/logger"

	_ "github.com/lib/pq"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	log := logger.NewLogger()

	log.Info("Starting Auth Service...")
	log.Info("Environment: %s", cfg.Environment)

	// Connect to database
	db, err := connectDB(cfg, log)
	if err != nil {
		log.Error("Failed to connect to database: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	log.Info("Connected to database")

	// Initialize repositories
	tenantRepo := repository.NewTenantRepository(db)
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	// permissionRepo := repository.NewPermissionRepository(postgresDB)
	userRoleRepo := repository.NewUserRoleRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	auditLogRepo := repository.NewAuditLogRepository(db)

	// Initialize JWT service
	jwtService := usecase.NewJWTService(cfg.JWTSecret, cfg.JWTExpiry)

	// Initialize use cases
	authUseCase := usecase.NewAuthUseCase(
		userRepo,
		tenantRepo,
		roleRepo,
		userRoleRepo,
		refreshTokenRepo,
		auditLogRepo,
		jwtService,
		cfg.RefreshTokenExpiry,
	)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authUseCase, jwtService)

	// Setup routes
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", authHandler.Health)

	// Auth endpoints
	mux.HandleFunc("/login", authHandler.Login)
	mux.HandleFunc("/refresh", authHandler.RefreshToken)

	// Tenant endpoints (SUPER_ADMIN only)
	mux.HandleFunc("/tenants", authHandler.CreateTenant)

	// User endpoints
	mux.HandleFunc("/users", authHandler.CreateUser)

	// Apply middleware
	var handlerFunc http.Handler = mux
	handlerFunc = middleware.CORS(handlerFunc)
	handlerFunc = middleware.ContentTypeJSON(handlerFunc)
	handlerFunc = middleware.Logger(log)(handlerFunc)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Info("Server listening on %s", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      handlerFunc,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Error("Server error: %v", err)
		os.Exit(1)
	}
}

// connectDB attempts to connect to the database with retries
func connectDB(cfg *config.Config, log *logger.Logger) (*sql.DB, error) {
	dsn := cfg.GetDSN()
	var db *sql.DB
	var err error

	// Retry logic
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			log.Error("Attempt %d: Failed to open database connection: %v", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		}

		// Test the connection
		err = db.Ping()
		if err != nil {
			log.Error("Attempt %d: Failed to ping database: %v", i+1, err)
			db.Close()
			time.Sleep(2 * time.Second)
			continue
		}

		// Connection successful
		log.Info("Database connection established")
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)
		return db, nil
	}

	return nil, fmt.Errorf("failed to connect to database after %d retries: %w", maxRetries, err)
}
