package ports

import (
	v1 "api/src/generated/google_maps/v1"
	"context"
)

type PlacesClient interface {
	SearchText(ctx context.Context, req *v1.SearchTextRequest) (*v1.SearchTextResponse, error)
	SearchRestaurants(ctx context.Context, req *v1.SearchRestaurantsRequest) (*v1.SearchTextResponse, error)
	GetPlace(ctx context.Context, req *v1.GetPlaceRequest) (*v1.Place, error)
	GetRestaurantDetails(ctx context.Context, req *v1.GetRestaurantDetailsRequest) (*v1.Place, error)
}
