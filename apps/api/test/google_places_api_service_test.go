package test

import (
	v1 "api/src/generated/google_maps/v1"
	service "api/src/services"
	environment "api/src/utils"
	"context"
	"log"
	"os"
	"strings"
	"testing"

	"connectrpc.com/connect"
	sdk "googlemaps.github.io/maps"
)

// MockGoogleMapsClient is a mock implementation of the Google Maps client
// This is kept for future use when mocking is needed
type MockGoogleMapsClient struct{}

func (m *MockGoogleMapsClient) FindPlaceFromText(ctx context.Context, req *sdk.FindPlaceFromTextRequest) (sdk.FindPlaceFromTextResponse, error) {
	// Return a mock response for "Banjaluka"
	return sdk.FindPlaceFromTextResponse{
		Candidates: []sdk.PlacesSearchResult{
			{
				FormattedAddress: "Banja Luka, Bosnia and Herzegovina",
				PlaceID:          "mock_place_id",
				Name:             "Banja Luka",
			},
		},
		HTMLAttributions: []string{},
	}, nil
}

func TestGooglePlacesApi(t *testing.T) {
	ctx := context.Background()
	environment.MustLoadEnvironmentVariables()

	apiKey := os.Getenv("GOOGLE_PLACES_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: GOOGLE_PLACES_API_KEY is not set")
	}
	log.Printf("Using API Key: %s...", apiKey[:8])

	// Create real Google Places API client
	client, err := service.NewGooglePlacesAPIClient()
	if err != nil {
		t.Fatalf("Failed to create Google Places API client: %v", err)
	}
	placesService := service.NewGooglePlacesAPIService(client)

	protoRequest := &v1.FindPlaceFromTextRequest{
		Input:     "Banja Luka",
		InputType: v1.InputType_INPUT_TYPE_TEXT_QUERY,
		Fields: []v1.PlaceSearchFieldMask{
			v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_FORMATTED_ADDRESS,
			v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_NAME,
			v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_PLACE_ID,
			v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_RATING,
			v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_TYPES,
			v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_BUSINESS_STATUS,
			v1.PlaceSearchFieldMask_PLACE_SEARCH_FIELD_MASK_ICON,
		},
	}

	connectRequest := connect.NewRequest(protoRequest)

	response, err := placesService.FindPlaceFromText(ctx, connectRequest)
	if err != nil {
		t.Fatalf("Service call failed: %v", err)
	}

	if len(response.Msg.Candidates) == 0 {
		t.Fatalf("Expected at least one candidate, got 0")
	}

	actualAddress := response.Msg.Candidates[0].FormattedAddress
	expectedCity := "Banja Luka"

	if !strings.Contains(actualAddress, expectedCity) {
		t.Errorf("Expected address to contain '%s', but got '%s'", expectedCity, actualAddress)
	}

	t.Logf("Test passed. Found address: %s", actualAddress)
}


