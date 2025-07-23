package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"go-app/src/database"
	restaurantpb "go-app/src/generated/resto-rate/generated/go/restaurants/v1"
	userpb "go-app/src/generated/resto-rate/generated/go/users/v1"
	"go-app/src/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func loadEnv() error {
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

func getDSN() string {
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
	dsn := getDSN()
	log.Println("Connecting to database ...")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	if db != nil {
		log.Println("Database connected")
	}
	return db, nil
}

func main() {
	log.Println("Application starting... ")
	if err := loadEnv(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Environment variables loaded")
	}

	db, err := connectToDatabase()
	if err != nil {
		log.Fatal(err)
	}

	if strings.EqualFold(os.Getenv("SEED"), "true") {
		log.Println("Seeding... ")
		if err := database.AutoMigrateAndSeed(db); err != nil {
			log.Fatal("Failed to auto-migrate and seed database: ", err)
		}
	}

	userService := &services.UserService{DB: db}
	restaurantsService := &services.RestaurantsService{DB: db}

	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		log.Fatal("API_PORT is not set in the environment variables")
	}

	listener, err := net.Listen("tcp", ":"+apiPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", apiPort, err)
	}

	grpcServer := grpc.NewServer()
	userpb.RegisterUsersServiceServer(grpcServer, userService)
	restaurantpb.RegisterRestaurantServiceServer(grpcServer, restaurantsService)

	reflection.Register(grpcServer)

	log.Println("Application started on port " + apiPort)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal("Failed to run application: ", err)
	}
}
