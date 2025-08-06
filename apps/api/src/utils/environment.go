package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func MustLoadEnvironmentVariables() {
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