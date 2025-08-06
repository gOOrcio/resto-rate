package services

import (
	v1 "api/src/generated/google_maps/v1"
	"api/src/generated/google_maps/v1/v1connect"
	"api/src/services/mappers"
	"context"
	"fmt"
	"os"

	"log"

	"connectrpc.com/connect"
	sdk "googlemaps.github.io/maps"
)

type GoogleMapsClientInterface interface {
	FindPlaceFromText(ctx context.Context, req *sdk.FindPlaceFromTextRequest) (sdk.FindPlaceFromTextResponse, error)
}

type GooglePlacesAPIService struct {
	v1connect.UnimplementedGoogleMapsServiceHandler
	Client GoogleMapsClientInterface
}

func NewGooglePlacesAPIClient() (*sdk.Client, error) {
	client, err := sdk.NewClient(sdk.WithAPIKey(os.Getenv("GOOGLE_PLACES_API_KEY")))
	if err != nil {
		return nil, err
	}
	return client, nil
}

func NewGooglePlacesAPIService(client GoogleMapsClientInterface) *GooglePlacesAPIService {
	return &GooglePlacesAPIService{
		Client: client,
	}
}

func (s *GooglePlacesAPIService) FindPlaceFromText(ctx context.Context, req *connect.Request[v1.FindPlaceFromTextRequest]) (
	*connect.Response[v1.FindPlaceFromTextResponse], error) {
	log.Println("FindPlaceFromText handler was called")
	if req.Msg.Input == "" {
		return nil, fmt.Errorf("input is required")
	}

	if len(req.Msg.Fields) == 0 {
		return nil, fmt.Errorf("fields to be returned are required")
	}

	request := mappers.MapFindPlaceFromTextRequestToGo(req.Msg)

	response, err := s.Client.FindPlaceFromText(ctx, request)
	if err != nil {
		return nil, err
	}

	protoResp := mappers.MapFindPlaceFromTextResponseToProto(&response)

	return connect.NewResponse(protoResp), nil
}
