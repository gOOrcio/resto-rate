package services

import (
	v1 "api/src/generated/google_maps/v1"
	"api/src/generated/google_maps/v1/v1connect"
	"api/src/internal/mappers"
	"context"
	"fmt"
	"log/slog"
	"os"

	places "cloud.google.com/go/maps/places/apiv1"
	placespb "cloud.google.com/go/maps/places/apiv1/placespb"
	"connectrpc.com/connect"
	"google.golang.org/api/option"
	"google.golang.org/grpc/metadata"
)

type GooglePlacesAPIService struct {
	v1connect.UnimplementedGoogleMapsServiceHandler
	client *places.Client
}

func NewGooglePlacesAPIClient() (*places.Client, error) {
	return places.NewClient(context.Background(), option.WithAPIKey(os.Getenv("GOOGLE_PLACES_API_KEY")))
}

func NewGooglePlacesAPIService(client *places.Client) *GooglePlacesAPIService {
	return &GooglePlacesAPIService{client: client}
}

func (s *GooglePlacesAPIService) AutocompletePlaces(
	ctx context.Context,
	req *connect.Request[v1.AutocompletePlacesRequest],
) (*connect.Response[v1.AutocompletePlacesResponse], error) {
	if req.Msg.Input == "" {
		return nil, fmt.Errorf("input is required")
	}

	pbReq := &placespb.AutocompletePlacesRequest{
		Input:                   req.Msg.Input,
		IncludedPrimaryTypes:    []string{"restaurant"},
		IncludedRegionCodes:     req.Msg.IncludedRegionCodes,
		LanguageCode:            req.Msg.LanguageCode,
		IncludeQueryPredictions: true,
		SessionToken:            req.Msg.SessionToken,
	}

	out, err := s.client.AutocompletePlaces(ctx, pbReq)
	if err != nil {
		slog.Debug("AutocompletePlaces failed", slog.Any("error", err))
		return nil, fmt.Errorf("autocomplete places failed: %w", err)
	}

	return connect.NewResponse(mappers.AutocompletePlacesResponseToProto(out)), nil
}

func (s *GooglePlacesAPIService) GetPlace(
	ctx context.Context,
	req *connect.Request[v1.GetPlaceRequest],
) (*connect.Response[v1.Place], error) {
	if req.Msg.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	pbReq := &placespb.GetPlaceRequest{
		Name:         req.Msg.Name,
		LanguageCode: req.Msg.LanguageCode,
		RegionCode:   req.Msg.RegionCode,
	}

	if len(req.Msg.RequestedFields) > 0 {
		mask := mappers.BuildFieldMask(req.Msg.RequestedFields)
		ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{"X-Goog-FieldMask": mask}))
	}

	out, err := s.client.GetPlace(ctx, pbReq)
	if err != nil {
		slog.Debug("GetPlace failed", slog.Any("error", err))
		return nil, fmt.Errorf("get place failed: %w", err)
	}
	return connect.NewResponse(mappers.PlaceToProto(out)), nil
}

func (s *GooglePlacesAPIService) GetRestaurantDetails(
	ctx context.Context,
	req *connect.Request[v1.GetRestaurantDetailsRequest],
) (*connect.Response[v1.Place], error) {
	if req.Msg.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	pbReq := &placespb.GetPlaceRequest{
		Name:         req.Msg.Name,
		LanguageCode: req.Msg.LanguageCode,
		RegionCode:   req.Msg.RegionCode,
	}

	mask := mappers.BuildFieldMask(predefinedRestaurantDetails())
	ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{"X-Goog-FieldMask": mask}))

	out, err := s.client.GetPlace(ctx, pbReq)
	if err != nil {
		slog.Debug("GetRestaurantDetails failed", slog.Any("error", err))
		return nil, fmt.Errorf("get place failed: %w", err)
	}
	return connect.NewResponse(mappers.PlaceToProto(out)), nil
}

func (s *GooglePlacesAPIService) SearchText(
	ctx context.Context,
	req *connect.Request[v1.SearchTextRequest],
) (*connect.Response[v1.SearchTextResponse], error) {
	if req.Msg.TextQuery == "" {
		return nil, fmt.Errorf("text_query is required")
	}

	pbReq := &placespb.SearchTextRequest{
		TextQuery:      req.Msg.TextQuery,
		LanguageCode:   req.Msg.LanguageCode,
		RegionCode:     req.Msg.RegionCode,
		IncludedType:   req.Msg.IncludedType,
		OpenNow:        req.Msg.OpenNow,
		MaxResultCount: req.Msg.MaxResultCount,
	}

	if len(req.Msg.RequestedFields) > 0 {
		mask := mappers.BuildFieldMask(req.Msg.RequestedFields)
		ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{"X-Goog-FieldMask": mask}))
	}

	out, err := s.client.SearchText(ctx, pbReq)
	if err != nil {
		slog.Debug("SearchText failed", slog.Any("error", err))
		return nil, fmt.Errorf("search failed: %w", err)
	}
	return connect.NewResponse(mappers.SearchTextResponseToProto(out)), nil
}

func (s *GooglePlacesAPIService) SearchRestaurants(
	ctx context.Context,
	req *connect.Request[v1.SearchRestaurantsRequest],
) (*connect.Response[v1.SearchTextResponse], error) {
	if req.Msg.TextQuery == "" {
		return nil, fmt.Errorf("text_query is required")
	}

	pbReq := &placespb.SearchTextRequest{
		TextQuery:      req.Msg.TextQuery,
		LanguageCode:   req.Msg.LanguageCode,
		RegionCode:     req.Msg.RegionCode,
		IncludedType:   "restaurant",
		MinRating:      4.0,
		MaxResultCount: 20,
	}

	ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		"X-Goog-FieldMask": mappers.BuildFieldMask(predefinedRestaurantDetails()),
	}))

	out, err := s.client.SearchText(ctx, pbReq)
	if err != nil {
		slog.Debug("SearchRestaurants failed", slog.Any("error", err))
		return nil, fmt.Errorf("search failed: %w", err)
	}
	return connect.NewResponse(mappers.SearchTextResponseToProto(out)), nil
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
