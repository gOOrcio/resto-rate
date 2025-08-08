package services

import (
	v1 "api/src/generated/google_maps/v1"
	"api/src/generated/google_maps/v1/v1connect"
	"api/src/internal/mappers"
	"context"
	"fmt"
	"strings"

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

	var fieldMask string
	if len(req.Msg.RequestedFields) > 0 {
		fields := make([]string, len(req.Msg.RequestedFields))
		for i, field := range req.Msg.RequestedFields {
			fields[i] = "places." + field
		}
		fieldMask = strings.Join(fields, ",")
	} else {
		fieldMask = "*"
	}

	md := metadata.New(map[string]string{
		"X-Goog-FieldMask": fieldMask,
	})

	ctxWithMetadata := metadata.NewOutgoingContext(ctx, md)

	resp, err := s.client.SearchText(ctxWithMetadata, searchReq) 	
	if err != nil {
		return nil, fmt.Errorf("search text failed: %v", err)
	}

	protoResp := mappers.MapSearchResponseToProto(resp)
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
		PriceLevels:  mappers.MapPriceLevels(req.Msg.PriceLevels),
		StrictTypeFiltering: req.Msg.StrictTypeFiltering,
		IncludePureServiceAreaBusinesses: req.Msg.IncludePureServiceAreaBusinesses,
	}

	var fieldMask string
	if len(req.Msg.RequestedFields) > 0 {
		fields := make([]string, len(req.Msg.RequestedFields))
		for i, field := range req.Msg.RequestedFields {
			fields[i] = "places." + field
		}
		fieldMask = strings.Join(fields, ",")
	} else {
		fieldMask = "*"
	}

	md := metadata.New(map[string]string{
		"X-Goog-FieldMask": fieldMask,
	})
	ctxWithMetadata := metadata.NewOutgoingContext(ctx, md)

	resp, err := s.client.SearchText(ctxWithMetadata, searchReq)
	if err != nil {
		return nil, fmt.Errorf("search text failed: %v", err)
	}

	protoResp := mappers.MapSearchResponseToProto(resp)
	return connect.NewResponse(protoResp), nil
}


