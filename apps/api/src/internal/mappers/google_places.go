package mappers

import (
	v1 "api/src/generated/google_maps/v1"

	placespb "cloud.google.com/go/maps/places/apiv1/placespb"
)

// MapPriceLevels converts v1.PriceLevel slice to placespb.PriceLevel slice
func MapPriceLevels(levels []v1.PriceLevel) []placespb.PriceLevel {
	result := make([]placespb.PriceLevel, len(levels))
	for i, level := range levels {
		result[i] = placespb.PriceLevel(level)
	}
	return result
}

// MapSearchResponseToProto converts placespb.SearchTextResponse to v1.SearchTextResponse
func MapSearchResponseToProto(resp *placespb.SearchTextResponse) *v1.SearchTextResponse {
	if resp == nil {
		return nil
	}

	places := make([]*v1.Place, 0, len(resp.Places))
	for _, place := range resp.Places {
		convertedPlace := MapPlaceToProto(place)
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

// MapPlaceToProto converts placespb.Place to v1.Place
func MapPlaceToProto(place *placespb.Place) *v1.Place {
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

	if place.DisplayName != nil {
		result.DisplayName = &v1.LocalizedText{
			Text:         place.DisplayName.Text,
			LanguageCode: place.DisplayName.LanguageCode,
		}
	}

	if place.PrimaryTypeDisplayName != nil {
		result.PrimaryTypeDisplayName = &v1.LocalizedText{
			Text:         place.PrimaryTypeDisplayName.Text,
			LanguageCode: place.PrimaryTypeDisplayName.LanguageCode,
		}
	}

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

	attributions := make([]*v1.Attribution, 0, len(place.Attributions))
	for _, attr := range place.Attributions {
		attributions = append(attributions, &v1.Attribution{
			Provider:    attr.Provider,
			ProviderUri: attr.ProviderUri,
		})
	}
	result.Attributions = attributions

	if place.UtcOffsetMinutes != nil {
		result.UtcOffsetMinutes = *place.UtcOffsetMinutes
	}

	return result
} 