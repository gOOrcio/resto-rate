package main

import (
	googlemapsv1connect "api/src/generated/google_maps/v1/v1connect"
	restaurantsv1connect "api/src/generated/restaurants/v1/v1connect"
	usersv1connect "api/src/generated/users/v1/v1connect"
	cache "api/src/internal/cache"
	"api/src/internal/database"
	environment "api/src/internal/utils"
	"api/src/services"
	"api/src/services/google_places"
	"time"

	"github.com/valkey-io/valkey-go"

	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"connectrpc.com/grpcreflect"
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
	environment.MustLoadEnvironmentVariables()
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
		log.Panicf("Failed to connect to database: %v", err)
	}

	log.Println("Database connected")
	return db
}

func mustConnectCache() valkey.Client {
	uri := os.Getenv("VALKEY_URI")
	username := "default" // for dev only!
	password := os.Getenv("VALKEY_PASSWORD")
	client, err := cache.NewValkey(uri, username, password)
	if err != nil {
		log.Panicf("Failed to connect to cache: %v", err)
	}
	log.Println("Cache connected")
	return client
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
		func() ServiceRegistration {
			gapic, err := google_places.NewGooglePlacesAPIClient()
			if err != nil {
				log.Fatalf("places client: %v", err)
			}
			base := google_places.NewDirectPlacesClient(gapic)
			kv := mustConnectCache()
			ttl := 30 * time.Minute
			cached := google_places.NewCachedPlaces(base, kv, ttl, log.Default())

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
			googlemapsv1connect.GoogleMapsServiceName,
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
