package adapters

import (
	gmv1 "api/src/generated/google_maps/v1"
	"api/src/services"
	gm "api/src/services"
	"context"

	"connectrpc.com/connect"
)

type PlacesSvcAdapter struct {
	svc *gm.GooglePlacesAPIService
}

func NewPlacesSvcAdapter(s *services.GooglePlacesAPIService) *PlacesSvcAdapter {
	return &PlacesSvcAdapter{svc: s}
}

func (a *PlacesSvcAdapter) SearchText(ctx context.Context, req *gmv1.SearchTextRequest) (*gmv1.SearchTextResponse, error) {
	resp, err := a.svc.SearchText(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}

func (a *PlacesSvcAdapter) SearchRestaurants(ctx context.Context, req *gmv1.SearchRestaurantsRequest) (*gmv1.SearchTextResponse, error) {
	resp, err := a.svc.SearchRestaurants(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}

func (a *PlacesSvcAdapter) GetPlace(ctx context.Context, req *gmv1.GetPlaceRequest) (*gmv1.Place, error) {
	resp, err := a.svc.GetPlace(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}

func (a *PlacesSvcAdapter) GetRestaurantDetails(ctx context.Context, req *gmv1.GetRestaurantDetailsRequest) (*gmv1.Place, error) {
	resp, err := a.svc.GetRestaurantDetails(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}
