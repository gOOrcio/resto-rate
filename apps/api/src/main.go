package main

import (
	googlemapsv1connect "api/src/generated/google_maps/v1/v1connect"
	restaurantsv1connect "api/src/generated/restaurants/v1/v1connect"
	usersv1connect "api/src/generated/users/v1/v1connect"
	"api/src/internal/cache"
	"api/src/internal/utils"
	"api/src/services"
	"api/src/services/google_places"
	"log"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"fmt"
	"net/http"
	"os"

	"connectrpc.com/grpcreflect"
	"github.com/valkey-io/valkey-go"
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
	slog.Info("Application starting...")
	utils.MustLoadEnvironmentVariables()
	slog.SetLogLoggerLevel(envLogLevel())

	db := mustConnectToDatabase()
	mux := setupHTTPHandlers(initializeServiceHandlers(db))

	err := utils.CreateSchema(db)
	if err != nil {
		slog.Error("Failed to create database schema", slog.Any("error", err))
		os.Exit(1)
	}

	err = utils.SeedDatabase(db)
	if err != nil {
		slog.Error("Failed to seed database", slog.Any("error", err))
		os.Exit(1)
	}

	optionallySetupGRPCReflection(mux)
	startServer(mux, getAPIPort())
}

func startServer(mux *http.ServeMux, apiPort string) {
	slog.Info("Starting HTTP server", slog.String("port", apiPort))
	if err := http.ListenAndServe(
		":"+apiPort,
		h2c.NewHandler(mux, &http2.Server{}),
	); err != nil {
		slog.Error("Failed to run application", slog.Any("error", err))
		os.Exit(1)
	}
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
	slog.Info("Connecting to database...")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		slog.Error("Failed to connect to database", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("Database connected")
	return db
}

func mustConnectCache() valkey.Client {
	uri := os.Getenv("VALKEY_URI")
	username := "default" // for dev only!
	password := os.Getenv("VALKEY_PASSWORD")
	client, err := cache.NewValkey(uri, username, password)
	if err != nil {
		slog.Error("Failed to connect to cache", slog.Any("error", err))
		os.Exit(1)
	}
	slog.Info("Cache connected")
	return client
}

func initializeServiceHandlers(db *gorm.DB) []ServiceRegistration {
	return []ServiceRegistration{
		func() ServiceRegistration {
			svc := services.NewUserService(db)
			path, handler := usersv1connect.NewUsersServiceHandler(svc)
			return ServiceRegistration{Path: path, Handler: handler}
		}(),
		func() ServiceRegistration {
			svc := services.NewRestaurantsService(db)
			path, handler := restaurantsv1connect.NewRestaurantsServiceHandler(svc)
			return ServiceRegistration{Path: path, Handler: handler}
		}(),
		func() ServiceRegistration {
			gapic, err := google_places.NewGooglePlacesAPIClient()
			if err != nil {
				slog.Error("places client", slog.Any("error", err))
				os.Exit(1)
			}
			base := google_places.NewDirectPlacesClient(gapic)
			kv := mustConnectCache()
			ttl := 30 * time.Minute
			cached := google_places.NewCachedPlaces(base, kv, ttl)

			svc := google_places.NewGooglePlacesAPIService(cached)
			path, h := googlemapsv1connect.NewGoogleMapsServiceHandler(svc)
			return ServiceRegistration{Path: path, Handler: h}
		}(),
	}
}

func setupHTTPHandlers(registrations []ServiceRegistration) *http.ServeMux {
	mux := http.NewServeMux()
	for _, reg := range registrations {
		mux.Handle(reg.Path, corsMiddleware(reg.Handler))
		slog.Info("Service available", slog.String("path", reg.Path))
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
		slog.Info("gRPC reflection support enabled. We are not in production, right?")
		reflector := grpcreflect.NewStaticReflector(
			usersv1connect.UsersServiceName,
			restaurantsv1connect.RestaurantsServiceName,
			googlemapsv1connect.GoogleMapsServiceName,
		)
		mux.Handle(grpcreflect.NewHandlerV1(reflector))
		mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
	}
}

func getAPIPort() string {
	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		log.Fatal("API_PORT is not set in the environment variables")
	}
	return apiPort
}

func envLogLevel() slog.Level {
	raw := strings.TrimSpace(os.Getenv("LOG_LEVEL"))
	println(raw)
	if raw == "" {
		slog.Info("LOG_LEVEL defaulting to INFO")
		return slog.LevelInfo
	}

	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(raw)); err == nil {
		slog.Info("LOG_LEVEL", slog.String("level", lvl.String()))
		return lvl
	}

	if n, err := strconv.Atoi(raw); err == nil {
		slog.Info("LOG_LEVEL", slog.String("level", slog.Level(n).String()))
		return slog.Level(n)
	}

	slog.Info("LOG_LEVEL fallback to INFO")
	return slog.LevelInfo
}
