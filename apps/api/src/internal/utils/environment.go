package utils

import (
	"log/slog"
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
			slog.Info("Environment variables loaded", slog.String("file", envFile))
			slog.Debug("ENV", slog.String("env", os.Getenv("ENV")))
			return
		}
	}

	// If no .env file found, log a warning but don't fail
	// This allows the application to run with environment variables set via other means
	slog.Warn("No .env file found. Make sure environment variables are set via other means (e.g., system environment, Docker, etc.)")
	slog.Debug("ENV", slog.String("env", os.Getenv("ENV")))
}
