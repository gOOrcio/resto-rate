package services

import (
	v1 "api/src/generated/google_maps/v1"
	"api/src/generated/google_maps/v1/v1connect"
	"connectrpc.com/connect"
	"context"
	"fmt"
	"os"

	sdk "googlemaps.github.io/maps"
)

type GooglePlacesAPIService struct {
	v1connect.UnimplementedGoogleMapsServiceHandler
	Client *sdk.Client
}

func NewGooglePlacesAPIService() (*GooglePlacesAPIService, error) {
	client, err := sdk.NewClient(sdk.WithAPIKey(os.Getenv("GOOGLE_PLACES_API_KEY")))
	if err != nil {
		return nil, err
	}
	return &GooglePlacesAPIService{Client: client}, nil
}

func (s *GooglePlacesAPIService) FindPlaceFromText(ctx context.Context, req *connect.Request[v1.FindPlaceFromTextRequest]) (
	*connect.Response[v1.FindPlaceFromTextResponse], error) {
	if req.Msg.Input == "" {
		return nil, fmt.Errorf("input is required")
	}

	if req.Msg.Fields == nil {
		return nil, fmt.Errorf("fields to be returned are required")
	}

	r := sdk.FindPlaceFromTextRequest{Input: req.Msg.Input}

	_, err := s.Client.FindPlaceFromText(ctx, &r)
	if err != nil {
		return nil, err
	}

	protoResp := &connect.Response[v1.FindPlaceFromTextResponse]{}

	return protoResp, nil
}
