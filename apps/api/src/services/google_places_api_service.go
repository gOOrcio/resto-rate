package services

import (
	v1 "api/src/generated/google_maps/v1"
	"api/src/generated/google_maps/v1/v1connect"
	"api/src/internal/mappers"
	"context"
	"fmt"

	places "cloud.google.com/go/maps/places/apiv1"
	placespb "cloud.google.com/go/maps/places/apiv1/placespb"
	"connectrpc.com/connect"
	"google.golang.org/grpc/metadata"
)

type GooglePlacesAPIService struct {
	v1connect.UnimplementedGoogleMapsServiceHandler
	client *places.Client
}

func NewGooglePlacesAPIClient() (*places.Client, error) {
	ctx := context.Background()
	client, err := places.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create places client: %v", err)
	}
	return client, nil
}

func NewGooglePlacesAPIService(client *places.Client) *GooglePlacesAPIService {
	return &GooglePlacesAPIService{
		client: client,
	}
}

func (s *GooglePlacesAPIService) GetPlace(ctx context.Context, req *connect.Request[v1.GetPlaceRequest]) (
	*connect.Response[v1.Place], error) {
		if req.Msg.Name == "" {
			return nil, fmt.Errorf("name is required")
		} 
		
		placeReq := &placespb.GetPlaceRequest{
			Name: req.Msg.Name,
			LanguageCode: req.Msg.LanguageCode,
			RegionCode: req.Msg.RegionCode,
			SessionToken: req.Msg.SessionToken,
		}

		var fieldMask = mappers.BuildFieldMask(req.Msg.RequestedFields)
		md := metadata.New(map[string]string{
			"X-Goog-FieldMask": fieldMask,
		})

		ctxWithMetadata := metadata.NewOutgoingContext(ctx, md)

		resp, err := s.client.GetPlace(ctxWithMetadata, placeReq)
		if err != nil {
			return nil, fmt.Errorf("get place failed: %v", err)
		}
		protoResp := mappers.PlaceToProto(resp)
		return connect.NewResponse(protoResp), nil
	}

func (s *GooglePlacesAPIService) GetRestaurantDetails(ctx context.Context, req *connect.Request[v1.GetRestaurantDetailsRequest]) (
	*connect.Response[v1.Place], error) {
		if req.Msg.Name == "" {
			return nil, fmt.Errorf("name is required")
		} 
		
		placeReq := &placespb.GetPlaceRequest{
			Name: req.Msg.Name,
			LanguageCode: req.Msg.LanguageCode,
			RegionCode: req.Msg.RegionCode,
			SessionToken: req.Msg.SessionToken,
		}

		var fieldMask = mappers.BuildFieldMask(predefinedRestaurantDetails())
		md := metadata.New(map[string]string{
			"X-Goog-FieldMask": fieldMask,
		})

		ctxWithMetadata := metadata.NewOutgoingContext(ctx, md)

		resp, err := s.client.GetPlace(ctxWithMetadata, placeReq)
		if err != nil {
			return nil, fmt.Errorf("get place failed: %v", err)
		}
		protoResp := mappers.PlaceToProto(resp)
		return connect.NewResponse(protoResp), nil
	}

func (s *GooglePlacesAPIService) SearchRestaurants(ctx context.Context, req *connect.Request[v1.SearchRestaurantsRequest]) (
	*connect.Response[v1.SearchTextResponse], error) {
		if req.Msg.TextQuery == "" {
		return nil, fmt.Errorf("text_query is required")
	}

	searchReq := &placespb.SearchTextRequest{
		TextQuery: req.Msg.TextQuery,
		LanguageCode: req.Msg.LanguageCode,
		RegionCode: req.Msg.RegionCode,
		RankPreference: placespb.SearchTextRequest_RELEVANCE,
		IncludedType: "restaurant",
		StrictTypeFiltering: true,
		IncludePureServiceAreaBusinesses: false,
	}

	var fieldMask = mappers.BuildFieldMask(predefinedRestaurantDetails())
	md := metadata.New(map[string]string{
		"X-Goog-FieldMask": fieldMask,
	})

	ctxWithMetadata := metadata.NewOutgoingContext(ctx, md)

	resp, err := s.client.SearchText(ctxWithMetadata, searchReq) 	
	if err != nil {
		return nil, fmt.Errorf("search text failed: %v", err)
	}

	protoResp := mappers.SearchTextResponseToProto(resp)
	return connect.NewResponse(protoResp), nil
}

func (s *GooglePlacesAPIService) SearchText(ctx context.Context, req *connect.Request[v1.SearchTextRequest]) (
	*connect.Response[v1.SearchTextResponse], error) {
	if req.Msg.TextQuery == "" {
		return nil, fmt.Errorf("text_query is required")
	}

	searchReq := &placespb.SearchTextRequest{
		TextQuery:    req.Msg.TextQuery,
		LanguageCode: req.Msg.LanguageCode,
		RegionCode:   req.Msg.RegionCode,
		RankPreference: placespb.SearchTextRequest_RankPreference(req.Msg.RankPreference),
		IncludedType: req.Msg.IncludedType,
		OpenNow:      req.Msg.OpenNow,
		MinRating:    req.Msg.MinRating,
		MaxResultCount: req.Msg.MaxResultCount,
		PriceLevels:  mappers.PriceLevelsToSDK(req.Msg.PriceLevels),
		StrictTypeFiltering: req.Msg.StrictTypeFiltering,
		IncludePureServiceAreaBusinesses: req.Msg.IncludePureServiceAreaBusinesses,
	}

	var fieldMask = mappers.BuildFieldMask(req.Msg.RequestedFields)
	
	md := metadata.New(map[string]string{
		"X-Goog-FieldMask": fieldMask,
	})
	
	ctxWithMetadata := metadata.NewOutgoingContext(ctx, md)

	resp, err := s.client.SearchText(ctxWithMetadata, searchReq)
	if err != nil {
		return nil, fmt.Errorf("search text failed: %v", err)
	}

	protoResp := mappers.SearchTextResponseToProto(resp)
	return connect.NewResponse(protoResp), nil
}

func predefinedRestaurantDetails() []string {
	return []string{
    		"id",
    		"name",
    		"display_name",
    		"formatted_address",
    		"short_formatted_address",
    		"rating",
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
		}
}

