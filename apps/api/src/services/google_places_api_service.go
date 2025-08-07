package services

import (
	v1 "api/src/generated/google_maps/v1"
	"api/src/generated/google_maps/v1/v1connect"
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
		PriceLevels:  convertPriceLevels(req.Msg.PriceLevels),
		StrictTypeFiltering: req.Msg.StrictTypeFiltering,
		IncludePureServiceAreaBusinesses: req.Msg.IncludePureServiceAreaBusinesses,
	}

	// Build FieldMask from requested fields
	var fieldMask string
	if len(req.Msg.RequestedFields) > 0 {
		fields := make([]string, len(req.Msg.RequestedFields))
		for i, field := range req.Msg.RequestedFields {
			fields[i] = "places." + field
		}
		fieldMask = strings.Join(fields, ",")
	} else {
		// If no fields specified, use all available fields
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

	protoResp := convertSearchResponseToProto(resp)
	return connect.NewResponse(protoResp), nil
}

// Convert price levels from our proto to Google's proto
func convertPriceLevels(levels []v1.PriceLevel) []placespb.PriceLevel {
	result := make([]placespb.PriceLevel, len(levels))
	for i, level := range levels {
		result[i] = placespb.PriceLevel(level)
	}
	return result
}

// Convert the Google API response to our protobuf format
func convertSearchResponseToProto(resp *placespb.SearchTextResponse) *v1.SearchTextResponse {
	if resp == nil {
		return nil
	}

	places := make([]*v1.Place, 0, len(resp.Places))
	for _, place := range resp.Places {
		convertedPlace := convertPlaceToProto(place)
		places = append(places, convertedPlace)
	}

	routingSummaries := make([]*v1.RoutingSummary, 0)
	contextualContents := make([]*v1.ContextualContent, 0)

	return &v1.SearchTextResponse{
		Places:            places,
		RoutingSummaries:  routingSummaries,
		ContextualContents: contextualContents,
	}
}

// Convert a place from Google's API to our protobuf format
func convertPlaceToProto(place *placespb.Place) *v1.Place {
	if place == nil {
		return nil
	}

	result := &v1.Place{
		Name:        place.Name,
		Id:          place.Id,
		Types:       place.Types,
		PrimaryType: place.PrimaryType,
		NationalPhoneNumber: place.NationalPhoneNumber,
		InternationalPhoneNumber: place.InternationalPhoneNumber,
		FormattedAddress: place.FormattedAddress,
		ShortFormattedAddress: place.ShortFormattedAddress,
		Rating: place.Rating,
		GoogleMapsUri: place.GoogleMapsUri,
		WebsiteUri: place.WebsiteUri,
		AdrFormatAddress: place.AdrFormatAddress,
		BusinessStatus: v1.BusinessStatus(place.BusinessStatus),
		PriceLevel: v1.PriceLevel(place.PriceLevel),
		IconMaskBaseUri: place.IconMaskBaseUri,
		IconBackgroundColor: place.IconBackgroundColor,
	}

	// Handle pointer types
	if place.UserRatingCount != nil {
		result.UserRatingCount = *place.UserRatingCount
	}
	if place.Takeout != nil {
		result.Takeout = *place.Takeout
	}
	if place.Delivery != nil {
		result.Delivery = *place.Delivery
	}
	if place.DineIn != nil {
		result.DineIn = *place.DineIn
	}
	if place.CurbsidePickup != nil {
		result.CurbsidePickup = *place.CurbsidePickup
	}
	if place.Reservable != nil {
		result.Reservable = *place.Reservable
	}
	if place.ServesBreakfast != nil {
		result.ServesBreakfast = *place.ServesBreakfast
	}
	if place.ServesLunch != nil {
		result.ServesLunch = *place.ServesLunch
	}
	if place.ServesDinner != nil {
		result.ServesDinner = *place.ServesDinner
	}
	if place.ServesBeer != nil {
		result.ServesBeer = *place.ServesBeer
	}
	if place.ServesWine != nil {
		result.ServesWine = *place.ServesWine
	}
	if place.ServesBrunch != nil {
		result.ServesBrunch = *place.ServesBrunch
	}
	if place.ServesVegetarianFood != nil {
		result.ServesVegetarianFood = *place.ServesVegetarianFood
	}
	if place.OutdoorSeating != nil {
		result.OutdoorSeating = *place.OutdoorSeating
	}
	if place.LiveMusic != nil {
		result.LiveMusic = *place.LiveMusic
	}
	if place.MenuForChildren != nil {
		result.MenuForChildren = *place.MenuForChildren
	}
	if place.ServesCocktails != nil {
		result.ServesCocktails = *place.ServesCocktails
	}
	if place.ServesDessert != nil {
		result.ServesDessert = *place.ServesDessert
	}
	if place.ServesCoffee != nil {
		result.ServesCoffee = *place.ServesCoffee
	}
	if place.GoodForChildren != nil {
		result.GoodForChildren = *place.GoodForChildren
	}
	if place.AllowsDogs != nil {
		result.AllowsDogs = *place.AllowsDogs
	}
	if place.Restroom != nil {
		result.Restroom = *place.Restroom
	}
	if place.GoodForGroups != nil {
		result.GoodForGroups = *place.GoodForGroups
	}
	if place.GoodForWatchingSports != nil {
		result.GoodForWatchingSports = *place.GoodForWatchingSports
	}
	if place.PureServiceAreaBusiness != nil {
		result.PureServiceAreaBusiness = *place.PureServiceAreaBusiness
	}

	// Convert display name
	if place.DisplayName != nil {
		result.DisplayName = &v1.LocalizedText{
			Text:         place.DisplayName.Text,
			LanguageCode: place.DisplayName.LanguageCode,
		}
	}

	// Convert primary type display name
	if place.PrimaryTypeDisplayName != nil {
		result.PrimaryTypeDisplayName = &v1.LocalizedText{
			Text:         place.PrimaryTypeDisplayName.Text,
			LanguageCode: place.PrimaryTypeDisplayName.LanguageCode,
		}
	}

	// Convert photos
	photos := make([]*v1.Photo, 0, len(place.Photos))
	for _, photo := range place.Photos {
		authorAttributions := make([]string, 0, len(photo.AuthorAttributions))
		for _, attr := range photo.AuthorAttributions {
			authorAttributions = append(authorAttributions, attr.Uri)
		}
		
		photos = append(photos, &v1.Photo{
			Name:                photo.Name,
			WidthPx:            photo.WidthPx,
			HeightPx:           photo.HeightPx,
			AuthorAttributions: authorAttributions,
		})
	}
	result.Photos = photos

	// Convert attributions
	attributions := make([]*v1.Attribution, 0, len(place.Attributions))
	for _, attr := range place.Attributions {
		attributions = append(attributions, &v1.Attribution{
			Provider:    attr.Provider,
			ProviderUri: attr.ProviderUri,
		})
	}
	result.Attributions = attributions

	// Convert UTC offset
	if place.UtcOffsetMinutes != nil {
		result.UtcOffsetMinutes = *place.UtcOffsetMinutes
	}

	return result
}
