// @title RestoRate API
// @version 0.1
// @description This is the RestoRate API documentation.
// @BasePath /api/v1

package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go-app/src/database"
	docs "go-app/src/docs"
	"go-app/src/routers"
	"go-app/src/services"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func setupRouter(userService *services.UserService, restaurantsService *services.RestaurantsService) *gin.Engine {
	r := gin.Default()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	api := r.Group("/api/v1")
	routers.RegisterUserRoutes(api, userService)
	routers.RegisterRestaurantsRoutes(api, restaurantsService)

	return r
}

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

	if strings.EqualFold(os.Getenv("SEED"), "true") {
		if err := database.AutoMigrateAndSeed(db); err != nil {
			log.Fatal("Failed to auto-migrate and seed database: ", err)
		}
	}
	r := setupRouter(userService, restaurantsService)

	apiPort := os.Getenv("API_PORT")
	// docs.SwaggerInfo.Title and Version are now set via annotation comments above.
	log.Printf("Swagger docs available at http://localhost:%s/swagger/index.html", apiPort)
	log.Printf("Swagger base path: %s", docs.SwaggerInfo.BasePath)

	if err := r.Run(":" + apiPort); err != nil {
		log.Fatal("Failed to run application: ", err)
	}
}
