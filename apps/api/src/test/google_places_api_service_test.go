package test

import (
	v1 "api/src/generated/google_maps/v1"
	environment "api/src/internal/utils"
	"api/src/services"

	"context"
	"log"
	"os"
	"strings"
	"testing"

	placespb "cloud.google.com/go/maps/places/apiv1/placespb"
	"connectrpc.com/connect"
	"google.golang.org/genproto/googleapis/type/localized_text"
)

// MockGooglePlacesClient is a mock implementation of the Google Places client
// This is kept for future use when mocking is needed
type MockGooglePlacesClient struct{}

func (m *MockGooglePlacesClient) SearchText(ctx context.Context, req *placespb.SearchTextRequest, opts ...interface{}) (*placespb.SearchTextResponse, error) {
	// Return a mock response for "Banja Luka"
	return &placespb.SearchTextResponse{
		Places: []*placespb.Place{
			{
				Id:               "mock_place_id",
				FormattedAddress: "Banja Luka, Bosnia and Herzegovina",
				DisplayName: &localized_text.LocalizedText{
					Text:         "Banja Luka",
					LanguageCode: "en",
				},
			},
		},
	}, nil
}

func TestGooglePlacesApi(t *testing.T) {
	ctx := context.Background()
	environment.MustLoadEnvironmentVariables()

	// Note: The new Google Cloud Maps API uses Application Default Credentials
	// or service account keys, not API keys like the old Places API
	// You'll need to set up authentication differently
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		t.Skip("Skipping test: GOOGLE_CLOUD_PROJECT is not set")
	}
	log.Printf("Using Project ID: %s", projectID)

	// Create real Google Places API client
	client, err := services.NewGooglePlacesAPIClient()
	if err != nil {
		t.Fatalf("Failed to create Google Places API client: %v", err)
	}
	defer client.Close()

	placesService := services.NewGooglePlacesAPIService(client)

	protoRequest := &v1.SearchTextRequest{
		TextQuery:           "Banja Luka",
		IncludedType:        "restaurant",
		StrictTypeFiltering: true,
		RankPreference:      v1.RankPreference_RANK_PREFERENCE_RELEVANCE,
		MaxResultCount:      20,
	}

	connectRequest := connect.NewRequest(protoRequest)

	response, err := placesService.SearchText(ctx, connectRequest)
	if err != nil {
		t.Fatalf("Service call failed: %v", err)
	}

	if len(response.Msg.Places) == 0 {
		t.Fatalf("Expected at least one place, got 0")
	}

	actualAddress := response.Msg.Places[0].FormattedAddress
	expectedCity := "Banja Luka"

	if !strings.Contains(actualAddress, expectedCity) {
		t.Errorf("Expected address to contain '%s', but got '%s'", expectedCity, actualAddress)
	}

	t.Logf("Test passed. Found address: %s", actualAddress)
}

func TestDynamicFieldMask(t *testing.T) {
	ctx := context.Background()
	environment.MustLoadEnvironmentVariables()

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		t.Skip("Skipping test: GOOGLE_CLOUD_PROJECT is not set")
	}

	client, err := services.NewGooglePlacesAPIClient()
	if err != nil {
		t.Fatalf("Failed to create Google Places API client: %v", err)
	}
	defer client.Close()

	placesService := services.NewGooglePlacesAPIService(client)

	// Test with specific requested fields
	protoRequest := &v1.SearchTextRequest{
		TextQuery:           "Banja Luka",
		IncludedType:        "restaurant",
		StrictTypeFiltering: true,
		RankPreference:      v1.RankPreference_RANK_PREFERENCE_RELEVANCE,
		MaxResultCount:      5,
		RequestedFields:     []string{"name", "displayName", "rating", "formattedAddress"},
	}

	connectRequest := connect.NewRequest(protoRequest)

	response, err := placesService.SearchText(ctx, connectRequest)
	if err != nil {
		t.Fatalf("Service call with dynamic fields failed: %v", err)
	}

	if len(response.Msg.Places) == 0 {
		t.Fatalf("Expected at least one place, got 0")
	}

	place := response.Msg.Places[0]

	// Verify that we got the requested fields
	if place.Name == "" {
		t.Error("Expected name field to be populated")
	}

	if place.DisplayName == nil || place.DisplayName.Text == "" {
		t.Error("Expected displayName field to be populated")
	}

	if place.FormattedAddress == "" {
		t.Error("Expected formattedAddress field to be populated")
	}

	t.Logf("Dynamic FieldMask test passed. Place: %s", place.DisplayName.Text)
}
