package main

import (
	"api/src/database"
	restaurantsv1connect "api/src/generated/restaurants/v1/v1connect"
	usersv1connect "api/src/generated/users/v1/v1connect"
	"api/src/services"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"connectrpc.com/grpcreflect"
	"github.com/joho/godotenv"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ServiceRegistration struct {
	Path    string
	Handler http.Handler
}

func main() {
	log.Println("Application starting...")
	mustLoadEnvironmentVariables()
	db := mustConnectToDatabase()
	mux := setupHTTPHandlers(initializeServiceHandlers(db))
	setupDatabaseSchema(db)
	optionallySeedDatabase(db)
	optionallySetupGRPCReflection(mux)
	startServer(mux, getAPIPort())
}

func startServer(mux *http.ServeMux, apiPort string) {
	log.Printf("Starting HTTP server on port %s", apiPort)
	if err := http.ListenAndServe(
		":"+apiPort,
		h2c.NewHandler(mux, &http2.Server{}),
	); err != nil {
		log.Fatal("Failed to run application: ", err)
	}
}

func mustLoadEnvironmentVariables() {
	// Try multiple locations for .env file in order of preference
	envLocations := []string{
		os.Getenv("ENV_FILE"), // Explicitly set ENV_FILE
		".env",                // Current working directory
		"../.env",             // Parent directory
		"../../.env",          // Two levels up
	}

	for _, envFile := range envLocations {
		if envFile == "" {
			continue
		}

		if err := godotenv.Load(envFile); err == nil {
			log.Printf("Environment variables loaded from: %s", envFile)
			log.Println("ENV:", os.Getenv("ENV"))
			return
		}
	}

	// If no .env file found, log a warning but don't fail
	// This allows the application to run with environment variables set via other means
	log.Println("Warning: No .env file found. Make sure environment variables are set via other means (e.g., system environment, Docker, etc.)")
	log.Println("ENV:", os.Getenv("ENV"))
}

func getDatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PORT"),
	)
}

func mustConnectToDatabase() *gorm.DB {
	dsn := getDatabaseDSN()
	log.Println("Connecting to database...")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connected")
	return db
}

func initializeServiceHandlers(db *gorm.DB) []ServiceRegistration {
	return []ServiceRegistration{
		func() ServiceRegistration {
			svc := &services.UserService{DB: db}
			path, handler := usersv1connect.NewUsersServiceHandler(svc)
			return ServiceRegistration{Path: path, Handler: handler}
		}(),
		func() ServiceRegistration {
			svc := &services.RestaurantsService{DB: db}
			path, handler := restaurantsv1connect.NewRestaurantsServiceHandler(svc)
			return ServiceRegistration{Path: path, Handler: handler}
		}(),
	}
}

func setupHTTPHandlers(registrations []ServiceRegistration) *http.ServeMux {
	mux := http.NewServeMux()

	for _, reg := range registrations {
		mux.Handle(reg.Path, corsMiddleware(reg.Handler))
		log.Printf("Service available at: %s", reg.Path)
	}

	return mux
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Connect-Protocol-Version")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func optionallySetupGRPCReflection(mux *http.ServeMux) {
	if os.Getenv("ENV") == "dev" {
		log.Println("!!! gRPC reflection support enabled. We are not in production, right?")

		reflector := grpcreflect.NewStaticReflector(
			usersv1connect.UsersServiceName,
			restaurantsv1connect.RestaurantsServiceName,
		)
		mux.Handle(grpcreflect.NewHandlerV1(reflector))
		mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
	}
}

func setupDatabaseSchema(db *gorm.DB) {
	if err := database.CreateSchema(db); err != nil {
		log.Fatal("Failed to create database schema: ", err)
	}
}

func optionallySeedDatabase(db *gorm.DB) {
	if os.Getenv("ENV") == "dev" && strings.EqualFold(os.Getenv("SEED"), "true") {
		log.Println("Development environment detected with SEED=true, seeding database...")
		if err := database.SeedDatabase(db); err != nil {
			log.Fatal("Failed to seed database: ", err)
		}
	}
}

func getAPIPort() string {
	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		log.Fatal("API_PORT is not set in the environment variables")
	}
	return apiPort
}
