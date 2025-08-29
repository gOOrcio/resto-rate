package main

import (
	googlemapsv1connect "api/src/generated/google_maps/v1/v1connect"
	restaurantsv1connect "api/src/generated/restaurants/v1/v1connect"
	usersv1connect "api/src/generated/users/v1/v1connect"
	"api/src/internal/cache"
	"api/src/internal/utils"
	"api/src/services"
	"log"
	"log/slog"
	"strconv"
	"strings"

	"fmt"
	"net/http"
	"os"

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	utils.MustLoadEnvironmentVariables()
	level := envLogLevel()
	if err := utils.SetupLogging(level); err != nil {
		log.Fatalf("Failed to setup logging: %v", err)
	}
	slog.Info("Application starting...")

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

func connectPrometheusInterceptor() connect.Interceptor {
	reg := prometheus.DefaultRegisterer
	rpcm := utils.NewRPCMetrics(reg)
	return rpcm.ConnectInterceptor()
}

func startServer(mux *http.ServeMux, apiPort string) {
	if os.Getenv("ENV") == "dev" {
		slog.Info("Starting HTTP server for development", slog.String("port", apiPort))
		if err := http.ListenAndServe(
			":"+apiPort,
			h2c.NewHandler(mux, &http2.Server{}),
		); err != nil {
			slog.Error("Failed to run application", slog.Any("error", err))
			os.Exit(1)
		}
	} else {
		slog.Info("Starting HTTPS server", slog.String("port", apiPort))
		if err := http.ListenAndServeTLS(
			":"+apiPort,
			"cert.pem",
			"key.pem",
			h2c.NewHandler(mux, &http2.Server{}),
		); err != nil {
			slog.Error("Failed to run application", slog.Any("error", err))
			os.Exit(1)
		}
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
	slog.Info("Connecting to cache...")
	uri := os.Getenv("VALKEY_URI")
	username := "default"
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
	prometheusInterceptor := connectPrometheusInterceptor()

	return []ServiceRegistration{
		func() ServiceRegistration {
			svc := services.NewUserService(db)
			path, handler := usersv1connect.NewUsersServiceHandler(svc, connect.WithInterceptors(prometheusInterceptor))
			return ServiceRegistration{Path: path, Handler: handler}
		}(),
		func() ServiceRegistration {
			svc := services.NewRestaurantsService(db)
			path, handler := restaurantsv1connect.NewRestaurantsServiceHandler(svc, connect.WithInterceptors(prometheusInterceptor))
			return ServiceRegistration{Path: path, Handler: handler}
		}(),
		func() ServiceRegistration {
			gapic, err := services.NewGooglePlacesAPIClient()
			if err != nil {
				slog.Error("Failed to create Google Places API client", slog.Any("error", err))
				os.Exit(1)
			}
			svc := services.NewGooglePlacesAPIService(gapic)
			path, h := googlemapsv1connect.NewGoogleMapsServiceHandler(
				svc,
				connect.WithInterceptors(prometheusInterceptor),
			)
			return ServiceRegistration{Path: path, Handler: h}
		}(),
	}
}

func setupHTTPHandlers(registrations []ServiceRegistration) *http.ServeMux {
	slog.Info("Registering services...")
	metricsPath := "/metrics"

	mux := http.NewServeMux()
	for _, reg := range registrations {
		mux.Handle(reg.Path, corsMiddleware(reg.Handler))
		slog.Info("Service available", slog.String("path", reg.Path))
	}
	mux.Handle(metricsPath, promhttp.Handler())
	slog.Info("Metrics available", slog.String("path", metricsPath))
	return mux
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestOrigin := r.Header.Get("Origin")

		if os.Getenv("ENV") == "dev" {
			// In development, allow both HTTP and HTTPS origins for flexibility
			allowedOrigins := []string{
				"http://localhost:" + getWebUiPort(),
				"https://localhost:" + getWebUiPort(),
				"http://" + getAPIHost() + ":" + getWebUiPort(),
				"https://" + getAPIHost() + ":" + getWebUiPort(),
			}

			for _, allowed := range allowedOrigins {
				if requestOrigin == allowed {
					w.Header().Set("Access-Control-Allow-Origin", requestOrigin)
					break
				}
			}
		} else {
			allowedOrigin := getAPIProtocol() + "://" + getAPIHost() + ":" + getWebUiPort()
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		}

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

func getAPIProtocol() string {
	apiProtocol := os.Getenv("API_PROTOCOL")
	if apiProtocol == "" {
		log.Fatal("API_PROTOCOL is not set in the environment variables")
	}
	return apiProtocol
}

func getAPIHost() string {
	apiHost := os.Getenv("API_HOST")
	if apiHost == "" {
		log.Fatal("API_HOST is not set in the environment variables")
	}
	return apiHost
}

func getWebUiPort() string {
	webUiPort := os.Getenv("WEB_UI_PORT")
	if webUiPort == "" {
		log.Fatal("WEB_UI_PORT is not set in the environment variables")
	}
	return webUiPort
}

func envLogLevel() slog.Level {
	raw := strings.TrimSpace(os.Getenv("LOG_LEVEL"))
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
