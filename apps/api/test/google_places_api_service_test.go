package test

import (
	service "api/src/services"
	environment "api/src/utils"
	"context"
	"log"
	"os"
	"testing"

	api "googlemaps.github.io/maps"
)

func TestGooglePlacesApi(t *testing.T) {
	ctx := context.Background()
	environment.MustLoadEnvironmentVariables()

	log.Print(os.Getenv("GOOGLE_PLACES_API_KEY"))

	places, err := service.NewGooglePlacesAPIService()
	if err != nil {
		log.Fatal("Failed to create client: ", err)
	}

	fields := []api.PlaceSearchFieldMask{
		api.PlaceSearchFieldMaskBusinessStatus,
		api.PlaceSearchFieldMaskPlaceID,
		api.PlaceSearchFieldMaskFormattedAddress,
	}

	textInput := api.FindPlaceFromTextRequest{
		Input:     "Banjaluka",
		InputType: api.FindPlaceFromTextInputTypeTextQuery,
		Fields:    fields,
	}

	response, err := places.Client.FindPlaceFromText(ctx, &textInput)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	if len(response.Candidates) == 0 {
		t.Fatalf("Expected at least one candidate, got 0")
	}

	expectedAddress := "Szkolna 2/4, 00-006 Warszawa, Polska"
	actualAddress := response.Candidates[0].FormattedAddress

	if actualAddress != expectedAddress {
		t.Errorf("Addresses do not match. Expected: %s, Got: %s", expectedAddress, actualAddress)
	}

	log.Printf("Test passed. Found address: %s", actualAddress)
}
