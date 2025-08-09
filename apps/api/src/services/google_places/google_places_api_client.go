package google_places

import (
	v1 "api/src/generated/google_maps/v1"
	"api/src/internal/mappers"
	"api/src/internal/ports"
	"context"
	"fmt"

	"log/slog"

	places "cloud.google.com/go/maps/places/apiv1"
	"cloud.google.com/go/maps/places/apiv1/placespb"
	"google.golang.org/grpc/metadata"
)

type DirectPlacesClient struct {
	gapic *places.Client
}

func NewGooglePlacesAPIClient() (*places.Client, error) {
	return places.NewClient(context.Background())
}

func NewDirectPlacesClient(gapic *places.Client) *DirectPlacesClient {
	return &DirectPlacesClient{gapic: gapic}
}

var _ ports.PlacesClient = (*DirectPlacesClient)(nil)

func (s *DirectPlacesClient) GetPlace(ctx context.Context, req *v1.GetPlaceRequest) (*v1.Place, error) {
	if req.Name == "" {
		slog.Debug("GetPlace: name is required")
		return nil, fmt.Errorf("name is required")
	}

	pbReq := &placespb.GetPlaceRequest{
		Name:         req.Name,
		LanguageCode: req.LanguageCode,
		RegionCode:   req.RegionCode,
		SessionToken: req.SessionToken,
	}

	mask := mappers.BuildFieldMask(req.RequestedFields)
	ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{"X-Goog-FieldMask": mask}))

	resp, err := s.gapic.GetPlace(ctx, pbReq)
	if err != nil {
		slog.Debug("GetPlace: get place failed", slog.Any("error", err))
		return nil, fmt.Errorf("get place failed: %w", err)
	}
	return mappers.PlaceToProto(resp), nil
}

func (s *DirectPlacesClient) GetRestaurantDetails(ctx context.Context, req *v1.GetRestaurantDetailsRequest) (*v1.Place, error) {
	if req.Name == "" {
		slog.Debug("GetRestaurantDetails: name is required")
		return nil, fmt.Errorf("name is required")
	}

	pbReq := &placespb.GetPlaceRequest{
		Name:         req.Name,
		LanguageCode: req.LanguageCode,
		RegionCode:   req.RegionCode,
		SessionToken: req.SessionToken,
	}

	mask := mappers.BuildFieldMask(predefinedRestaurantDetails())
	ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{"X-Goog-FieldMask": mask}))

	resp, err := s.gapic.GetPlace(ctx, pbReq)
	if err != nil {
		slog.Debug("GetRestaurantDetails: get place failed", slog.Any("error", err))
		return nil, fmt.Errorf("get place failed: %w", err)
	}
	return mappers.PlaceToProto(resp), nil
}

func (s *DirectPlacesClient) SearchRestaurants(ctx context.Context, req *v1.SearchRestaurantsRequest) (*v1.SearchTextResponse, error) {
	if req.TextQuery == "" {
		return nil, fmt.Errorf("text_query is required")
	}

	pbReq := &placespb.SearchTextRequest{
		TextQuery:                        req.TextQuery,
		LanguageCode:                     req.LanguageCode,
		RegionCode:                       req.RegionCode,
		RankPreference:                   placespb.SearchTextRequest_RELEVANCE,
		IncludedType:                     "restaurant",
		StrictTypeFiltering:              true,
		IncludePureServiceAreaBusinesses: false,
	}
	// fixed field mask (matches your service)
	ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		"X-Goog-FieldMask": mappers.BuildFieldMask(predefinedRestaurantDetails()),
	}))

	resp, err := s.gapic.SearchText(ctx, pbReq)
	if err != nil {
		return nil, fmt.Errorf("search restaurants failed: %w", err)
	}
	return mappers.SearchTextResponseToProto(resp), nil
}

func (s *DirectPlacesClient) SearchText(ctx context.Context, req *v1.SearchTextRequest) (*v1.SearchTextResponse, error) {
	if req.TextQuery == "" {
		return nil, fmt.Errorf("text_query is required")
	}

	pbReq := &placespb.SearchTextRequest{
		TextQuery:                        req.TextQuery,
		LanguageCode:                     req.LanguageCode,
		RegionCode:                       req.RegionCode,
		RankPreference:                   placespb.SearchTextRequest_RankPreference(req.RankPreference),
		IncludedType:                     req.IncludedType,
		OpenNow:                          req.OpenNow,
		MinRating:                        req.MinRating,
		MaxResultCount:                   req.MaxResultCount,
		PriceLevels:                      mappers.PriceLevelsToSDK(req.PriceLevels),
		StrictTypeFiltering:              req.StrictTypeFiltering,
		IncludePureServiceAreaBusinesses: req.IncludePureServiceAreaBusinesses,
	}

	ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		"X-Goog-FieldMask": mappers.BuildFieldMask(req.RequestedFields),
	}))

	resp, err := s.gapic.SearchText(ctx, pbReq)
	if err != nil {
		return nil, fmt.Errorf("search text failed: %w", err)
	}
	return mappers.SearchTextResponseToProto(resp), nil
}

func predefinedRestaurantDetails() []string {
	return []string{
		"id",
		"name",
		"display_name",
		"formatted_address",
		"short_formatted_address",
		"rating",
		"business_status",
		"google_maps_uri",
		"website_uri",
		"price_level",
		"user_rating_count",
		"current_opening_hours",
		"dine_in",
		"curbside_pickup",
		"reservable",
		"serves_breakfast",
		"serves_lunch",
		"serves_dinner",
		"serves_beer",
		"serves_wine",
		"serves_brunch",
		"serves_vegetarian_food",
		"outdoor_seating",
		"live_music",
		"menu_for_children",
		"serves_cocktails",
		"serves_dessert",
		"serves_coffee",
		"good_for_children",
		"allows_dogs",
		"restroom",
		"good_for_groups",
		"good_for_watching_sports",
		"takeout",
		"generative_summary",
		"review_summary",
		"editorial_summary",
	}
}
