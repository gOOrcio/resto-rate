package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go-app/src/routers"
	"go-app/src/services"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
)

func setupRouter(userService *services.UserService, restaurantsService *services.RestaurantsService) *gin.Engine {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	users := r.Group("/users")
	routers.RegisterUserRoutes(users, userService)

	restaurants := r.Group("/restaurants")
	routers.RegisterRestaurantsRoutes(restaurants, restaurantsService)

	return r
}

func loadEnv() error {
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = "./apps/go-app/.env"
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
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}

func main() {
	if err := loadEnv(); err != nil {
		log.Fatal(err)
	}
	db, err := connectToDatabase()
	if err != nil {
		log.Fatal(err)
	}
	userService := &services.UserService{DB: db}
	restaurantsService := &services.RestaurantsService{DB: db}
	r := setupRouter(userService, restaurantsService)
	if err := r.Run(":" + os.Getenv("API_PORT")); err != nil {
		log.Fatal("Failed to run application: ", err)
	}
}
