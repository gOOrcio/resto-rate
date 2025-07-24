package main

import (
	"api/src/database"
	restaurantsv1connect "api/src/generated/restaurants/v1/v1connect"
	usersv1connect "api/src/generated/users/v1/v1connect"
	"api/src/services"
	"connectrpc.com/grpcreflect"
	"fmt"
	"github.com/joho/godotenv"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		_, filename, _, _ := runtime.Caller(0)
		dir := filepath.Dir(filename)
		envFile = filepath.Join(dir, "../.env")
	}

	if err := godotenv.Load(envFile); err != nil {
		log.Fatalf("Failed to load environment from %s: %v", envFile, err)
	}

	log.Println("Environment variables loaded")
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
			path, handler := restaurantsv1connect.NewRestaurantServiceHandler(svc)
			return ServiceRegistration{Path: path, Handler: handler}
		}(),
	}
}

func setupHTTPHandlers(registrations []ServiceRegistration) *http.ServeMux {
	mux := http.NewServeMux()

	for _, reg := range registrations {
		mux.Handle(reg.Path, reg.Handler)
		log.Printf("Service available at: %s", reg.Path)
	}

	return mux
}

func optionallySetupGRPCReflection(mux *http.ServeMux) {
	if os.Getenv("ENV") == "dev" {
		log.Println("!!! gRPC reflection support enabled. We are not in production, right?")

		reflector := grpcreflect.NewStaticReflector(
			usersv1connect.UsersServiceName,
			restaurantsv1connect.RestaurantServiceName,
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
