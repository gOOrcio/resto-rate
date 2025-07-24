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
	"path/filepath"
	"runtime"
	"strings"

	"connectrpc.com/grpcreflect"
	"github.com/joho/godotenv"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func loadEnvironmentVariables() error {
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		_, filename, _, _ := runtime.Caller(0)
		dir := filepath.Dir(filename)
		envFile = filepath.Join(dir, "../.env")
	}
	if err := godotenv.Load(envFile); err != nil {
		return fmt.Errorf("error loading %s file: %w", envFile, err)
	}
	return nil
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

func connectToDatabase() (*gorm.DB, error) {
	dsn := getDatabaseDSN()
	log.Println("Connecting to database...")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Println("Database connected")
	return db, nil
}

func initializeServices(db *gorm.DB) (*services.UserService, *services.RestaurantsService) {
	userService := &services.UserService{DB: db}
	restaurantsService := &services.RestaurantsService{DB: db}
	return userService, restaurantsService
}

func setupHTTPHandlers(userService *services.UserService, restaurantsService *services.RestaurantsService) (*http.ServeMux, string, string) {
	mux := http.NewServeMux()

	// Register user service
	userPath, userHandler := usersv1connect.NewUsersServiceHandler(userService)
	mux.Handle(userPath, userHandler)

	// Register restaurant service
	restaurantPath, restaurantHandler := restaurantsv1connect.NewRestaurantServiceHandler(restaurantsService)
	mux.Handle(restaurantPath, restaurantHandler)

	return mux, userPath, restaurantPath
}

func setupGRPCReflection(mux *http.ServeMux) {
	if os.Getenv("ENV") == "dev" {
		log.Println("-------------------!!!-------------------")
		log.Println("gRPC reflection support enabled. We are not in production, right?")
		log.Println("-------------------!!!-------------------")

		reflector := grpcreflect.NewStaticReflector(
			usersv1connect.UsersServiceName,
			restaurantsv1connect.RestaurantServiceName,
		)
		mux.Handle(grpcreflect.NewHandlerV1(reflector))
		mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
	}
}

func setupDatabase(db *gorm.DB) {
	if err := database.CreateSchema(db); err != nil {
		log.Fatal("Failed to create database schema: ", err)
	}

	// Seed database only in dev environment when SEED=true
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

func startServer(mux *http.ServeMux, apiPort string) {
	log.Printf("Starting HTTP server on port %s", apiPort)
	if err := http.ListenAndServe(
		":"+apiPort,
		h2c.NewHandler(mux, &http2.Server{}),
	); err != nil {
		log.Fatal("Failed to run application: ", err)
	}
}

func main() {
	log.Println("Application starting...")

	_ = loadEnvironmentVariables()
	log.Println("Environment variables loaded")
	log.Println("ENV: " + os.Getenv("ENV"))

	db, err := connectToDatabase()
	if err != nil {
		log.Fatal(err)
	}

	userService, restaurantsService := initializeServices(db)

	mux, userPath, restaurantPath := setupHTTPHandlers(userService, restaurantsService)

	setupGRPCReflection(mux)

	setupDatabase(db)

	apiPort := getAPIPort()

	log.Printf("User service available at: %s", userPath)
	log.Printf("Restaurant service available at: %s", restaurantPath)

	startServer(mux, apiPort)
}
