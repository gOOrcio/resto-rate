package google_places

import (
	v1 "api/src/generated/google_maps/v1"
	"api/src/generated/google_maps/v1/v1connect"
	"api/src/internal/ports"
	"context"

	"connectrpc.com/connect"
)

type GooglePlacesAPIService struct {
	v1connect.UnimplementedGoogleMapsServiceHandler
	client ports.PlacesClient
}

func NewGooglePlacesAPIService(client ports.PlacesClient) *GooglePlacesAPIService {
	return &GooglePlacesAPIService{
		client: client,
	}
}
func (s *GooglePlacesAPIService) GetPlace(ctx context.Context, req *connect.Request[v1.GetPlaceRequest]) (*connect.Response[v1.Place], error) {
	out, err := s.client.GetPlace(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(out), nil
}

func (s *GooglePlacesAPIService) GetRestaurantDetails(ctx context.Context, req *connect.Request[v1.GetRestaurantDetailsRequest]) (*connect.Response[v1.Place], error) {
	out, err := s.client.GetRestaurantDetails(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(out), nil
}

func (s *GooglePlacesAPIService) SearchText(ctx context.Context, req *connect.Request[v1.SearchTextRequest]) (*connect.Response[v1.SearchTextResponse], error) {
	out, err := s.client.SearchText(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(out), nil
}

func (s *GooglePlacesAPIService) SearchRestaurants(ctx context.Context, req *connect.Request[v1.SearchRestaurantsRequest]) (*connect.Response[v1.SearchTextResponse], error) {
	out, err := s.client.SearchRestaurants(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(out), nil
}
